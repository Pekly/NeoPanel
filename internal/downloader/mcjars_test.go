package downloader

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Pekly/NeoPanel/internal/rooted"
)

const testMCJarsID = "e60f6fa0-ae4d-480b-b4dd-56af5afd48bf"
const testMCJarsURL = "https://files.mcjars.app/server.jar"

func testMCJarsPayload(kind, version, java, location string, size int) string {
	mc, project := fmt.Sprintf("%q", version), "null"
	if mcjarsProxy(kind) {
		mc, project = "null", fmt.Sprintf("%q", version)
	}
	return fmt.Sprintf(`{"build":{"uuid":%q,"type":%q,"version_id":%s,"project_version_id":%s,"name":"#232","java":%s,"installation":[[{"type":"download","url":%q,"file":"server.jar","size":%d}]]}}`, testMCJarsID, kind, mc, project, java, location, size)
}
func testMCJarsPrepared(size int) string {
	var p mcjarsResponse
	json.Unmarshal([]byte(testMCJarsPayload("FORGE", "1.20.1", "17", testMCJarsURL, size)), &p)
	p.Build.Project = new(string)
	*p.Build.Project = "47.4.29"
	p.Build.Installation = [][]mcjarsStep{{{Type: "download", URL: testMCJarsURL, File: "mcvapi.server.jar.zip", Size: int64(size)}, {Type: "remove", Location: "libraries"}}, {{Type: "unzip", File: "mcvapi.server.jar.zip", Location: "."}}, {{Type: "remove", Location: "mcvapi.server.jar.zip"}}}
	data, _ := json.Marshal(p)
	return string(data)
}
func testMCJarsZIP(t *testing.T, files map[string]string) string {
	t.Helper()
	var b bytes.Buffer
	w := zip.NewWriter(&b)
	for n, data := range files {
		f, err := w.Create(n)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = f.Write([]byte(data)); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return b.String()
}
func testMCJarsJar(t *testing.T) string {
	return testMCJarsZIP(t, map[string]string{"META-INF/MANIFEST.MF": "Manifest-Version: 1.0\r\nMain-Class: Test\r\n\r\n", "Test.class": "fixture"})
}
func mockMCJarsInstall(t *testing.T, kind, version, payload, artifact string) *[]string {
	return mockLatestMetadata(t, map[string]string{mcjarsAPI + "types/" + kind + "/versions/" + version + "/latest": payload, testMCJarsURL: artifact})
}
func TestMCJarsMappingsAndInstalledInfo(t *testing.T) {
	for _, flavor := range []string{"paper", "purpur", "leaf", "leafmc", "folia", "fabric", "forge", "neoforge", "vanilla", "velocity", "waterfall", "bungeecord"} {
		t.Run(flavor, func(t *testing.T) {
			kind, _ := mcjarsType(flavor)
			version := "1.21.4"
			if mcjarsProxy(kind) {
				version = "4.3.2-SNAPSHOT"
			}
			jar := testMCJarsJar(t)
			calls := mockMCJarsInstall(t, kind, "latest", testMCJarsPayload(kind, version, "21", testMCJarsURL, len(jar)), jar)
			dir := t.TempDir()
			info, err := DownloadServerJarWithInfo(context.Background(), flavor, "latest", dir, "custom.jar")
			if err != nil {
				t.Fatal(err)
			}
			sum := sha256.Sum256([]byte(jar))
			if info.SHA256 != hex.EncodeToString(sum[:]) || info.Size != int64(len(jar)) || info.Modified.IsZero() || info.Filename != "custom.jar" || info.Source != "mcjars" || info.JavaSource != "mcjars-build" {
				t.Fatalf("bad info: %+v", info)
			}
			data, err := os.ReadFile(filepath.Join(dir, "custom.jar"))
			if err != nil || string(data) != jar {
				t.Fatal("wrong bytes", err)
			}
			if len(*calls) != 2 {
				t.Fatal(*calls)
			}
			_, err = os.Stat(filepath.Join(dir, "eula.txt"))
			if mcjarsProxy(kind) {
				if !os.IsNotExist(err) || info.MinecraftVersion != "" || info.ProjectVersion != version {
					t.Fatalf("proxy conflation: %+v, %v", info, err)
				}
			} else if err != nil || info.MinecraftVersion != version {
				t.Fatal(info, err)
			}
		})
	}
}
func TestMCJarsJavaDetailAndIdentity(t *testing.T) {
	initial := testMCJarsPayload("PAPER", "1.21.4", "null", testMCJarsURL, 100)
	for _, tc := range []struct {
		name, detail string
		java         int
		source       string
		fail         bool
	}{
		{"version", strings.TrimSuffix(initial, "}") + `,"version":{"id":"1.21.4","java":21}}`, 21, "mcjars-version", false},
		{"build-override", strings.TrimSuffix(strings.Replace(initial, `"java":null`, `"java":17`, 1), "}") + `,"version":{"id":"1.21.4","java":21}}`, 17, "mcjars-build", false},
		{"unknown", initial, 0, "", false},
		{"wrong-version", strings.TrimSuffix(initial, "}") + `,"version":{"id":"1.20.1","java":17}}`, 0, "", true},
		{"wrong-id", strings.Replace(initial, testMCJarsID, "11111111-1111-1111-1111-111111111111", 1), 0, "", true},
		{"wrong-engine", strings.Replace(initial, "PAPER", "PURPUR", 1), 0, "", true},
		{"invalid-java", strings.Replace(initial, `"java":null`, `"java":999`, 1), 0, "", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := mockLatestMetadata(t, map[string]string{mcjarsAPI + "types/PAPER/versions/1.21.4/latest": initial, mcjarsAPI + testMCJarsID: tc.detail})
			b, err := ResolveMCJarsBuild(context.Background(), "paper", "1.21.4")
			if tc.fail {
				if err == nil {
					t.Fatal("accepted invalid detail")
				}
				return
			}
			if err != nil || b.Info.RequiredJava != tc.java || b.Info.JavaSource != tc.source {
				t.Fatal(b, err)
			}
			if len(*calls) != 2 {
				t.Fatal(*calls)
			}
		})
	}
}
func TestMCJarsExplicitProxyAndHashLookup(t *testing.T) {
	for _, version := range []string{"1.21.4", "4.3.2-SNAPSHOT"} {
		t.Run(version, func(t *testing.T) {
			calls := mockLatestMetadata(t, map[string]string{mcjarsAPI + "types/VELOCITY/versions/" + version + "/latest": testMCJarsPayload("VELOCITY", "4.3.2-SNAPSHOT", "25", testMCJarsURL, 100)})
			b, err := ResolveMCJarsBuild(context.Background(), "velocity", version)
			if version == "1.21.4" {
				if err == nil {
					t.Fatal("reinterpreted explicit version")
				}
			} else if err != nil || b.Info.MinecraftVersion != "" || b.Info.RequiredJava != 25 {
				t.Fatal(b, err)
			}
			if len(*calls) != 1 {
				t.Fatal(*calls)
			}
		})
	}
	hash := strings.Repeat("a", 64)
	payload := testMCJarsPayload("PAPER", "1.16.5", "11", testMCJarsURL, 100)
	payload = strings.TrimSuffix(payload, "}") + `,"latest":{"uuid":"not-the-matched-build","type":"FORGE"}}`
	calls := mockLatestMetadata(t, map[string]string{mcjarsAPI + hash: payload})
	b, err := LookupMCJarsBuild(context.Background(), strings.ToUpper(hash))
	if err != nil || b.Info.SHA256 != hash || b.Info.BuildID != testMCJarsID || b.Info.MinecraftVersion != "1.16.5" || len(*calls) != 1 {
		t.Fatal(b, err, *calls)
	}
	if _, err = LookupMCJarsBuild(context.Background(), "../bad"); err == nil || len(*calls) != 1 {
		t.Fatal("bad hash requested")
	}
}
func TestMCJarsMalformedAndLimits(t *testing.T) {
	good := testMCJarsPayload("PAPER", "1.21.4", "21", testMCJarsURL, 100)
	for _, body := range []string{"null", "{}", `{"build":null}`, "not json", good + "{}", strings.Repeat(" ", (4<<20)+1), strings.Replace(good, "PAPER", "PURPUR", 1), strings.Replace(good, "1.21.4", "1.20.1", 1), strings.Replace(good, `"size":100`, `"size":0`, 1), strings.Replace(good, `"size":100`, `"size":1073741825`, 1), strings.Replace(good, `"java":21`, `"java":"21"`, 1), strings.Replace(good, `"type":"download"`, `"type":"execute"`, 1)} {
		t.Run(fmt.Sprintf("%d-%s", len(body), body[:min(20, len(body))]), func(t *testing.T) {
			calls := mockLatestMetadata(t, map[string]string{mcjarsAPI + "types/PAPER/versions/1.21.4/latest": body})
			if _, err := ResolveMCJarsBuild(context.Background(), "paper", "1.21.4"); err == nil {
				t.Fatal("accepted malformed response")
			}
			if len(*calls) > 2 {
				t.Fatal(*calls)
			}
		})
	}
	for _, version := range []string{"", "../x", "1.21?x", "1.é", "latest\n"} {
		calls := mockLatestMetadata(t, nil)
		if _, err := ResolveMCJarsBuild(context.Background(), "paper", version); err == nil || len(*calls) != 0 {
			t.Fatal("invalid identifier accepted")
		}
	}
}
func TestMCJarsHTTPPolicy(t *testing.T) {
	for _, status := range []int{404, 503, 302} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			old := httpClient
			t.Cleanup(func() { httpClient = old })
			calls := 0
			httpClient = &http.Client{Transport: modrinthTransport(func(r *http.Request) (*http.Response, error) {
				calls++
				if r.URL.Host != "mcjars.app" || r.URL.RawQuery != "" {
					t.Error("origin or query changed")
				}
				deadline, ok := r.Context().Deadline()
				if !ok || time.Until(deadline) > 30*time.Second {
					t.Error("missing resolver deadline")
				}
				return &http.Response{StatusCode: status, Header: http.Header{"Location": []string{"https://files.mcjars.app/redirect"}}, Body: io.NopCloser(strings.NewReader(""))}, nil
			})}
			_, err := LookupMCJarsBuild(context.Background(), strings.Repeat("a", 64))
			if err == nil || calls != 1 {
				t.Fatal(err, calls)
			}
			if status == 404 && !errors.Is(err, ErrMCJarsNotFound) {
				t.Fatal(err)
			}
		})
	}
	old := httpClient
	t.Cleanup(func() { httpClient = old })
	httpClient = &http.Client{Transport: modrinthTransport(func(r *http.Request) (*http.Response, error) { <-r.Context().Done(); return nil, r.Context().Err() })}
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	if _, err := ResolveMCJarsBuild(ctx, "paper", "latest"); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal(err)
	}
}
func TestMCJarsURLPolicy(t *testing.T) {
	for _, u := range []string{"http://files.mcjars.app/a", "https://user@files.mcjars.app/a", "https://127.0.0.1/a", "https://[::1]/a", "https://10.0.0.1/a", "https://localhost/a", "https://files.mcjars.app.evil/a", "https://files.mcjars.app:444/a", "https://evil.invalid/a"} {
		if _, err := mcjarsDownloadURL(u); err == nil {
			t.Fatal("accepted", u)
		}
		req, _ := http.NewRequest("GET", u, nil)
		if err := mcjarsClient(true).CheckRedirect(req, []*http.Request{req}); err == nil {
			t.Fatal("accepted redirect", u)
		}
	}
	for _, ip := range []string{"127.0.0.1", "::1", "10.0.0.1", "192.168.1.1", "169.254.169.254", "100.64.0.1", "fc00::1", "0.1.2.3", "198.18.0.1"} {
		if mcjarsPublicIP(net.ParseIP(ip)) {
			t.Fatal("accepted private IP", ip)
		}
	}
	if !mcjarsPublicIP(net.ParseIP("1.1.1.1")) {
		t.Fatal("rejected public IP")
	}
}
func TestMCJarsPreparedPublication(t *testing.T) {
	jar := testMCJarsJar(t)
	bundle := testMCJarsZIP(t, map[string]string{"server.jar": jar, "libraries/a/b.jar": "library"})
	mockMCJarsInstall(t, "FORGE", "1.20.1", testMCJarsPrepared(len(bundle)), bundle)
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "libraries"), 0755)
	os.WriteFile(filepath.Join(dir, "libraries", "old"), []byte("old"), 0644)
	os.WriteFile(filepath.Join(dir, "world"), []byte("preserve"), 0644)
	os.WriteFile(filepath.Join(dir, "server.properties"), []byte("preserve"), 0644)
	info, err := DownloadServerJarWithInfo(context.Background(), "forge", "1.20.1", dir, "forge.jar")
	if err != nil {
		t.Fatal(err)
	}
	if info.Size != int64(len(jar)) || info.ProjectVersion != "47.4.29" {
		t.Fatal(info)
	}
	for name, want := range map[string]string{"forge.jar": jar, "libraries/a/b.jar": "library", "world": "preserve", "server.properties": "preserve"} {
		got, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil || string(got) != want {
			t.Fatal(name, err)
		}
	}
	for _, name := range []string{"libraries/old", "server.jar", "mcvapi.server.jar.zip", "payload"} {
		if _, err = os.Stat(filepath.Join(dir, name)); !os.IsNotExist(err) {
			t.Fatal("unexpected published file", name)
		}
	}
}
func TestMCJarsStagedFailuresPreserveFiles(t *testing.T) {
	jar := testMCJarsJar(t)
	for _, tc := range []struct {
		name, body string
		size       int
		prepared   bool
	}{
		{"not-jar", "not jar", 7, false}, {"short", jar, len(jar) + 1, false}, {"long", jar, len(jar) - 1, false},
		{"traversal", testMCJarsZIP(t, map[string]string{"../world": "bad", "server.jar": jar}), 0, true},
		{"script", testMCJarsZIP(t, map[string]string{"run.sh": "bad", "server.jar": jar}), 0, true},
		{"missing-jar", testMCJarsZIP(t, map[string]string{"libraries/a": "bad"}), 0, true},
		{"path-conflict", testMCJarsZIP(t, map[string]string{"libraries/a": "bad", "libraries/a/b": "bad", "server.jar": jar}), 0, true},
		{"case-conflict", testMCJarsZIP(t, map[string]string{"libraries/A": "bad", "libraries/a": "bad", "server.jar": jar}), 0, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			size := tc.size
			if size == 0 {
				size = len(tc.body)
			}
			kind, version := "PAPER", "1.21.4"
			payload := testMCJarsPayload(kind, version, "21", testMCJarsURL, size)
			if tc.prepared {
				kind, version = "FORGE", "1.20.1"
				payload = testMCJarsPrepared(size)
			}
			mockMCJarsInstall(t, kind, version, payload, tc.body)
			dir := t.TempDir()
			for _, n := range []string{"server.jar", "world", "eula.txt"} {
				os.WriteFile(filepath.Join(dir, n), []byte("old"), 0644)
			}
			if err := DownloadServerJar(strings.ToLower(kind), version, dir, "server.jar"); err == nil {
				t.Fatal("accepted broken artifact")
			}
			for _, n := range []string{"server.jar", "world", "eula.txt"} {
				b, _ := os.ReadFile(filepath.Join(dir, n))
				if string(b) != "old" {
					t.Fatal("modified", n)
				}
			}
			entries, _ := os.ReadDir(dir)
			if len(entries) != 3 {
				t.Fatal("leaked staging", entries)
			}
		})
	}
}
func TestMCJarsRejectArbitraryRemovalBeforeDownload(t *testing.T) {
	payload := strings.Replace(testMCJarsPrepared(100), `"location":"libraries"`, `"location":"world"`, 1)
	calls := mockMCJarsInstall(t, "FORGE", "1.20.1", payload, "")
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "world"), []byte("old"), 0644)
	if err := DownloadServerJar("forge", "1.20.1", dir, "server.jar"); err == nil || len(*calls) != 1 {
		t.Fatal(err, *calls)
	}
	b, _ := os.ReadFile(filepath.Join(dir, "world"))
	if string(b) != "old" {
		t.Fatal("removed world")
	}
}
func TestMCJarsPublicationRollback(t *testing.T) {
	dir := t.TempDir()
	root, err := rooted.Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	root.WriteFile("server.jar", []byte("old jar"), 0644)
	root.Mkdir("libraries", 0755)
	root.WriteFile("libraries/old", []byte("old lib"), 0644)
	name, stage, err := root.TempDir()
	if err != nil {
		t.Fatal(err)
	}
	defer root.RemoveAll(name)
	defer stage.Close()
	stage.WriteFile("server.jar", []byte("new jar"), 0644)
	// Missing staged libraries forces failure after the JAR has been published.
	keep, err := mcjarsPublish(root, name, []mcjarsPublication{{"server.jar", "server.jar"}, {"libraries", "libraries"}})
	if err == nil || keep {
		t.Fatal("expected recoverable rollback", keep, err)
	}
	for n, want := range map[string]string{"server.jar": "old jar", "libraries/old": "old lib"} {
		b, e := root.ReadFile(n)
		if e != nil || string(b) != want {
			t.Fatal("rollback failed", n, e)
		}
	}
}
func TestMCJarsOfficialPathChecksums(t *testing.T) {
	jar := testMCJarsJar(t)
	sum := sha256.Sum256([]byte(jar))
	for _, digest := range []string{hex.EncodeToString(sum[:]), strings.Repeat("0", 64)} {
		t.Run(digest, func(t *testing.T) {
			u := "https://fill-data.papermc.io/v1/objects/" + digest + "/paper.jar"
			mockLatestMetadata(t, map[string]string{mcjarsAPI + "types/PAPER/versions/latest/latest": testMCJarsPayload("PAPER", "1.21.4", "21", u, len(jar)), u: jar})
			err := DownloadServerJar("paper", "latest", t.TempDir(), "server.jar")
			if (err == nil) != (digest == hex.EncodeToString(sum[:])) {
				t.Fatal(err)
			}
		})
	}
}

func TestMCJarsArchiveBudgetsAndLinks(t *testing.T) {
	for _, tc := range []struct {
		name    string
		members []*zip.File
	}{
		{"expanded", []*zip.File{{FileHeader: zip.FileHeader{Name: "server.jar", UncompressedSize64: 1<<30 + 1}}}},
		{"depth", []*zip.File{{FileHeader: zip.FileHeader{Name: "libraries/" + strings.Repeat("a/", 33) + "x"}}}},
		{"traversal", []*zip.File{{FileHeader: zip.FileHeader{Name: "libraries/../world"}}}},
		{"backslash", []*zip.File{{FileHeader: zip.FileHeader{Name: `libraries\world`}}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := mcjarsZIPMembers(&zip.Reader{File: tc.members}, true); err == nil {
				t.Fatal("accepted invalid archive")
			}
		})
	}
	var b bytes.Buffer
	w := zip.NewWriter(&b)
	h := &zip.FileHeader{Name: "libraries/link"}
	h.SetMode(os.ModeSymlink | 0777)
	f, err := w.CreateHeader(h)
	if err != nil {
		t.Fatal(err)
	}
	f.Write([]byte("../../world"))
	w.Close()
	zr, err := zip.NewReader(bytes.NewReader(b.Bytes()), int64(b.Len()))
	if err != nil {
		t.Fatal(err)
	}
	if err = mcjarsZIPMembers(zr, true); err == nil {
		t.Fatal("accepted symlink")
	}
	// EOCD entry count is checked before archive/zip can allocate entries.
	data := []byte(testMCJarsJar(t))
	data[len(data)-12] = 0xff
	data[len(data)-11] = 0xff
	root, err := rooted.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	root.WriteFile("bad.jar", data, 0600)
	file, err := root.Open("bad.jar")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	if _, err = mcjarsZIP(file); err == nil {
		t.Fatal("accepted huge entry count")
	}
}

func TestMCJarsProxyNullVersionsAndHistoricalJava(t *testing.T) {
	for _, java := range []string{"7", "8", "11", "16", "17", "21", "25", "null"} {
		t.Run(java, func(t *testing.T) {
			payload := testMCJarsPayload("BUNGEECORD", "1.21", "7", testMCJarsURL, 100)
			payload = strings.Replace(payload, `"project_version_id":"1.21"`, `"project_version_id":null`, 1)
			payload = strings.Replace(payload, `"java":7`, `"java":`+java, 1)
			mockLatestMetadata(t, map[string]string{mcjarsAPI + "types/BUNGEECORD/versions/latest/latest": payload, mcjarsAPI + testMCJarsID: payload})
			b, err := ResolveMCJarsBuild(context.Background(), "bungeecord", "latest")
			if err != nil || b.Info.MinecraftVersion != "" || b.Info.ProjectVersion != "" {
				t.Fatal(b, err)
			}
			if java == "null" && (b.Info.RequiredJava != 0 || b.Info.JavaSource != "") {
				t.Fatal("invented Java", b)
			}
		})
	}
	bad := strings.Replace(testMCJarsPayload("VELOCITY", "4.3.2-SNAPSHOT", "25", testMCJarsURL, 100), `"version_id":null`, `"version_id":"1.21.4"`, 1)
	mockLatestMetadata(t, map[string]string{mcjarsAPI + "types/VELOCITY/versions/latest/latest": bad})
	if _, err := ResolveMCJarsBuild(context.Background(), "velocity", "latest"); err == nil {
		t.Fatal("accepted proxy Minecraft version")
	}
}

func TestMCJarsHashContradictionAndUnknownSteps(t *testing.T) {
	hash := strings.Repeat("a", 64)
	payload := testMCJarsPayload("PAPER", "1.21.4", "21", "https://fill-data.papermc.io/v1/objects/"+strings.Repeat("b", 64)+"/paper.jar", 100)
	mockLatestMetadata(t, map[string]string{mcjarsAPI + hash: payload})
	if _, err := LookupMCJarsBuild(context.Background(), hash); err == nil {
		t.Fatal("accepted contradictory hash")
	}
	payload = strings.Replace(testMCJarsPayload("PAPER", "1.21.4", "21", testMCJarsURL, 100), `"type":"download"`, `"type":"download","command":"run.sh"`, 1)
	mockLatestMetadata(t, map[string]string{mcjarsAPI + "types/PAPER/versions/latest/latest": payload})
	if _, err := ResolveMCJarsBuild(context.Background(), "paper", "latest"); err == nil {
		t.Fatal("accepted unknown step fields")
	}
}

func TestMCJarsBinaryRedirectChecks(t *testing.T) {
	for _, target := range []string{"https://127.0.0.1/server.jar", "http://files.mcjars.app/server.jar", "https://user@files.mcjars.app/server.jar", "https://files.mcjars.app/allowed.jar"} {
		t.Run(target, func(t *testing.T) {
			old := httpClient
			t.Cleanup(func() { httpClient = old })
			calls := 0
			jar := testMCJarsJar(t)
			httpClient = &http.Client{Transport: modrinthTransport(func(r *http.Request) (*http.Response, error) {
				calls++
				if calls == 1 {
					return &http.Response{StatusCode: 302, Header: http.Header{"Location": []string{target}}, Body: io.NopCloser(strings.NewReader(""))}, nil
				}
				return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(jar))}, nil
			})}
			root, err := rooted.Open(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			defer root.Close()
			err = mcjarsDownload(context.Background(), root, mcjarsStep{URL: testMCJarsURL, Size: int64(len(jar))})
			if strings.Contains(target, "allowed.jar") {
				if err != nil || calls != 2 {
					t.Fatal(err, calls)
				}
			} else if err == nil || calls != 1 {
				t.Fatal("followed unsafe redirect", err, calls)
			}
		})
	}
}
