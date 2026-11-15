package runner

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	javaMetadataLimit    = 2 << 20
	javaDownloadLimit    = 512 << 20
	javaExpandedLimit    = 2 << 30
	javaEntryLimit       = 20000
	javaZIPMetadataLimit = 8 << 20
)

type temurinPackage struct {
	Link     string `json:"link"`
	Checksum string `json:"checksum"`
	Size     int64  `json:"size"`
	Name     string `json:"name"`
}

type temurinAsset struct {
	Binary struct {
		Architecture string         `json:"architecture"`
		OS           string         `json:"os"`
		ImageType    string         `json:"image_type"`
		JVMImpl      string         `json:"jvm_impl"`
		Package      temurinPackage `json:"package"`
	} `json:"binary"`
	Version struct {
		Major int `json:"major"`
	} `json:"version"`
}

// Internal seams keep unit tests offline; neither the API nor an environment
// variable can supply metadata URLs, checksums, executable probes or HTTP clients.
type windowsJavaInstaller struct {
	client *http.Client
	detect func(string) string
	rename func(*os.Root, string, string) error
}

func defaultWindowsJavaInstaller() windowsJavaInstaller {
	return windowsJavaInstaller{
		client: &http.Client{Timeout: 5 * time.Minute},
		detect: detectWindowsJavaMajor,
		rename: func(r *os.Root, a, b string) error { return r.Rename(a, b) },
	}
}

func javaHTTPSURL(u *url.URL) bool {
	return u != nil && u.Scheme == "https" && u.User == nil && u.Port() == "" && u.Fragment == ""
}

func officialJavaPackageURL(raw, ver string) bool {
	u, err := url.Parse(raw)
	return err == nil && javaHTTPSURL(u) && u.Host == "github.com" &&
		strings.HasPrefix(u.Path, "/adoptium/temurin"+ver+"-binaries/releases/download/") &&
		strings.HasSuffix(u.Path, ".zip") && !strings.Contains(u.Path, "/../") && !strings.Contains(u.Path, "\\")
}

func javaHTTPClient(base *http.Client, ver string) *http.Client {
	c := *base
	c.Timeout = 5 * time.Minute
	c.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= 5 || !javaHTTPSURL(req.URL) {
			return fmt.Errorf("unsafe Java download redirect")
		}
		if req.URL.Host == "api.adoptium.net" && strings.HasPrefix(req.URL.Path, "/v3/assets/latest/") {
			return nil
		}
		if officialJavaPackageURL(req.URL.String(), ver) {
			return nil
		}
		// GitHub's signed release-asset CDN, not arbitrary administrator URLs.
		if req.URL.Host == "release-assets.githubusercontent.com" || req.URL.Host == "objects.githubusercontent.com" {
			return nil
		}
		return fmt.Errorf("untrusted Java download redirect host %q", req.URL.Host)
	}
	return &c
}

type javaHTTPStatusError int

func (e javaHTTPStatusError) Error() string {
	return fmt.Sprintf("Adoptium download returned HTTP %d", int(e))
}

func javaGET(ctx context.Context, c *http.Client, endpoint string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "NeoPanel Java installer")
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, javaHTTPStatusError(resp.StatusCode)
	}
	return resp, nil
}

func fetchTemurinPackage(ctx context.Context, c *http.Client, ver, arch string) (temurinPackage, error) {
	for _, image := range []string{"jre", "jdk"} {
		endpoint := fmt.Sprintf("https://api.adoptium.net/v3/assets/latest/%s/hotspot?architecture=%s&image_type=%s&os=windows&vendor=eclipse", ver, arch, image)
		resp, err := javaGET(ctx, c, endpoint)
		if err != nil {
			if err == javaHTTPStatusError(http.StatusNotFound) || err == javaHTTPStatusError(http.StatusNoContent) {
				continue // No JRE for this release/architecture: try the JDK.
			}
			return temurinPackage{}, err
		}
		body, err := io.ReadAll(io.LimitReader(resp.Body, javaMetadataLimit+1))
		resp.Body.Close()
		if err != nil {
			return temurinPackage{}, err
		}
		if len(body) > javaMetadataLimit {
			return temurinPackage{}, fmt.Errorf("Adoptium metadata exceeds limit")
		}
		var assets []temurinAsset
		if err := json.Unmarshal(body, &assets); err != nil {
			return temurinPackage{}, fmt.Errorf("invalid Adoptium metadata: %w", err)
		}
		for _, a := range assets {
			if strconv.Itoa(a.Version.Major) != ver || a.Binary.Architecture != arch || a.Binary.OS != "windows" || a.Binary.ImageType != image || a.Binary.JVMImpl != "hotspot" {
				continue
			}
			p := a.Binary.Package
			checksum, err := hex.DecodeString(p.Checksum)
			if err != nil || len(checksum) != sha256.Size || p.Size <= 0 || p.Size > javaDownloadLimit || !strings.HasSuffix(p.Name, ".zip") || !officialJavaPackageURL(p.Link, ver) {
				return temurinPackage{}, fmt.Errorf("unsafe or invalid Temurin package metadata")
			}
			return p, nil
		}
	}
	return temurinPackage{}, fmt.Errorf("no official Temurin Java %s Windows/%s runtime available", ver, arch)
}

func downloadTemurinZIP(ctx context.Context, c *http.Client, p temurinPackage, target string) error {
	resp, err := javaGET(ctx, c, p.Link)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.ContentLength > javaDownloadLimit || (resp.ContentLength >= 0 && resp.ContentLength != p.Size) {
		return fmt.Errorf("Java download size mismatch")
	}
	f, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	h := sha256.New()
	n, copyErr := io.Copy(io.MultiWriter(f, h), io.LimitReader(resp.Body, p.Size+1))
	closeErr := f.Close()
	if copyErr != nil {
		return copyErr
	}
	if closeErr != nil {
		return closeErr
	}
	if n != p.Size {
		return fmt.Errorf("Java download size mismatch")
	}
	if !strings.EqualFold(hex.EncodeToString(h.Sum(nil)), p.Checksum) {
		return fmt.Errorf("Java download SHA256 mismatch")
	}
	return nil
}

func installWindowsJava(ver, goarch, rootPath string, deps windowsJavaInstaller) (JavaRuntimeInfo, error) {
	if !javaSupported[ver] {
		return JavaRuntimeInfo{}, fmt.Errorf("unsupported Java version %q", ver)
	}
	arch := map[string]string{"amd64": "x64", "arm64": "aarch64"}[goarch]
	if arch == "" {
		return JavaRuntimeInfo{}, fmt.Errorf("unsupported Windows Java architecture %s", goarch)
	}
	if !filepath.IsAbs(rootPath) {
		return JavaRuntimeInfo{}, fmt.Errorf("Java install root must be absolute")
	}
	if err := os.MkdirAll(rootPath, 0700); err != nil {
		return JavaRuntimeInfo{}, err
	}
	root, err := os.OpenRoot(rootPath)
	if err != nil {
		return JavaRuntimeInfo{}, err
	}
	defer root.Close()
	// Serialize separate panel processes too. A crash may leave this lock behind;
	// an operator must remove it only after confirming no installer is running.
	lock, err := root.OpenFile(".install.lock", os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return JavaRuntimeInfo{}, fmt.Errorf("Java installer is busy (or stale .install.lock): %w", err)
	}
	lock.Close()
	defer root.Remove(".install.lock")
	dest := "java-" + ver + "-temurin"
	finalBin := filepath.Join(rootPath, dest, "bin", "java.exe")
	info := JavaRuntimeInfo{Version: ver, Name: premadeDisplayName(ver), Path: finalBin, Installed: true}
	if st, e := root.Lstat(dest); e == nil {
		if !st.IsDir() || st.Mode()&os.ModeSymlink != 0 {
			return JavaRuntimeInfo{}, fmt.Errorf("unsafe Java destination")
		}
		if deps.detect(finalBin) == ver {
			return info, nil
		}
	} else if !os.IsNotExist(e) {
		return JavaRuntimeInfo{}, e
	}
	workPath, err := os.MkdirTemp(rootPath, ".java-"+ver+"-stage-")
	if err != nil {
		return JavaRuntimeInfo{}, err
	}
	work := filepath.Base(workPath)
	preserveBackup := false
	defer func() {
		if !preserveBackup {
			_ = root.RemoveAll(work)
		}
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	client := javaHTTPClient(deps.client, ver)
	p, err := fetchTemurinPackage(ctx, client, ver, arch)
	if err != nil {
		return JavaRuntimeInfo{}, err
	}
	archive := filepath.Join(workPath, "runtime.zip")
	if err := downloadTemurinZIP(ctx, client, p, archive); err != nil {
		return JavaRuntimeInfo{}, err
	}
	payload := filepath.Join(work, "payload")
	if err := root.Mkdir(payload, 0700); err != nil {
		return JavaRuntimeInfo{}, err
	}
	// No downloaded code is executed until the authenticated metadata checksum
	// and the entire archive have both passed validation.
	if err := extractWindowsJavaZIP(ctx, archive, filepath.Join(rootPath, payload)); err != nil {
		return JavaRuntimeInfo{}, err
	}
	if deps.detect(filepath.Join(rootPath, payload, "bin", "java.exe")) != ver {
		return JavaRuntimeInfo{}, fmt.Errorf("downloaded Java failed major-version %s validation", ver)
	}
	backup := filepath.Join(work, "previous")
	hadPrevious := false
	if _, err := root.Lstat(dest); err == nil {
		if err := deps.rename(root, dest, backup); err != nil {
			return JavaRuntimeInfo{}, fmt.Errorf("preserve previous Java runtime: %w", err)
		}
		hadPrevious = true
	} else if !os.IsNotExist(err) {
		return JavaRuntimeInfo{}, err
	}
	rollback := func(cause error, published bool) (JavaRuntimeInfo, error) {
		if published {
			if err := root.Rename(dest, filepath.Join(work, "failed")); err != nil {
				preserveBackup = true
				return JavaRuntimeInfo{}, fmt.Errorf("%w; rollback could not remove new runtime: %v; recovery directory: %s", cause, err, workPath)
			}
		}
		if hadPrevious {
			if err := deps.rename(root, backup, dest); err != nil {
				preserveBackup = true
				return JavaRuntimeInfo{}, fmt.Errorf("%w; rollback failed: %v; previous runtime preserved at %s", cause, err, filepath.Join(workPath, "previous"))
			}
		}
		return JavaRuntimeInfo{}, cause
	}
	if err := deps.rename(root, payload, dest); err != nil {
		return rollback(fmt.Errorf("publish Java runtime: %w", err), false)
	}
	if deps.detect(finalBin) != ver {
		return rollback(fmt.Errorf("published Java failed major-version %s validation", ver), true)
	}
	return info, nil
}
