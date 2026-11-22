package api

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Pekly/NeoPanel/internal/downloader"
	"github.com/Pekly/NeoPanel/internal/rooted"
	"github.com/Pekly/NeoPanel/internal/runner"
	"github.com/Pekly/NeoPanel/internal/serverjar"
)

// All downloads in these tests are local synthetic bytes. Nothing invokes a
// server, Java, a provider installer, or a remote transport.
func softwareArtifactJar(t *testing.T) []byte {
	t.Helper()
	var b bytes.Buffer
	z := zip.NewWriter(&b)
	for _, entry := range []struct{ name, data string }{
		{"META-INF/MANIFEST.MF", "Manifest-Version: 1.0\nMain-Class: net.minecraft.server.Main\n\n"},
		{"version.json", `{"id":"1.21.4","java_version":21}`},
	} {
		w, err := z.Create(entry.name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = w.Write([]byte(entry.data)); err != nil {
			t.Fatal(err)
		}
	}
	if err := z.Close(); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

func softwareArtifactDownload(t *testing.T) func(context.Context, string, string, string, string) (serverjar.Info, error) {
	t.Helper()
	data := softwareArtifactJar(t)
	return func(ctx context.Context, flavor, version, dir, filename string) (serverjar.Info, error) {
		if err := ctx.Err(); err != nil {
			return serverjar.Info{}, err
		}
		if err := os.WriteFile(filepath.Join(dir, filename), data, 0644); err != nil {
			return serverjar.Info{}, err
		}
		st, err := os.Stat(filepath.Join(dir, filename))
		if err != nil {
			return serverjar.Info{}, err
		}
		sum := sha256.Sum256(data)
		// Deliberately different from configured "latest", and explicit provider
		// metadata: record must bind it to the actual installed synthetic bytes.
		return serverjar.Info{Status: "identified", Source: "mcjars", Software: flavor, MinecraftVersion: "1.21.4", ProjectVersion: "test-release", RequiredJava: 21, JavaSource: "mcjars", Filename: filename, SHA256: hex.EncodeToString(sum[:]), Size: st.Size(), Modified: st.ModTime()}, nil
	}
}

func TestSoftwareArtifactChangeRequiresStoppedGateForEveryEngine(t *testing.T) {
	for _, flavor := range []string{"paper", "vanilla", "fabric", "forge", "neoforge", "velocity", "bungeecord", "steel"} {
		t.Run(flavor, func(t *testing.T) {
			f := newReliabilityFixture(t)
			called := false
			f.router.software.beginChange = func(context.Context, string) (context.Context, func(), error) {
				called = true
				return nil, nil, runner.ErrProxyConfigRunning
			}
			f.router.software.download = func(context.Context, string, string, string, string) (serverjar.Info, error) {
				t.Fatal("download before stopped-process gate")
				return serverjar.Info{}, nil
			}
			w := f.request("POST", "/api/servers/alpha/software", fmt.Sprintf(`{"software":%q,"version":"latest"}`, flavor), "owner")
			if !called || w.Code != http.StatusConflict || !strings.Contains(w.Body.String(), "Stop the server") {
				t.Fatal(called, w.Code, w.Body.String())
			}
			srv, _ := f.db.GetServer("alpha")
			if srv.Type != "paper" || srv.JavaPath != "java-custom" {
				t.Fatalf("modified config: %+v", srv)
			}
		})
	}
}

func TestSoftwareArtifactChangeRetainsExclusiveLeaseAndOperation(t *testing.T) {
	f := newReliabilityFixture(t)
	dir := f.manager.GetWorkingDir("alpha")
	download := softwareArtifactDownload(t)
	f.router.software.download = func(ctx context.Context, flavor, version, dest, filename string) (serverjar.Info, error) {
		if release, err := rooted.TryMutation(dir); !errors.Is(err, rooted.ErrMutationBusy) {
			if release != nil {
				release()
			}
			t.Fatal("missing exclusive lease", err)
		}
		if release, err := f.manager.BeginServerOperation("alpha"); !errors.Is(err, runner.ErrServerOperation) {
			if release != nil {
				release()
			}
			t.Fatal("missing operation gate", err)
		}
		return download(ctx, flavor, version, dest, filename)
	}
	w := f.request("POST", "/api/servers/alpha/software", `{"software":"vanilla","version":"latest"}`, "owner")
	if w.Code != http.StatusOK {
		t.Fatal(w.Code, w.Body.String())
	}
	release, err := rooted.TryMutation(dir)
	if err != nil {
		t.Fatal("leaked lease", err)
	}
	release()
}

func TestSoftwareArtifactChangeRejectsBusyFilesAndShutdown(t *testing.T) {
	for _, shutdown := range []bool{false, true} {
		t.Run(fmt.Sprint(shutdown), func(t *testing.T) {
			f := newReliabilityFixture(t)
			if shutdown {
				f.manager.ShutdownAll()
			} else {
				release, err := rooted.TryMutation(f.manager.GetWorkingDir("alpha"))
				if err != nil {
					t.Fatal(err)
				}
				defer release()
			}
			f.router.software.download = func(context.Context, string, string, string, string) (serverjar.Info, error) {
				t.Fatal("download despite busy/shutdown")
				return serverjar.Info{}, nil
			}
			w := f.request("POST", "/api/servers/alpha/software", `{"software":"vanilla","version":"latest"}`, "owner")
			if w.Code != http.StatusConflict {
				t.Fatal(w.Code, w.Body.String())
			}
		})
	}
}

func TestSoftwareArtifactFailedReinstallPreservesExistingFiles(t *testing.T) {
	for _, flavor := range []string{"vanilla", "forge", "steel"} {
		t.Run(flavor, func(t *testing.T) {
			f := newReliabilityFixture(t)
			dir := f.manager.GetWorkingDir("alpha")
			old := map[string]string{"server.jar": "old jar", "world/level.dat": "old world", "config/config.toml": "old config", "libraries/old.jar": "old library"}
			for name, data := range old {
				p := filepath.Join(dir, filepath.FromSlash(name))
				if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(p, []byte(data), 0644); err != nil {
					t.Fatal(err)
				}
			}
			f.router.software.download = func(_ context.Context, _, _, dest, filename string) (serverjar.Info, error) {
				if dest == dir || filepath.Dir(dest) != dir {
					t.Fatal("download not privately staged", dest)
				}
				if err := os.WriteFile(filepath.Join(dest, filename), []byte("partial"), 0644); err != nil {
					t.Fatal(err)
				}
				return serverjar.Info{}, errors.New("secret upstream URL and C:\\host\\private")
			}
			w := f.request("POST", "/api/servers/alpha/reinstall", fmt.Sprintf(`{"software":%q,"version":"latest","wipe_files":true}`, flavor), "owner")
			if w.Code != http.StatusBadGateway || strings.Contains(w.Body.String(), "secret") || strings.Contains(w.Body.String(), "host") {
				t.Fatal(w.Code, w.Body.String())
			}
			for name, want := range old {
				got, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(name)))
				if err != nil || string(got) != want {
					t.Fatalf("lost %s: %s %v", name, got, err)
				}
			}
			entries, err := os.ReadDir(dir)
			if err != nil {
				t.Fatal(err)
			}
			for _, entry := range entries {
				if strings.HasPrefix(entry.Name(), ".neopanel-") {
					t.Fatal("leaked staging directory", entry.Name())
				}
			}
			srv, _ := f.db.GetServer("alpha")
			if srv.Type != "paper" || srv.Version != "1.21" || srv.JavaPath != "java-custom" {
				t.Fatalf("failed download changed config: %+v", srv)
			}
		})
	}
}

func TestSoftwareArtifactVersionsAndManualJavaPreserved(t *testing.T) {
	for _, route := range []string{"software", "reinstall"} {
		for _, java := range []string{"java-custom", "", "auto"} {
			t.Run(route+"/"+java, func(t *testing.T) {
				f := newReliabilityFixture(t)
				srv, _ := f.db.GetServer("alpha")
				srv.JavaPath = java
				if err := f.db.UpdateServer(srv); err != nil {
					t.Fatal(err)
				}
				f.router.software.download = softwareArtifactDownload(t)
				w := f.request("POST", "/api/servers/alpha/"+route, `{"software":"vanilla","version":"latest"}`, "owner")
				if w.Code != http.StatusOK {
					t.Fatal(w.Code, w.Body.String())
				}
				srv, _ = f.db.GetServer("alpha")
				if srv.JavaPath != java || srv.JavaArgs != "-Dkeep=true" || srv.Version != "latest" || srv.Type != "vanilla" {
					t.Fatalf("configured values overwritten: %+v", srv)
				}
				var response struct {
					Version  string          `json:"version"`
					JavaPath string          `json:"java_path"`
					Artifact *serverjar.Info `json:"artifact"`
				}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatal(err)
				}
				if response.Version != "latest" || response.JavaPath != java || response.Artifact == nil || response.Artifact.MinecraftVersion != "1.21.4" || response.Artifact.RequiredJava != 21 || response.Artifact.Source != "mcjars" {
					t.Fatalf("bad artifact response: %s", w.Body.String())
				}
				local, err := serverjar.Inspect(context.Background(), f.manager.GetWorkingDir("alpha"), srv.JarFile)
				if err != nil || response.Artifact.SHA256 != local.SHA256 {
					t.Fatalf("metadata not bound to current bytes: %+v %v", local, err)
				}
				stored, err := f.db.GetPanelSetting("server_artifact_v1:alpha")
				if err != nil || !strings.Contains(stored, local.SHA256) {
					t.Fatal("artifact not persisted", stored, err)
				}
			})
		}
	}
}

func TestSoftwareArtifactNativeToJavaSelectsAuto(t *testing.T) {
	for _, route := range []string{"software", "reinstall"} {
		t.Run(route, func(t *testing.T) {
			f := newReliabilityFixture(t)
			srv, _ := f.db.GetServer("alpha")
			srv.Type, srv.JavaPath = "steel", ""
			if err := f.db.UpdateServer(srv); err != nil {
				t.Fatal(err)
			}
			f.router.software.download = softwareArtifactDownload(t)
			w := f.request("POST", "/api/servers/alpha/"+route, `{"software":"vanilla","version":"latest"}`, "owner")
			if w.Code != http.StatusOK {
				t.Fatal(w.Code, w.Body.String())
			}
			srv, _ = f.db.GetServer("alpha")
			if srv.JavaPath != "auto" {
				t.Fatalf("native runtime carried over: %+v", srv)
			}
		})
	}
}

func TestSoftwareArtifactMetadataRecordFailureIsNotSuccess(t *testing.T) {
	f := newReliabilityFixture(t)
	download := softwareArtifactDownload(t)
	f.router.software.download = func(ctx context.Context, flavor, version, dir, filename string) (serverjar.Info, error) {
		info, err := download(ctx, flavor, version, dir, filename)
		info.SHA256 = strings.Repeat("0", 64)
		return info, err
	}
	w := f.request("POST", "/api/servers/alpha/software", `{"software":"vanilla","version":"latest"}`, "owner")
	if w.Code != http.StatusInternalServerError || !strings.Contains(w.Body.String(), "metadata") || strings.Contains(w.Body.String(), f.manager.GetWorkingDir("alpha")) {
		t.Fatal(w.Code, w.Body.String())
	}
	srv, _ := f.db.GetServer("alpha")
	if srv.JavaPath != "java-custom" || f.manager.ArtifactSummary(srv) != nil {
		t.Fatal("bad metadata adopted")
	}
}

func TestSoftwareArtifactPreparedLibrariesPublished(t *testing.T) {
	f := newReliabilityFixture(t)
	dir := f.manager.GetWorkingDir("alpha")
	if err := os.WriteFile(filepath.Join(dir, "obsolete.jar"), []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}
	download := softwareArtifactDownload(t)
	f.router.software.download = func(ctx context.Context, flavor, version, dest, filename string) (serverjar.Info, error) {
		if err := os.MkdirAll(filepath.Join(dest, "libraries", "test"), 0755); err != nil {
			return serverjar.Info{}, err
		}
		if err := os.WriteFile(filepath.Join(dest, "libraries", "test", "dependency.jar"), []byte("synthetic library"), 0644); err != nil {
			return serverjar.Info{}, err
		}
		if err := os.WriteFile(filepath.Join(dest, "eula.txt"), []byte("eula=true\n"), 0644); err != nil {
			return serverjar.Info{}, err
		}
		return download(ctx, flavor, version, dest, filename)
	}
	w := f.request("POST", "/api/servers/alpha/reinstall", `{"software":"forge","version":"latest"}`, "owner")
	if w.Code != http.StatusOK {
		t.Fatal(w.Code, w.Body.String())
	}
	if data, err := os.ReadFile(filepath.Join(dir, "libraries", "test", "dependency.jar")); err != nil || string(data) != "synthetic library" {
		t.Fatal(string(data), err)
	}
	if _, err := os.Stat(filepath.Join(dir, "obsolete.jar")); !os.IsNotExist(err) {
		t.Fatal("old runtime not cleaned", err)
	}
}

func TestSoftwareArtifactBadStagePreservesTree(t *testing.T) {
	for _, kind := range []string{"hash", "script", "symlink"} {
		t.Run(kind, func(t *testing.T) {
			f := newReliabilityFixture(t)
			dir := f.manager.GetWorkingDir("alpha")
			if err := os.WriteFile(filepath.Join(dir, "keep.txt"), []byte("keep"), 0644); err != nil {
				t.Fatal(err)
			}
			download := softwareArtifactDownload(t)
			f.router.software.download = func(ctx context.Context, flavor, version, dest, filename string) (serverjar.Info, error) {
				info, err := download(ctx, flavor, version, dest, filename)
				if err != nil {
					return info, err
				}
				switch kind {
				case "hash":
					info.SHA256 = strings.Repeat("0", 64)
				case "script":
					err = os.WriteFile(filepath.Join(dest, "install.sh"), []byte("do not run"), 0755)
				case "symlink":
					err = os.Symlink(dir, filepath.Join(dest, "libraries"))
					if err != nil {
						t.Skipf("symlinks unavailable: %v", err)
					}
				}
				return info, err
			}
			w := f.request("POST", "/api/servers/alpha/reinstall", `{"software":"forge","version":"latest"}`, "owner")
			if w.Code != http.StatusBadGateway {
				t.Fatal(w.Code, w.Body.String())
			}
			if data, err := os.ReadFile(filepath.Join(dir, "keep.txt")); err != nil || string(data) != "keep" {
				t.Fatal("validation happened after cleanup", string(data), err)
			}
		})
	}
}

func TestSoftwareArtifactNativeReinstallPreservesConfigAndPort(t *testing.T) {
	f := newReliabilityFixture(t)
	srv, _ := f.db.GetServer("alpha")
	srv.Type, srv.Port = "steel", 25589
	if err := f.db.UpdateServer(srv); err != nil {
		t.Fatal(err)
	}
	dir := f.manager.GetWorkingDir("alpha")
	if err := os.MkdirAll(filepath.Join(dir, "config"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config", "config.toml"), []byte("[server]\nserver_port = 25565\nmotd = \"Keep me\"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	f.router.software.download = func(_ context.Context, _, _, dest, filename string) (serverjar.Info, error) {
		if filename != downloader.ServerFileName("steel") {
			t.Fatal(filename)
		}
		if err := os.WriteFile(filepath.Join(dest, filename), []byte("synthetic native executable - never run"), 0755); err != nil {
			return serverjar.Info{}, err
		}
		if err := os.MkdirAll(filepath.Join(dest, "config"), 0755); err != nil {
			return serverjar.Info{}, err
		}
		if err := os.WriteFile(filepath.Join(dest, "config", "config.toml"), []byte("[server]\nserver_port = 25565\nmotd = \"New release\"\n"), 0644); err != nil {
			return serverjar.Info{}, err
		}
		return serverjar.Info{Status: "native", Source: "official", Software: "steel", Filename: filename}, nil
	}
	w := f.request("POST", "/api/servers/alpha/reinstall", `{"software":"steel","version":"latest","preserve_config":true}`, "owner")
	if w.Code != http.StatusOK {
		t.Fatal(w.Code, w.Body.String())
	}
	data, err := os.ReadFile(filepath.Join(dir, "config", "config.toml"))
	if err != nil || !strings.Contains(string(data), "Keep me") || !strings.Contains(string(data), "25589") {
		t.Fatal(string(data), err)
	}
	srv, _ = f.db.GetServer("alpha")
	if srv.JavaPath != "" || srv.JavaArgs != "" || srv.MemoryMinMB != 0 {
		t.Fatal("native retained Java settings", srv)
	}
}
