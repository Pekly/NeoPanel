package downloader

import (
	"archive/zip"
	"context"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/Pekly/NeoPanel/internal/rooted"
	"github.com/Pekly/NeoPanel/internal/serverjar"
)

// DownloadServerJarWithInfo stages and validates all bytes before replacing files.
// Callers must serialize server maintenance against other writers through publication.
// Steel retains its native installer and does not contact MCJars.
func DownloadServerJarWithInfo(ctx context.Context, flavor, version, destDir, jarName string) (serverjar.Info, error) {
	var empty serverjar.Info
	if err := ValidateSoftware(flavor, version); err != nil {
		return empty, err
	}
	if !rooted.Basename(jarName) || strings.TrimRight(jarName, ". ") != jarName || strings.EqualFold(jarName, "libraries") || strings.EqualFold(jarName, "eula.txt") || strings.HasPrefix(strings.ToLower(jarName), ".neopanel-") {
		return empty, fmt.Errorf("invalid server filename")
	}
	if err := ctx.Err(); err != nil {
		return empty, err
	}
	if IsNativeSoftware(flavor) {
		return serverjar.Info{Status: "native", Source: "official", Software: "steel", Filename: jarName}, downloadSteel(version, destDir, jarName)
	}
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return empty, err
	}
	root, err := rooted.Open(destDir)
	if err != nil {
		return empty, err
	}
	defer root.Close()
	build, err := ResolveMCJarsBuild(ctx, flavor, version)
	if err != nil {
		return empty, err
	}
	stageName, stage, err := root.TempDir()
	if err != nil {
		return empty, err
	}
	keep := false
	defer func() {
		stage.Close()
		if !keep {
			root.RemoveAll(stageName)
		}
	}()
	if err = mcjarsDownload(ctx, stage, build.download); err != nil {
		return empty, err
	}
	if build.prepared {
		if err = mcjarsExtract(ctx, stage); err != nil {
			return empty, err
		}
	} else if err = stage.Rename("payload", "server.jar"); err != nil {
		return empty, err
	}
	if err = mcjarsValidateJar(ctx, stage, "server.jar"); err != nil {
		return empty, err
	}
	if err = stage.Chmod("server.jar", 0644); err != nil {
		return empty, err
	}
	info := build.Info
	f, err := stage.Open("server.jar")
	if err != nil {
		return empty, err
	}
	sum := sha256.New()
	n, err := rooted.Copy(ctx, sum, f, mcjarsBinaryLimit)
	st, statErr := f.Stat()
	f.Close()
	if err != nil {
		return empty, err
	}
	if statErr != nil {
		return empty, statErr
	}
	info.SHA256, info.Size, info.Modified, info.Filename = hex.EncodeToString(sum.Sum(nil)), n, st.ModTime(), jarName
	names := []mcjarsPublication{{from: "server.jar", to: jarName}}
	if build.prepared {
		names = append(names, mcjarsPublication{from: "libraries", to: "libraries"})
	}
	if !mcjarsProxy(strings.ToUpper(info.Software)) {
		if err = stage.WriteFile("eula.txt", []byte("# EULA auto-accepted by NeoPanel\neula=true\n"), 0644); err != nil {
			return empty, err
		}
		names = append(names, mcjarsPublication{from: "eula.txt", to: "eula.txt"})
	}
	if err = ctx.Err(); err != nil {
		return empty, err
	}
	keep, err = mcjarsPublish(root, stageName, names)
	if err != nil {
		return empty, err
	}
	return info, nil
}

// Hashes are derived only from documented content-addressed official URL paths.
func mcjarsPathHash(raw string) (hash.Hash, string) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, ""
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if u.Hostname() == "fill-data.papermc.io" && len(parts) == 4 && parts[0] == "v1" && parts[1] == "objects" && mcjarsSHA256.MatchString(parts[2]) {
		return sha256.New(), strings.ToLower(parts[2])
	}
	if (u.Hostname() == "piston-data.mojang.com" || u.Hostname() == "launcher.mojang.com") && len(parts) >= 3 && (parts[0] == "v1" || parts[0] == "v2") && parts[1] == "objects" && len(parts[2]) == 40 {
		if _, err := hex.DecodeString(parts[2]); err == nil {
			return sha1.New(), strings.ToLower(parts[2])
		}
	}
	return nil, ""
}
func mcjarsDownload(ctx context.Context, stage *rooted.Root, d mcjarsStep) error {
	if _, err := mcjarsDownloadURL(d.URL); err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, d.URL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", userAgent)
	client := mcjarsClient(true)
	defer client.CloseIdleConnections()
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("MCJars download HTTP %d", resp.StatusCode)
	}
	if resp.ContentLength > 0 && resp.ContentLength != d.Size {
		return fmt.Errorf("MCJars Content-Length mismatch")
	}
	var input io.Reader = resp.Body
	var hashes []hash.Hash
	var expected []string
	for _, raw := range []string{d.URL, reqURL(resp)} {
		if h, want := mcjarsPathHash(raw); h != nil {
			hashes = append(hashes, h)
			expected = append(expected, want)
			input = io.TeeReader(input, h)
		}
	}
	n, err := stage.Copy(ctx, "payload", input, d.Size, 0600)
	if err != nil {
		return err
	}
	if n != d.Size {
		return fmt.Errorf("MCJars download size mismatch: got %d, want %d", n, d.Size)
	}
	for i, h := range hashes {
		if hex.EncodeToString(h.Sum(nil)) != expected[i] {
			return fmt.Errorf("MCJars official path checksum mismatch")
		}
	}
	return nil
}
func reqURL(resp *http.Response) string {
	if resp.Request == nil {
		return ""
	}
	return resp.Request.URL.String()
}

// Reject ZIP64/multi-disk archives and bound central-directory allocation before
// archive/zip parses it. Both executable JARs and prepared bundles use this path.
func mcjarsZIP(f *os.File) (*zip.Reader, error) {
	st, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if st.Size() < 22 || st.Size() > mcjarsBinaryLimit {
		return nil, fmt.Errorf("invalid ZIP size")
	}
	n := int64(65557)
	if st.Size() < n {
		n = st.Size()
	}
	tail := make([]byte, n)
	if _, err = f.ReadAt(tail, st.Size()-n); err != nil {
		return nil, err
	}
	for i := len(tail) - 22; i >= 0; i-- {
		if string(tail[i:i+4]) != "PK\x05\x06" {
			continue
		}
		e := tail[i:]
		if i+22+int(binary.LittleEndian.Uint16(e[20:])) != len(tail) {
			continue
		}
		count := binary.LittleEndian.Uint16(e[10:])
		size := binary.LittleEndian.Uint32(e[12:])
		offset := binary.LittleEndian.Uint32(e[16:])
		if binary.LittleEndian.Uint16(e[4:]) != 0 || binary.LittleEndian.Uint16(e[6:]) != 0 || binary.LittleEndian.Uint16(e[8:]) != count || count == 0 || count > 60000 || size > 8<<20 || int64(offset)+int64(size) != st.Size()-n+int64(i) {
			return nil, fmt.Errorf("ZIP metadata exceeds limits or unsupported ZIP layout")
		}
		zr, err := zip.NewReader(f, st.Size())
		if err != nil {
			return nil, err
		}
		if len(zr.File) != int(count) {
			return nil, fmt.Errorf("ZIP entry count mismatch")
		}
		return zr, nil
	}
	return nil, fmt.Errorf("invalid ZIP directory")
}

func mcjarsArchivePath(name string) (string, error) {
	n := strings.TrimSuffix(name, "/")
	if len(n) == 0 || len(n) > 1024 || strings.Count(n, "/") > 32 || strings.ContainsAny(n, "\\:\x00") || path.Clean(n) != n || strings.HasPrefix(n, "/") {
		return "", fmt.Errorf("unsafe ZIP path")
	}
	for _, part := range strings.Split(n, "/") {
		if part == "." || part == ".." || strings.TrimRight(part, ". ") != part || !rooted.Basename(part) {
			return "", fmt.Errorf("unsafe ZIP component")
		}
		for _, c := range part {
			if c < 32 || c == 127 {
				return "", fmt.Errorf("unsafe ZIP component")
			}
		}
	}
	return n, nil
}
func mcjarsZIPMembers(zr *zip.Reader, prepared bool) error {
	if len(zr.File) > 60000 || (prepared && len(zr.File) > 10000) {
		return fmt.Errorf("ZIP entry count exceeds limit")
	}
	seen := map[string]bool{}
	var expanded uint64
	for _, z := range zr.File {
		n, err := mcjarsArchivePath(z.Name)
		if err != nil {
			return err
		}
		if !z.Mode().IsRegular() && !z.Mode().IsDir() {
			return fmt.Errorf("ZIP links and special files forbidden")
		}
		if prepared && n != "server.jar" && n != "libraries" && !strings.HasPrefix(n, "libraries/") {
			return fmt.Errorf("unsupported prepared ZIP member %q", n)
		}
		if prepared && ((n == "server.jar" && z.Mode().IsDir()) || (n == "libraries" && !z.Mode().IsDir())) {
			return fmt.Errorf("prepared ZIP member type conflict")
		}
		key := strings.ToLower(n)
		if _, ok := seen[key]; ok {
			return fmt.Errorf("duplicate ZIP member")
		}
		seen[key] = z.Mode().IsDir()
		if z.UncompressedSize64 > uint64(mcjarsBinaryLimit)-expanded {
			return fmt.Errorf("ZIP expanded size exceeds limit")
		}
		expanded += z.UncompressedSize64
	}
	for n := range seen {
		for p := path.Dir(n); p != "."; p = path.Dir(p) {
			if dir, ok := seen[p]; ok && !dir {
				return fmt.Errorf("ZIP path conflict")
			}
		}
	}
	return nil
}
func mcjarsExtract(ctx context.Context, stage *rooted.Root) error {
	f, err := stage.Open("payload")
	if err != nil {
		return err
	}
	defer f.Close()
	zr, err := mcjarsZIP(f)
	if err != nil {
		return err
	}
	if err = mcjarsZIPMembers(zr, true); err != nil {
		return err
	}
	// The provider's remove/unzip instructions are not executed. Only this private
	// tree is populated, then the two approved top-level artifacts are published.
	if err = stage.Mkdir("libraries", 0755); err != nil {
		return err
	}
	var total int64
	for _, z := range zr.File {
		n, _ := mcjarsArchivePath(z.Name)
		if z.Mode().IsDir() {
			if err = stage.MkdirAll(n, 0755); err != nil {
				return err
			}
			continue
		}
		if err = stage.MkdirAll(filepath.Dir(n), 0755); err != nil {
			return err
		}
		r, err := z.Open()
		if err != nil {
			return err
		}
		got, err := stage.Copy(ctx, n, r, mcjarsBinaryLimit-total, 0644)
		r.Close()
		total += got
		if err != nil {
			return err
		}
		if uint64(got) != z.UncompressedSize64 {
			return fmt.Errorf("ZIP member size mismatch")
		}
	}
	return nil
}
func mcjarsValidateJar(ctx context.Context, stage *rooted.Root, name string) error {
	f, err := stage.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()
	zr, err := mcjarsZIP(f)
	if err != nil {
		return fmt.Errorf("invalid server JAR: %w", err)
	}
	if err = mcjarsZIPMembers(zr, false); err != nil {
		return err
	}
	manifest := false
	var total int64
	for _, z := range zr.File {
		if z.Mode().IsDir() {
			continue
		}
		r, err := z.Open()
		if err != nil {
			return err
		}
		var got int64
		if z.Name == "META-INF/MANIFEST.MF" {
			data, readErr := rooted.ReadBounded(r, 1<<20)
			err = readErr
			got = int64(len(data))
			text := strings.ReplaceAll(string(data), "\r\n", "\n")
			for _, line := range strings.Split(text, "\n") {
				if strings.HasPrefix(line, "Main-Class:") && strings.TrimSpace(strings.TrimPrefix(line, "Main-Class:")) != "" {
					manifest = true
				}
			}
		} else {
			got, err = rooted.Copy(ctx, io.Discard, r, mcjarsBinaryLimit-total)
		}
		r.Close()
		total += got
		if err != nil {
			return err
		}
		if uint64(got) != z.UncompressedSize64 || total > mcjarsBinaryLimit {
			return fmt.Errorf("JAR expanded size mismatch")
		}
	}
	if !manifest {
		return fmt.Errorf("server JAR lacks executable manifest")
	}
	return ctx.Err()
}

type mcjarsPublication struct{ from, to string }

// Rollback uses same-root renames. A failed rollback retains the private backup
// directory and names it in the error; it is never erased by deferred cleanup.
func mcjarsPublish(root *rooted.Root, stage string, items []mcjarsPublication) (keep bool, err error) {
	for _, item := range items {
		if err = root.Check(item.to); err != nil {
			return false, err
		}
		st, e := root.Lstat(item.to)
		if e != nil && !os.IsNotExist(e) {
			return false, e
		}
		if e == nil && ((item.to == "libraries" && !st.IsDir()) || (item.to != "libraries" && !st.Mode().IsRegular())) {
			return false, fmt.Errorf("publication destination type conflict: %s", item.to)
		}
	}
	moved := make([]bool, len(items))
	published := make([]bool, len(items))
	rollback := func(cause error) (bool, error) {
		var failures []error
		for i := len(items) - 1; i >= 0; i-- {
			item := items[i]
			if published[i] {
				if e := root.Rename(item.to, filepath.Join(stage, item.from)); e != nil {
					failures = append(failures, e)
					continue
				}
			}
			if moved[i] {
				if e := root.Rename(filepath.Join(stage, fmt.Sprintf("backup-%d", i)), item.to); e != nil {
					failures = append(failures, e)
				}
			}
		}
		if len(failures) > 0 {
			return true, fmt.Errorf("publication failed: %w; rollback incomplete, recover backups in %s: %v", cause, stage, errors.Join(failures...))
		}
		return false, cause
	}
	for i, item := range items {
		if err = root.Check(item.to); err != nil {
			return rollback(err)
		}
		if _, e := root.Lstat(item.to); e == nil {
			if err = root.Rename(item.to, filepath.Join(stage, fmt.Sprintf("backup-%d", i))); err != nil {
				return rollback(err)
			}
			moved[i] = true
		} else if !os.IsNotExist(e) {
			return rollback(e)
		}
		if err = root.Rename(filepath.Join(stage, item.from), item.to); err != nil {
			return rollback(err)
		}
		published[i] = true
	}
	return false, nil
}
