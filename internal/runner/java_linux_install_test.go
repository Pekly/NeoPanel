package runner

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Linux archives are tested on every host using probes of fixture contents;
// never execute Linux binaries, install runtimes or contact upstream in tests.
func javaTestTar(t *testing.T, members ...tar.Header) []byte {
	t.Helper()
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	tw := tar.NewWriter(gz)
	for _, h := range members {
		if h.Mode == 0 {
			h.Mode = 0755
		}
		if h.Typeflag == 0 {
			h.Typeflag = tar.TypeReg
		}
		body := h.PAXRecords["test.body"]
		h.PAXRecords = nil
		h.Size = int64(len(body))
		if err := tw.WriteHeader(&h); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(body)); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

func linuxJavaFixture(t *testing.T, archive []byte, ver, arch string, mutate func(*temurinAsset)) (windowsJavaInstaller, *int) {
	return javaTestInstaller(t, archive, func(a *temurinAsset) {
		fmt.Sscan(ver, &a.Version.Major)
		a.Binary.OS, a.Binary.Architecture = "linux", arch
		a.Binary.Package.Name = "runtime.tar.gz"
		a.Binary.Package.Link = "https://github.com/adoptium/temurin" + ver + "-binaries/releases/download/tag/runtime.tar.gz"
		if mutate != nil {
			mutate(a)
		}
	})
}

func TestJavaLinuxInstallAllFamiliesAndReuse(t *testing.T) {
	for _, ver := range SupportedJavaMajors() {
		for _, arch := range []string{"amd64", "arm64"} {
			t.Run(ver+"/"+arch, func(t *testing.T) {
				archive := javaTestTar(t, tar.Header{Name: "jdk/bin/java", PAXRecords: map[string]string{"test.body": ver}})
				deps, calls := linuxJavaFixture(t, archive, ver, map[string]string{"amd64": "x64", "arm64": "aarch64"}[arch], nil)
				root := t.TempDir()
				for i := 0; i < 2; i++ {
					info, err := installTemurinJava(ver, "linux", arch, root, deps)
					if err != nil || info.Version != ver || deps.detect(info.Path) != ver {
						t.Fatalf("%+v %v", info, err)
					}
				}
				if *calls != 2 {
					t.Fatalf("repeat install downloaded: %d calls", *calls)
				}
			})
		}
	}
}

func TestJavaLinuxFailuresPreservePrevious(t *testing.T) {
	for _, scenario := range []string{"checksum", "wrong-major", "unparseable", "publish", "post-publish", "metadata-os", "metadata-arch", "unavailable", "lock"} {
		t.Run(scenario, func(t *testing.T) {
			body := "21"
			if scenario == "wrong-major" {
				body = "17"
			}
			if scenario == "unparseable" {
				body = ""
			}
			archive := javaTestTar(t, tar.Header{Name: "jdk/bin/java", PAXRecords: map[string]string{"test.body": body}})
			deps, calls := linuxJavaFixture(t, archive, "21", "x64", func(a *temurinAsset) {
				if scenario == "checksum" {
					a.Binary.Package.Checksum = strings.Repeat("0", 64)
				}
				if scenario == "metadata-os" {
					a.Binary.OS = "windows"
				}
				if scenario == "metadata-arch" {
					a.Binary.Architecture = "aarch64"
				}
			})
			root := t.TempDir()
			old := filepath.Join(root, "java-21-temurin", "bin", "java")
			if err := os.MkdirAll(filepath.Dir(old), 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(old, []byte("previous"), 0755); err != nil {
				t.Fatal(err)
			}
			if scenario == "lock" {
				if err := os.WriteFile(filepath.Join(root, ".install.lock"), nil, 0600); err != nil {
					t.Fatal(err)
				}
			}
			if scenario == "publish" {
				deps.rename = func(r *os.Root, a, b string) error {
					if filepath.Base(a) == "payload" {
						return fmt.Errorf("injected failure")
					}
					return r.Rename(a, b)
				}
			}
			if scenario == "post-publish" {
				detect := deps.detect
				deps.detect = func(bin string) string {
					if bin == old {
						return ""
					}
					return detect(bin)
				}
			}
			if scenario == "unavailable" {
				deps.client.Transport = javaRoundTrip(func(r *http.Request) (*http.Response, error) {
					return &http.Response{StatusCode: 404, Body: http.NoBody, Request: r}, nil
				})
			}
			if _, err := installTemurinJava("21", "linux", "amd64", root, deps); err == nil {
				t.Fatal("accepted failed install")
			}
			got, err := os.ReadFile(old)
			if err != nil || string(got) != "previous" {
				t.Fatalf("lost previous: %q %v", got, err)
			}
			if scenario == "lock" && *calls != 0 {
				t.Fatal("download despite lock")
			}
		})
	}
}

func TestJavaLinuxTarRejectsUnsafeArchives(t *testing.T) {
	cases := map[string][]tar.Header{}
	for _, name := range []string{"/jdk/bin/java", "../bin/java", "jdk/../../escape", "jdk/../escape", "jdk//java", `jdk\bin\java`, "jdk/C:/evil", "java"} {
		cases[name] = []tar.Header{{Name: name}}
	}
	cases["symlink"] = []tar.Header{{Name: "jdk/bin/java", Typeflag: tar.TypeSymlink, Linkname: "../../escape"}}
	cases["hardlink"] = []tar.Header{{Name: "jdk/bin/java", Typeflag: tar.TypeLink, Linkname: "jdk/lib/real"}}
	cases["device"] = []tar.Header{{Name: "jdk/bin/java", Typeflag: tar.TypeChar}}
	cases["duplicate"] = []tar.Header{{Name: "jdk/bin/java"}, {Name: "jdk/bin/java"}}
	cases["roots"] = []tar.Header{{Name: "jdk/bin/java"}, {Name: "other/lib/modules"}}
	for name, members := range cases {
		t.Run(name, func(t *testing.T) {
			archive := filepath.Join(t.TempDir(), "runtime.tar.gz")
			if err := os.WriteFile(archive, javaTestTar(t, members...), 0600); err != nil {
				t.Fatal(err)
			}
			if err := extractJavaTarGz(context.Background(), archive, t.TempDir()); err == nil {
				t.Fatal("accepted unsafe archive")
			}
		})
	}
}

func TestJavaLinuxDiscoveryDoesNotTrustDirectoryName(t *testing.T) {
	for _, actual := range []string{"17", "", "21"} {
		root := t.TempDir()
		bin := filepath.Join(root, "java-21-temurin", "bin", "java")
		if err := os.MkdirAll(filepath.Dir(bin), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(bin, []byte(actual), 0755); err != nil {
			t.Fatal(err)
		}
		got := scanLinuxJavaRoots("21", []string{root}, func(path string) string { data, _ := os.ReadFile(path); return string(data) })
		if (got == bin) != (actual == "21") {
			t.Fatalf("major %q resolved to %q", actual, got)
		}
	}
}

func TestJava16HistoricalMetadataIsBoundedAndUnavailableExplicit(t *testing.T) {
	calls := 0
	deps := defaultWindowsJavaInstaller()
	deps.client.Transport = javaRoundTrip(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.URL.Path != "/v3/assets/feature_releases/16/ga" || r.URL.Query().Get("page_size") != "10" || r.URL.Query().Get("page") != "0" {
			t.Fatalf("unbounded/non-historical endpoint: %s", r.URL)
		}
		return &http.Response{StatusCode: 404, Body: http.NoBody, Request: r}, nil
	})
	_, err := fetchTemurinPackageForOS(context.Background(), deps.client, "16", "aarch64", "windows")
	if err == nil || !strings.Contains(err.Error(), "Java 16 windows/aarch64") || calls != 2 {
		t.Fatalf("%v (%d calls)", err, calls)
	}
}

func TestJavaLinuxTarMaterializesConfinedFileLinks(t *testing.T) {
	archive := javaTestTar(t,
		tar.Header{Name: "jdk/lib/server/libjsig.so", Typeflag: tar.TypeSymlink, Linkname: "../libjsig.so"},
		tar.Header{Name: "jdk/lib/hard", Typeflag: tar.TypeLink, Linkname: "jdk/lib/libjsig.so"},
		tar.Header{Name: "jdk/lib/libjsig.so", PAXRecords: map[string]string{"test.body": "library"}})
	file := filepath.Join(t.TempDir(), "runtime.tar.gz")
	if err := os.WriteFile(file, archive, 0600); err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	if err := extractJavaTarGz(context.Background(), file, root); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"lib/server/libjsig.so", "lib/hard"} {
		st, err := os.Lstat(filepath.Join(root, name))
		if err != nil || !st.Mode().IsRegular() {
			t.Fatalf("not materialized: %v %v", st, err)
		}
		data, err := os.ReadFile(filepath.Join(root, name))
		if err != nil || string(data) != "library" {
			t.Fatalf("%q %v", data, err)
		}
	}
}

func TestJavaLinuxTarLimitsCRCAndCancellation(t *testing.T) {
	for _, scenario := range []string{"entries", "size", "crc", "cancel", "cycle", "directory-link"} {
		t.Run(scenario, func(t *testing.T) {
			var archive []byte
			if scenario == "size" {
				var b bytes.Buffer
				gz := gzip.NewWriter(&b)
				tw := tar.NewWriter(gz)
				if err := tw.WriteHeader(&tar.Header{Name: "jdk/lib/modules", Size: javaDownloadLimit + 1, Mode: 0644}); err != nil {
					t.Fatal(err)
				}
				// Deliberately incomplete body: extraction must reject the header first.
				_ = gz.Close()
				archive = b.Bytes()
			} else {
				members := []tar.Header{{Name: "jdk/bin/java"}}
				if scenario == "entries" {
					members = nil
					for i := 0; i <= javaEntryLimit; i++ {
						members = append(members, tar.Header{Name: fmt.Sprintf("jdk/f%d", i)})
					}
				}
				if scenario == "cycle" {
					members = []tar.Header{{Name: "jdk/a", Typeflag: tar.TypeSymlink, Linkname: "b"}, {Name: "jdk/b", Typeflag: tar.TypeSymlink, Linkname: "a"}}
				}
				if scenario == "directory-link" {
					members = []tar.Header{{Name: "jdk/lib/", Typeflag: tar.TypeDir}, {Name: "jdk/linked", Typeflag: tar.TypeSymlink, Linkname: "lib"}}
				}
				archive = javaTestTar(t, members...)
				if scenario == "crc" {
					archive[len(archive)-8] ^= 0xff
				}
			}
			file := filepath.Join(t.TempDir(), "runtime.tar.gz")
			if err := os.WriteFile(file, archive, 0600); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if scenario == "cancel" {
				cancel()
			}
			if err := extractJavaTarGz(ctx, file, t.TempDir()); err == nil {
				t.Fatal("accepted invalid/unbounded archive")
			}
		})
	}
}
