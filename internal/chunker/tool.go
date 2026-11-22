// Package chunker is the pinned, non-interactive Chunker CLI adapter.
package chunker

import (
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const Version = "1.21.0"
const JarURL = "https://github.com/HiveGamesOSS/Chunker/releases/download/1.21.0/chunker-cli-1.21.0.jar"
const JarSHA256 = "63148e587f5ee5a2ca7ebae1079b235c82ae4d1eff8f7a8a8a3398332513dda6"
const JarSize int64 = 31883497

var installMu sync.Mutex

//go:embed LICENSE.chunker
var licenseNotice []byte

// EnsureTool accepts only daemon-controlled storage, never a request-supplied path.
// An existing artifact is rehashed before every use. Publication is atomic.
func EnsureTool(ctx context.Context, dir string) (string, error) {
	for !installMu.TryLock() {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(100 * time.Millisecond):
		}
	}
	defer installMu.Unlock()
	if err := PrivateDir(dir); err != nil {
		return "", err
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		return "", err
	}
	defer root.Close()
	if _, err = root.Lstat("LICENSE.chunker"); os.IsNotExist(err) {
		f, e := root.OpenFile("LICENSE.chunker", os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if e != nil {
			return "", e
		}
		_, e = f.Write(licenseNotice)
		e = errors.Join(e, f.Close())
		if e != nil {
			return "", e
		}
	}
	target := filepath.Join(dir, "chunker-cli-"+Version+".jar")
	if err := verifyFile(target, JarSize, JarSHA256); err == nil {
		return target, nil
	}
	client := &http.Client{Timeout: 3 * time.Minute, CheckRedirect: func(r *http.Request, via []*http.Request) error {
		if len(via) >= 5 || r.URL.Scheme != "https" || r.URL.User != nil || r.URL.Port() != "" || !downloadHost(r.URL.Hostname()) {
			return errors.New("untrusted tool download redirect")
		}
		return nil
	}}
	if err := download(ctx, client, JarURL, target, JarSize, JarSHA256); err != nil {
		return "", err
	}
	return target, nil
}
func downloadHost(h string) bool {
	return h == "github.com" || h == "release-assets.githubusercontent.com" || h == "objects.githubusercontent.com"
}

func PrivateDir(dir string) error {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	// Reject symlink components, including existing parents. Never chmod a linked path.
	for p := abs; ; p = filepath.Dir(p) {
		st, e := os.Lstat(p)
		if e == nil && (!st.IsDir() || st.Mode()&os.ModeSymlink != 0) {
			return errors.New("tool storage must use real directories")
		}
		if e != nil && !os.IsNotExist(e) {
			return e
		}
		if filepath.Dir(p) == p {
			break
		}
	}
	if err := os.MkdirAll(abs, 0700); err != nil {
		return err
	}
	return os.Chmod(abs, 0700)
}
func verifyFile(path string, size int64, digest string) error {
	st, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !st.Mode().IsRegular() || st.Size() != size {
		return errors.New("invalid tool artifact")
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	actual, err := f.Stat()
	if err != nil || !os.SameFile(st, actual) {
		return errors.New("tool artifact changed")
	}
	h := sha256.New()
	n, err := io.Copy(h, io.LimitReader(f, size+1))
	if err != nil {
		return err
	}
	if n != size || hex.EncodeToString(h.Sum(nil)) != digest {
		return errors.New("tool checksum mismatch")
	}
	return nil
}
func download(ctx context.Context, client *http.Client, url, target string, size int64, digest string) error {
	if !strings.HasPrefix(url, "https://") {
		return errors.New("tool download requires HTTPS")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 || (resp.ContentLength >= 0 && resp.ContentLength != size) {
		return errors.New("unexpected tool download response")
	}
	f, err := os.CreateTemp(filepath.Dir(target), ".chunker-download-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	defer f.Close()
	h := sha256.New()
	n, err := io.Copy(io.MultiWriter(f, h), io.LimitReader(resp.Body, size+1))
	if err != nil {
		return err
	}
	if n != size || hex.EncodeToString(h.Sum(nil)) != digest {
		return fmt.Errorf("Chunker download failed pinned size/SHA-256 verification")
	}
	if err = f.Sync(); err != nil {
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	if err = os.Chmod(f.Name(), 0400); err != nil {
		return err
	}
	return os.Rename(f.Name(), target)
}

// Reviewed subset of release commit 50aaf6d82e8b68929279161b790d95fdddfe0a40.
// No nearest-version fallback, INPUT, PREVIEW, SETTINGS or arbitrary CLI flags.
type Format struct {
	ID          string `json:"id"`
	Edition     string `json:"edition"`
	Version     string `json:"version"`
	DataVersion int32  `json:"-"`
}

var formats = []Format{
	{"JAVA_1_19_4", "java", "1.19.4", 3337},
	{"JAVA_1_20", "java", "1.20", 3463},
	{"JAVA_1_20_1", "java", "1.20.1", 3465},
	{"JAVA_1_20_2", "java", "1.20.2", 3578},
	{"JAVA_1_20_3", "java", "1.20.3", 3698},
	{"JAVA_1_20_4", "java", "1.20.4", 3700},
	{"JAVA_1_20_5", "java", "1.20.5", 3837},
	{"JAVA_1_20_6", "java", "1.20.6", 3839},
	{"JAVA_1_21", "java", "1.21", 3953},
	{"JAVA_1_21_1", "java", "1.21.1", 3955},
	{"JAVA_1_21_2", "java", "1.21.2", 4080},
	{"JAVA_1_21_3", "java", "1.21.3", 4082},
	{"JAVA_1_21_4", "java", "1.21.4", 4189},
	{"JAVA_1_21_5", "java", "1.21.5", 4325},
	{"JAVA_1_21_6", "java", "1.21.6", 4435},
	{"JAVA_1_21_7", "java", "1.21.7", 4438},
	{"JAVA_1_21_8", "java", "1.21.8", 4440},
	{"JAVA_1_21_9", "java", "1.21.9", 4554},
	{"JAVA_1_21_10", "java", "1.21.10", 4556},
	{"JAVA_1_21_11", "java", "1.21.11", 4671},
	{"JAVA_26_1", "java", "26.1", 4786},
	{"JAVA_26_1_1", "java", "26.1.1", 4788},
	{"JAVA_26_1_2", "java", "26.1.2", 4790},
	{"JAVA_26_2", "java", "26.2", 4903},
	{"JAVA_26_3", "java", "26.3", 5017},
	{"BEDROCK_R20_80", "bedrock", "1.20.80", 671},
	{"BEDROCK_R21", "bedrock", "1.21", 685},
}

func Formats() []Format { return append([]Format(nil), formats...) }
func Lookup(id string) (Format, bool) {
	for _, f := range formats {
		if f.ID == id {
			return f, true
		}
	}
	return Format{}, false
}
