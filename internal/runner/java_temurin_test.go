package runner

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type javaRoundTrip func(*http.Request) (*http.Response, error)

func (f javaRoundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type javaZIPMember struct {
	name, body string
	mode       os.FileMode
}

func javaTestZIP(t *testing.T, members ...javaZIPMember) []byte {
	t.Helper()
	var b bytes.Buffer
	w := zip.NewWriter(&b)
	for _, m := range members {
		h := &zip.FileHeader{Name: m.name, Method: zip.Deflate}
		if m.mode == 0 {
			m.mode = 0600
		}
		h.SetMode(m.mode)
		out, err := w.CreateHeader(h)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := io.WriteString(out, m.body); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

func javaTestInstaller(t *testing.T, archive []byte, mutate func(*temurinAsset)) (windowsJavaInstaller, *int) {
	t.Helper()
	hash := sha256.Sum256(archive)
	var asset temurinAsset
	asset.Version.Major = 21
	asset.Binary.Architecture, asset.Binary.OS, asset.Binary.ImageType, asset.Binary.JVMImpl = "x64", "windows", "jre", "hotspot"
	asset.Binary.Package = temurinPackage{Link: "https://github.com/adoptium/temurin21-binaries/releases/download/jdk-21/runtime.zip", Checksum: hex.EncodeToString(hash[:]), Size: int64(len(archive)), Name: "runtime.zip"}
	if mutate != nil {
		mutate(&asset)
	}
	meta, err := json.Marshal([]temurinAsset{asset})
	if err != nil {
		t.Fatal(err)
	}
	calls := new(int)
	deps := defaultWindowsJavaInstaller()
	deps.client = &http.Client{Transport: javaRoundTrip(func(r *http.Request) (*http.Response, error) {
		*calls++
		var data []byte
		switch r.URL.Host {
		case "api.adoptium.net":
			data = meta
		case "github.com":
			data = archive
		default:
			t.Fatalf("unexpected network request: %s", r.URL)
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader(data)), ContentLength: int64(len(data)), Header: make(http.Header), Request: r}, nil
	})}
	deps.detect = func(bin string) string { data, _ := os.ReadFile(bin); return string(data) }
	return deps, calls
}

func TestJavaWindowsInstallAndReuse(t *testing.T) {
	archive := javaTestZIP(t, javaZIPMember{name: "jdk-21/bin/java.exe", body: "21"}, javaZIPMember{name: "jdk-21/lib/modules", body: "modules"})
	deps, calls := javaTestInstaller(t, archive, nil)
	root := filepath.Join(t.TempDir(), "Runtime Folder With Spaces")
	info, err := installWindowsJava("21", "amd64", root, deps)
	if err != nil {
		t.Fatal(err)
	}
	if !info.Installed || info.Version != "21" || !filepath.IsAbs(info.Path) {
		t.Fatalf("bad runtime: %+v", info)
	}
	if *calls != 2 {
		t.Fatalf("requests = %d", *calls)
	}
	if _, err := installWindowsJava("21", "amd64", root, deps); err != nil {
		t.Fatal(err)
	}
	if *calls != 2 {
		t.Fatal("existing runtime was downloaded again")
	}
	entries, _ := os.ReadDir(root)
	if len(entries) != 1 || entries[0].Name() != "java-21-temurin" {
		t.Fatalf("staging/lock leaked: %v", entries)
	}
}

func TestJavaWindowsInstallFailuresPreserveExisting(t *testing.T) {
	for _, scenario := range []string{"checksum", "partial", "traversal", "missing-binary", "version-mismatch", "metadata-major", "metadata-url", "metadata-size", "publish", "post-publish", "rollback"} {
		t.Run(scenario, func(t *testing.T) {
			members := []javaZIPMember{{name: "jdk-21/bin/java.exe", body: "21"}}
			if scenario == "traversal" {
				members = append(members, javaZIPMember{name: "jdk-21/../escape", body: "escape"})
			}
			if scenario == "missing-binary" {
				members[0].name = "jdk-21/lib/modules"
			}
			if scenario == "version-mismatch" {
				members[0].body = "17"
			}
			archive := javaTestZIP(t, members...)
			deps, _ := javaTestInstaller(t, archive, func(a *temurinAsset) {
				switch scenario {
				case "checksum":
					a.Binary.Package.Checksum = strings.Repeat("0", 64)
				case "partial":
					a.Binary.Package.Size++
				case "metadata-major":
					a.Version.Major = 17
				case "metadata-url":
					a.Binary.Package.Link = "http://127.0.0.1/runtime.zip"
				case "metadata-size":
					a.Binary.Package.Size = javaDownloadLimit + 1
				}
			})
			root := t.TempDir()
			old := filepath.Join(root, "java-21-temurin", "bin", "java.exe")
			if err := os.MkdirAll(filepath.Dir(old), 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(old, []byte("broken previous installation"), 0600); err != nil {
				t.Fatal(err)
			}
			probes := 0
			detect := deps.detect
			deps.detect = func(path string) string {
				if strings.Contains(path, "-stage-") {
					probes++
				}
				if scenario == "post-publish" && path == old {
					return ""
				}
				return detect(path)
			}
			if scenario == "publish" || scenario == "rollback" {
				deps.rename = func(r *os.Root, a, b string) error {
					if filepath.Base(a) == "payload" || (scenario == "rollback" && filepath.Base(a) == "previous") {
						return fmt.Errorf("injected rename failure")
					}
					return r.Rename(a, b)
				}
			}
			_, err := installWindowsJava("21", "amd64", root, deps)
			if err == nil {
				t.Fatal("accepted failed installation")
			}
			if scenario == "rollback" {
				backups, _ := filepath.Glob(filepath.Join(root, ".java-21-stage-*", "previous", "bin", "java.exe"))
				if len(backups) != 1 || !strings.Contains(err.Error(), "previous runtime preserved") {
					t.Fatalf("lost rollback backup: %v, %v", backups, err)
				}
				old = backups[0]
			} else {
				entries, _ := os.ReadDir(root)
				if len(entries) != 1 {
					t.Fatalf("partial staging leaked: %v", entries)
				}
			}
			data, readErr := os.ReadFile(old)
			if readErr != nil || string(data) != "broken previous installation" {
				t.Fatalf("previous runtime damaged: %q %v", data, readErr)
			}
			if (scenario == "checksum" || scenario == "partial" || scenario == "traversal" || strings.HasPrefix(scenario, "metadata-")) && probes != 0 {
				t.Fatal("executed an unverified download")
			}
		})
	}
}

func TestJavaWindowsZIPRejectsUnsafeMembers(t *testing.T) {
	for _, name := range []string{"../bin/java.exe", "/jdk/bin/java.exe", "C:/jdk/bin/java.exe", "jdk/../escape", "jdk/bin/../../escape", `jdk\bin\java.exe`, "jdk/bin/java.exe:evil", "jdk/NUL.txt", "jdk/COM1", "jdk/LPT².txt", "CON/bin/java.exe", "jdk/bin./java.exe", "jdk/bin /java.exe", "jdk//java.exe", "jdk/./java.exe", "jdk/bin/evil\x01", "jdk/bin/foo?"} {
		t.Run(name, func(t *testing.T) {
			if _, _, err := windowsJavaZIPName(name); err == nil {
				t.Fatalf("accepted %q", name)
			}
		})
	}
	cases := map[string][]javaZIPMember{
		"symlink":   {{name: "jdk/bin/java.exe", body: "target", mode: os.ModeSymlink | 0777}},
		"device":    {{name: "jdk/bin/java.exe", mode: os.ModeDevice | 0600}},
		"duplicate": {{name: "jdk/bin/java.exe"}, {name: "jdk/BIN/JAVA.EXE"}},
		"roots":     {{name: "jdk/bin/java.exe"}, {name: "other/bin/java.exe"}},
		"no-root":   {{name: "java.exe"}},
	}
	for name, members := range cases {
		t.Run(name, func(t *testing.T) {
			archive := filepath.Join(t.TempDir(), "runtime.zip")
			if err := os.WriteFile(archive, javaTestZIP(t, members...), 0600); err != nil {
				t.Fatal(err)
			}
			if err := extractWindowsJavaZIP(context.Background(), archive, t.TempDir()); err == nil {
				t.Fatal("accepted unsafe archive")
			}
		})
	}
}

func TestJavaWindowsZIPLimitsAndCRC(t *testing.T) {
	original := javaTestZIP(t, javaZIPMember{name: "jdk/bin/java.exe", body: "21"})
	for _, scenario := range []string{"entries", "metadata", "expanded", "crc", "truncated", "encrypted"} {
		t.Run(scenario, func(t *testing.T) {
			data := bytes.Clone(original)
			end := len(data) - 22
			central := bytes.Index(data, []byte("PK\x01\x02"))
			switch scenario {
			case "entries":
				binary.LittleEndian.PutUint16(data[end+10:], javaEntryLimit+1)
			case "metadata":
				binary.LittleEndian.PutUint32(data[end+12:], javaZIPMetadataLimit+1)
			case "expanded":
				binary.LittleEndian.PutUint32(data[central+24:], javaDownloadLimit+1)
			case "crc":
				data[central+16] ^= 0xff
			case "truncated":
				data = data[:len(data)-5]
			case "encrypted":
				data[central+8] |= 1
			}
			archive := filepath.Join(t.TempDir(), "runtime.zip")
			if err := os.WriteFile(archive, data, 0600); err != nil {
				t.Fatal(err)
			}
			if err := extractWindowsJavaZIP(context.Background(), archive, t.TempDir()); err == nil {
				t.Fatal("accepted corrupt/excessive ZIP")
			}
		})
	}
}

func TestJavaWindowsMetadataFallbackAndRedirects(t *testing.T) {
	archive := javaTestZIP(t, javaZIPMember{name: "jdk/bin/java.exe", body: "21"})
	deps, _ := javaTestInstaller(t, archive, func(a *temurinAsset) { a.Binary.ImageType = "jdk" })
	if _, err := fetchTemurinPackage(context.Background(), javaHTTPClient(deps.client, "21"), "21", "x64"); err != nil {
		t.Fatalf("JDK fallback: %v", err)
	}
	transport := deps.client.Transport
	deps.client.Transport = javaRoundTrip(func(r *http.Request) (*http.Response, error) {
		if r.URL.Query().Get("image_type") == "jre" {
			return &http.Response{StatusCode: 404, Body: io.NopCloser(strings.NewReader("")), Request: r}, nil
		}
		return transport.RoundTrip(r)
	})
	if _, err := fetchTemurinPackage(context.Background(), javaHTTPClient(deps.client, "21"), "21", "x64"); err != nil {
		t.Fatalf("404 JRE fallback: %v", err)
	}
	for _, target := range []string{"http://github.com/adoptium/temurin21-binaries/releases/download/a/b.zip", "https://127.0.0.1/runtime.zip", "https://github.com/evil/runtime/releases/download/a/b.zip", "https://api.adoptium.net.evil.test/a", "https://user:pass@release-assets.githubusercontent.com/a"} {
		base := &http.Client{Transport: javaRoundTrip(func(r *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 302, Header: http.Header{"Location": {target}}, Body: io.NopCloser(strings.NewReader("")), Request: r}, nil
		})}
		if _, err := javaGET(context.Background(), javaHTTPClient(base, "21"), "https://api.adoptium.net/v3/assets/latest/21/hotspot"); err == nil {
			t.Fatalf("followed unsafe redirect %s", target)
		}
	}
	base := &http.Client{Transport: javaRoundTrip(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(strings.Repeat(" ", javaMetadataLimit+1))), ContentLength: -1, Request: r}, nil
	})}
	if _, err := fetchTemurinPackage(context.Background(), javaHTTPClient(base, "21"), "21", "x64"); err == nil {
		t.Fatal("accepted oversized metadata")
	}
}

func TestJavaWindowsSupportedVersionsAndArchitectures(t *testing.T) {
	for _, ver := range []string{"25", "21", "17", "8"} {
		for _, arch := range []string{"amd64", "arm64"} {
			t.Run(ver+"/"+arch, func(t *testing.T) {
				archive := javaTestZIP(t, javaZIPMember{name: "jdk/bin/java.exe", body: ver})
				deps, _ := javaTestInstaller(t, archive, func(a *temurinAsset) {
					fmt.Sscan(ver, &a.Version.Major)
					a.Binary.Architecture = map[string]string{"amd64": "x64", "arm64": "aarch64"}[arch]
					a.Binary.Package.Link = "https://github.com/adoptium/temurin" + ver + "-binaries/releases/download/tag/runtime.zip"
				})
				if _, err := installWindowsJava(ver, arch, t.TempDir(), deps); err != nil {
					t.Fatal(err)
				}
			})
		}
	}
	deps, _ := javaTestInstaller(t, nil, nil)
	if _, err := installWindowsJava("21", "386", t.TempDir(), deps); err == nil {
		t.Fatal("accepted unsupported architecture")
	}
	if _, err := installWindowsJava("11", "amd64", t.TempDir(), deps); err == nil {
		t.Fatal("accepted unsupported release")
	}
}

func TestJavaWindowsDownloadUnknownLength(t *testing.T) {
	data := []byte("checksum verified archive")
	hash := sha256.Sum256(data)
	p := temurinPackage{Link: "https://github.com/adoptium/temurin21-binaries/releases/download/tag/runtime.zip", Size: int64(len(data)), Checksum: hex.EncodeToString(hash[:])}
	for _, scenario := range []string{"valid", "short", "long"} {
		t.Run(scenario, func(t *testing.T) {
			body := bytes.Clone(data)
			if scenario == "short" {
				body = body[:len(body)-1]
			}
			if scenario == "long" {
				body = append(body, 'x')
			}
			client := &http.Client{Transport: javaRoundTrip(func(r *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader(body)), ContentLength: -1, Request: r}, nil
			})}
			err := downloadTemurinZIP(context.Background(), client, p, filepath.Join(t.TempDir(), "runtime.zip"))
			if (err == nil) != (scenario == "valid") {
				t.Fatalf("%s: %v", scenario, err)
			}
		})
	}
}

func TestJavaWindowsInstallerLock(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".install.lock"), nil, 0600); err != nil {
		t.Fatal(err)
	}
	deps, calls := javaTestInstaller(t, nil, nil)
	if _, err := installWindowsJava("21", "amd64", root, deps); err == nil {
		t.Fatal("ignored concurrent/stale installer lock")
	}
	if *calls != 0 {
		t.Fatal("downloaded despite lock")
	}
}
