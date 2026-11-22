package runner

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Pekly/NeoPanel/internal/chunker"
	"github.com/Pekly/NeoPanel/internal/rooted"
)

var conversionJavaMu sync.Mutex

// Reuse managed discovery and the existing bounded/checksummed Temurin download
// helper. The legacy Linux InstallJavaRuntime extractor is not bounded/cancellable,
// so auxiliary jobs use this confined extractor and private versioned publication.
func conversionJava(ctx context.Context, toolDir string) (string, error) {
	for _, version := range []string{"21", "17", "25"} {
		path, _ := ResolveJavaBinary(version)
		if filepath.IsAbs(path) {
			if n, _ := strconv.Atoi(detectJavaBinaryMajor(path)); n >= 17 {
				return path, nil
			}
		}
	}
	for !conversionJavaMu.TryLock() {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(100 * time.Millisecond):
		}
	}
	defer conversionJavaMu.Unlock()
	if runtime.GOOS != "linux" {
		return "", errors.New("automatic converter Java provisioning requires Linux")
	}
	rootPath := filepath.Join(toolDir, "java21")
	binaryPath := filepath.Join(rootPath, "bin", "java")
	if detectJavaBinaryMajor(binaryPath) == "21" {
		return binaryPath, nil
	}
	arch := "x64"
	if runtime.GOARCH == "arm64" {
		arch = "aarch64"
	} else if runtime.GOARCH != "amd64" {
		return "", errors.New("unsupported Java provisioning architecture")
	}
	ctx, cancel := context.WithTimeout(ctx, 6*time.Minute)
	defer cancel()
	client := javaHTTPClient(&http.Client{Timeout: 5 * time.Minute}, "21")
	endpoint := "https://api.adoptium.net/v3/assets/latest/21/hotspot?architecture=" + arch + "&image_type=jre&os=linux&vendor=eclipse"
	resp, err := javaGET(ctx, client, endpoint)
	if err != nil {
		return "", err
	}
	body, err := rooted.ReadBounded(resp.Body, javaMetadataLimit)
	resp.Body.Close()
	if err != nil {
		return "", err
	}
	var assets []temurinAsset
	if err = json.Unmarshal(body, &assets); err != nil {
		return "", err
	}
	var pkg temurinPackage
	for _, a := range assets {
		if a.Version.Major == 21 && a.Binary.Architecture == arch && a.Binary.OS == "linux" && a.Binary.ImageType == "jre" && a.Binary.JVMImpl == "hotspot" {
			pkg = a.Binary.Package
			break
		}
	}
	u, e := url.Parse(pkg.Link)
	checksum, e2 := hex.DecodeString(pkg.Checksum)
	if e != nil || !javaHTTPSURL(u) || u.Host != "github.com" || !strings.HasPrefix(u.Path, "/adoptium/temurin21-binaries/releases/download/") || !strings.HasSuffix(u.Path, ".tar.gz") || strings.Contains(u.Path, "/../") || strings.Contains(u.Path, "\\") || e2 != nil || len(checksum) != sha256.Size || pkg.Size <= 0 || pkg.Size > javaDownloadLimit {
		return "", errors.New("invalid official Java package metadata")
	}
	if err = chunker.PrivateDir(toolDir); err != nil {
		return "", err
	}
	stage, err := os.MkdirTemp(toolDir, ".java21-*")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(stage)
	archive := filepath.Join(stage, "runtime.tar.gz")
	// Despite the legacy name this helper verifies an opaque bounded package;
	// archive-type handling is deliberately separate.
	if err = downloadTemurinZIP(ctx, client, pkg, archive); err != nil {
		return "", err
	}
	fresh := filepath.Join(stage, "runtime")
	if err = os.Mkdir(fresh, 0700); err != nil {
		return "", err
	}
	if err = conversionExtractJava(ctx, archive, fresh); err != nil {
		return "", err
	}
	if detectJavaBinaryMajor(filepath.Join(fresh, "bin", "java")) != "21" {
		return "", errors.New("downloaded managed runtime did not report Java 21")
	}
	if _, err = os.Lstat(rootPath); !os.IsNotExist(err) {
		return "", errors.New("existing converter Java runtime is invalid; administrator review required")
	}
	if err = os.Rename(fresh, rootPath); err != nil {
		return "", err
	}
	return binaryPath, nil
}
func conversionExtractJava(ctx context.Context, archive, dest string) error {
	in, err := os.Open(archive)
	if err != nil {
		return err
	}
	defer in.Close()
	gz, err := gzip.NewReader(in)
	if err != nil {
		return err
	}
	defer gz.Close()
	tr := tar.NewReader(io.LimitReader(gz, javaExpandedLimit+(32<<20)))
	root, err := os.OpenRoot(dest)
	if err != nil {
		return err
	}
	defer root.Close()
	var total int64
	count := 0
	links := map[string]string{}
	for {
		if err = ctx.Err(); err != nil {
			return err
		}
		h, e := tr.Next()
		if e == io.EOF {
			break
		}
		if e != nil {
			return e
		}
		count++
		if count > javaEntryLimit {
			return errors.New("Java archive entry limit")
		}
		name := stripFirstPathSegment(h.Name)
		if name == "" || name == "." {
			continue
		}
		if !filepath.IsLocal(name) || strings.ContainsAny(name, "\\:\x00") {
			return errors.New("unsafe Java archive path")
		}
		for _, part := range strings.Split(strings.TrimSuffix(name, "/"), "/") {
			if !rooted.Basename(part) {
				return errors.New("unsafe Java archive component")
			}
		}
		switch h.Typeflag {
		case tar.TypeDir:
			if e = root.MkdirAll(name, 0755); e != nil {
				return e
			}
		case tar.TypeReg:
			if h.Size < 0 || h.Size > javaExpandedLimit-total {
				return errors.New("Java archive expanded byte limit")
			}
			if e = root.MkdirAll(filepath.Dir(name), 0755); e != nil {
				return e
			}
			f, e := root.OpenFile(name, os.O_CREATE|os.O_EXCL|os.O_WRONLY, os.FileMode(h.Mode)&0755)
			if e != nil {
				return e
			}
			n, e := rooted.Copy(ctx, f, tr, h.Size)
			e = errors.Join(e, f.Close())
			if e != nil {
				return e
			}
			if n != h.Size {
				return io.ErrUnexpectedEOF
			}
			total += n
		case tar.TypeSymlink:
			target := filepath.Clean(filepath.Join(filepath.Dir(name), h.Linkname))
			if filepath.IsAbs(h.Linkname) || !filepath.IsLocal(target) || strings.ContainsAny(h.Linkname, "\\:\x00") {
				return errors.New("unsafe Java runtime symlink")
			}
			links[name] = h.Linkname
		default:
			return fmt.Errorf("unsupported Java archive entry type %d", h.Typeflag)
		}
	}
	// Symlinks are created last: they cannot redirect any extraction writes.
	for name, target := range links {
		if err = root.Symlink(target, name); err != nil {
			return err
		}
	}
	// Resolve the completed link graph before any executable probe. Individually
	// local relative targets can still escape through a directory-link chain.
	for name := range links {
		real, e := filepath.EvalSymlinks(filepath.Join(dest, name))
		if e != nil {
			return e
		}
		rel, e := filepath.Rel(dest, real)
		if e != nil || !filepath.IsLocal(rel) {
			return errors.New("Java runtime symlink graph escapes installation")
		}
	}
	return nil
}
