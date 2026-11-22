package serverjar

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Pekly/NeoPanel/internal/rooted"
)

type entry struct {
	name string
	data []byte
}

func text(name, data string) entry { return entry{name, []byte(data)} }
func jar(t *testing.T, entries ...entry) []byte {
	t.Helper()
	var b bytes.Buffer
	w := zip.NewWriter(&b)
	for _, e := range entries {
		f, err := w.Create(e.name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := f.Write(e.data); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}
func manifest(main, extra string) entry {
	return text("META-INF/MANIFEST.MF", "Manifest-Version: 1.0\r\nMain-Class: "+main+"\r\n"+extra+"\r\n")
}
func class(main string, major int) entry {
	b := []byte{0xca, 0xfe, 0xba, 0xbe, 0, 0, 0, 0}
	binary.BigEndian.PutUint16(b[6:], uint16(major))
	return entry{strings.ReplaceAll(main, ".", "/") + ".class", b}
}
func inspectBytes(t *testing.T, data []byte) (Info, error) {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "paper-99.99.jar"), data, 0600); err != nil {
		t.Fatal(err)
	}
	info, err := Inspect(context.Background(), dir, "paper-99.99.jar")
	if err == nil {
		sum := sha256.Sum256(data)
		if info.SHA256 != hex.EncodeToString(sum[:]) || info.Size != int64(len(data)) || info.Filename != "paper-99.99.jar" || info.Modified.IsZero() || info.Source != "embedded" {
			t.Fatalf("bad file facts: %+v", info)
		}
	}
	return info, err
}

func TestUnknownDeceptiveMetadata(t *testing.T) {
	for _, entries := range [][]entry{
		{text("README", "paper 1.21")},
		{manifest("example.Main", "Implementation-Title: Paper\r\nImplementation-Version: 1.21.4\r\nBuild-Jdk: 21\r\n"), text("version.json", `{"id":"1.21.4","java_version":21}`)},
		{text("META-INF/MANIFEST.MF", "Manifest-Version: 1.0\nImplementation-Version: 1.21.4\n\nName: ignored\nMain-Class: net.minecraft.server.Main\n")},
		{manifest("cpw.mods.modlauncher.Launcher", "Implementation-Version: 1.21.4\r\n")},
	} {
		info, err := inspectBytes(t, jar(t, entries...))
		if err != nil || info.Status != "unknown" || info.Software != "" || info.MinecraftVersion != "" || info.ProjectVersion != "" || info.RequiredJava != 0 {
			t.Fatalf("invented evidence: %+v %v", info, err)
		}
	}
}

func TestSoftwareAndSeparateVersions(t *testing.T) {
	cases := []struct{ main, software, extra string }{
		{"net.minecraft.server.Main", "vanilla", ""},
		{"org.bukkit.craftbukkit.Main", "paper", "Brand-Id: papermc:paper\r\n"},
		{"com.velocitypowered.proxy.Velocity", "velocity", ""},
		{"net.md_5.bungee.Bootstrap", "bungeecord", ""},
		{"net.fabricmc.loader.impl.launch.server.FabricServerLauncher", "fabric", ""},
		{"net.minecraftforge.server.ServerMain", "forge", ""},
		{"net.neoforged.server.ServerMain", "neoforge", ""},
	}
	for _, tc := range cases {
		t.Run(tc.software, func(t *testing.T) {
			info, err := inspectBytes(t, jar(t, manifest(tc.main, tc.extra+"Implementation-Version: 3.4.0\r\nBuild-Jdk: 25\r\n"), class(tc.main, 61), text("version.json", `{"id":"1.21.4","java_version":{"majorVersion":21}}`)))
			if err != nil || info.Software != tc.software || info.Status != "identified" || info.MinimumJava != 17 {
				t.Fatalf("%+v %v", info, err)
			}
			if tc.software == "vanilla" {
				if info.ProjectVersion != "" {
					t.Fatal(info)
				}
			} else if info.ProjectVersion != "3.4.0" {
				t.Fatal(info)
			}
			if tc.software == "velocity" || tc.software == "bungeecord" {
				if info.MinecraftVersion != "" || info.RequiredJava != 0 || info.JavaSource != "" {
					t.Fatalf("proxy version was Minecraft: %+v", info)
				}
			} else if info.MinecraftVersion != "1.21.4" || info.RequiredJava != 21 || info.JavaSource != "embedded" {
				t.Fatal(info)
			}
		})
	}
}

func TestFoldedManifestAndMultiRelease(t *testing.T) {
	mr := class("net.minecraft.server.Main", 69)
	mr.name = "META-INF/versions/25/" + mr.name
	info, err := inspectBytes(t, jar(t,
		text("META-INF/MANIFEST.MF", "Manifest-Version: 1.0\r\nMain-Class: net.minecraft.\r\n server.Main\r\nMulti-Release: true\r\nBuild-Jdk-Spec: 25\r\n\r\nName: ignored\r\nMain-Class: fake.Main\r\n"),
		class("net.minecraft.server.Main", 52), mr, class("other.Optional", 74)))
	if err != nil || info.Software != "vanilla" || info.MinimumJava != 8 || info.RequiredJava != 0 || info.JavaSource != "" || info.Status != "partial" {
		t.Fatalf("%+v %v", info, err)
	}
}

func TestVersionAndManifestMalformed(t *testing.T) {
	for _, version := range []string{`{`, `{"id":"1.21","id":"1.20","java_version":21}`, `{"id":"1.21","ID":"1.20"}`, `{"id":"1.21"} {}`, `{"id":"1.21","java_version":31}`, `{"id":"1.21","java_version":"21"}`, `{"id":"1.21","java_version":{"majorVersion":21,"majorVersion":17}}`} {
		info, err := inspectBytes(t, jar(t, manifest("net.minecraft.server.Main", ""), text("version.json", version)))
		if err != nil || info.Status != "partial" || info.RequiredJava != 0 {
			t.Fatalf("%q: %+v %v", version, info, err)
		}
	}
	for _, m := range []string{"Main-Class: net.minecraft.server.Main\nmain-class: com.velocitypowered.proxy.Velocity\n", " orphan\n", "Main-Class net.minecraft.server.Main\n"} {
		info, err := inspectBytes(t, jar(t, text("META-INF/MANIFEST.MF", m)))
		if err != nil || info.Status != "partial" || info.Software != "" {
			t.Fatalf("%+v %v", info, err)
		}
	}
	info, err := inspectBytes(t, jar(t, manifest("net.minecraft.server.Main", ""), text("version.json", `{"id":"1.20.4","java_version":17}`)))
	if err != nil || info.RequiredJava != 17 {
		t.Fatalf("numeric Java: %+v %v", info, err)
	}
}

func TestFabricLoaderDescriptor(t *testing.T) {
	info, err := inspectBytes(t, jar(t, manifest("net.fabricmc.loader.impl.launch.server.FabricServerLauncher", ""), text("fabric.mod.json", `{"id":"fabricloader","version":"0.16.10"}`)))
	if err != nil || info.Software != "fabric" || info.ProjectVersion != "0.16.10" || info.MinecraftVersion != "" || info.RequiredJava != 0 {
		t.Fatalf("%+v %v", info, err)
	}
}

func bundled(t *testing.T, main string, inner []byte, listOverride string) []byte {
	sum := sha256.Sum256(inner)
	list := fmt.Sprintf("%x\t1.21.4\t1.21.4/server.jar\n", sum)
	if listOverride != "" {
		list = listOverride
	}
	return jar(t, manifest(main, ""), text("META-INF/versions.list", list), entry{"META-INF/versions/1.21.4/server.jar", inner}, text("META-INF/libraries/danger.jar", "not a jar; must never be inspected"))
}
func TestSingleReferencedBundle(t *testing.T) {
	inner := jar(t, manifest("net.minecraft.server.Main", ""), class("net.minecraft.server.Main", 65), text("version.json", `{"id":"1.21.4","java_version":21}`))
	for _, main := range []string{"net.minecraft.bundler.Main", "io.papermc.paperclip.Main"} {
		info, err := inspectBytes(t, bundled(t, main, inner, ""))
		if err != nil || info.Software != softwareForMain(main) || info.MinecraftVersion != "1.21.4" || info.RequiredJava != 21 || info.MinimumJava != 21 || info.Status != "identified" {
			t.Fatalf("%+v %v", info, err)
		}
	}
	for _, list := range []string{
		strings.Repeat("0", 64) + "\t1.21.4\t1.21.4/server.jar\n",
		strings.Repeat("0", 64) + "\t1.21.4\t../server.jar\n",
		strings.Repeat("0", 64) + "\t1.21.4\t1.21.4/server.jar\nsecond\n",
	} {
		info, err := inspectBytes(t, bundled(t, "net.minecraft.bundler.Main", inner, list))
		if err != nil || info.Status != "partial" || info.MinecraftVersion != "" || info.RequiredJava != 0 {
			t.Fatalf("%+v %v", info, err)
		}
	}
	info, err := inspectBytes(t, bundled(t, "net.minecraft.bundler.Main", bundled(t, "net.minecraft.bundler.Main", inner, ""), ""))
	if err != nil || info.Status != "partial" || info.RequiredJava != 0 {
		t.Fatalf("depth limit: %+v %v", info, err)
	}
}

func TestArchivePreflight(t *testing.T) {
	base := jar(t, text("README", "data"))
	mutate := func(fn func([]byte, int)) []byte { b := bytes.Clone(base); fn(b, len(b)-22); return b }
	bad := [][]byte{
		base[:len(base)-1], []byte("not a jar"), append([]byte("prefix"), base...),
		jar(t, manifest("net.minecraft.server.Main", ""), manifest("example.Main", "")),
		mutate(func(b []byte, e int) { binary.LittleEndian.PutUint16(b[e+10:], 50001) }),
		mutate(func(b []byte, e int) { binary.LittleEndian.PutUint32(b[e+12:], centralLimit+1) }),
		mutate(func(b []byte, e int) { binary.LittleEndian.PutUint16(b[e+10:], 0) }),
		mutate(func(b []byte, e int) { binary.LittleEndian.PutUint16(b[e+4:], 1) }),
		mutate(func(b []byte, e int) { binary.LittleEndian.PutUint32(b[e+16:], 0xffffffff) }),
		mutate(func(b []byte, e int) {
			c := int(binary.LittleEndian.Uint32(b[e+16:]))
			binary.LittleEndian.PutUint32(b[c+24:], 0xffffffff)
		}),
		mutate(func(b []byte, e int) { b[30] = 'X' }),
	}
	for i, data := range bad {
		info, err := inspectBytes(t, data)
		if err == nil || info.Status != "error" || info.Software != "" || strings.Contains(err.Error(), "\\") {
			t.Fatalf("case %d: %+v %v", i, info, err)
		}
	}
	if info, err := inspectBytes(t, jar(t)); err != nil || info.Status != "unknown" {
		t.Fatalf("empty ZIP: %+v %v", info, err)
	}
}

func TestResourceBudgets(t *testing.T) {
	info, err := inspectBytes(t, jar(t, text("META-INF/MANIFEST.MF", strings.Repeat("A", metadataLimit+1))))
	if err != nil || info.Status != "partial" || info.Software != "" {
		t.Fatalf("%+v %v", info, err)
	}
	for _, tc := range []struct {
		b      inspectionBudget
		n      uint64
		nested bool
	}{
		{inspectionBudget{}, metadataLimit + 1, false},
		{inspectionBudget{metadata: metadataTotalLimit}, 1, false},
		{inspectionBudget{}, nestedLimit + 1, true},
		{inspectionBudget{expanded: expandedLimit}, 1, true},
	} {
		f := &zip.File{FileHeader: zip.FileHeader{UncompressedSize64: tc.n}}
		if _, err := tc.b.read(context.Background(), f, tc.nested); !errors.Is(err, errLimit) {
			t.Fatalf("budget not enforced: %v", err)
		}
	}
}

func TestCancellationAndConfinement(t *testing.T) {
	dir := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := Inspect(ctx, dir, "server.jar"); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	ctx, cancel = context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()
	if _, err := Inspect(ctx, dir, "server.jar"); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal(err)
	}
	for _, name := range []string{"../outside.jar", "a/b.jar", `a\b.jar`, "C:server.jar", ".", "server.jar ", "server.jar.", "server\x00.jar"} {
		if _, err := Inspect(context.Background(), dir, name); err == nil || strings.Contains(err.Error(), dir) {
			t.Fatalf("name %q: %v", name, err)
		}
	}
	info, err := Inspect(context.Background(), dir, "absent.jar")
	if err == nil || info.Status != "missing" || strings.Contains(err.Error(), dir) {
		t.Fatalf("%+v %v", info, err)
	}
	if err := os.Mkdir(filepath.Join(dir, "directory.jar"), 0700); err != nil {
		t.Fatal(err)
	}
	if _, err := Inspect(context.Background(), dir, "directory.jar"); err == nil {
		t.Fatal("accepted directory")
	}
	t.Run("symlink", func(t *testing.T) {
		outside := filepath.Join(t.TempDir(), "outside.jar")
		if err := os.WriteFile(outside, jar(t), 0600); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(outside, filepath.Join(dir, "link.jar")); err != nil {
			t.Skipf("symlink unavailable: %v", err)
		}
		if _, err := Inspect(context.Background(), dir, "link.jar"); err == nil {
			t.Fatal("accepted symlink")
		}
	})
}

func TestStableFileRejectsReplacementAndMutation(t *testing.T) {
	for _, replacement := range []bool{false, true} {
		t.Run(fmt.Sprint(replacement), func(t *testing.T) {
			dir := t.TempDir()
			name := filepath.Join(dir, "server.jar")
			if err := os.WriteFile(name, jar(t, text("README", "old")), 0600); err != nil {
				t.Fatal(err)
			}
			root, err := rooted.Open(dir)
			if err != nil {
				t.Fatal(err)
			}
			defer root.Close()
			f, err := root.Open("server.jar")
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()
			before, err := f.Stat()
			if err != nil {
				t.Fatal(err)
			}
			if err := stableFile(root, f, "server.jar", before); err != nil {
				t.Fatal(err)
			}
			if replacement {
				if err := os.Rename(name, filepath.Join(dir, "old.jar")); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(name, jar(t, text("README", "new")), 0600); err != nil {
					t.Fatal(err)
				}
				if err := os.Chtimes(name, before.ModTime(), before.ModTime()); err != nil {
					t.Fatal(err)
				}
			} else {
				if err := os.WriteFile(name, []byte("changed"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			if err := stableFile(root, f, "server.jar", before); !errors.Is(err, errChanged) {
				t.Fatalf("missed change: %v", err)
			}
		})
	}
}

func TestClassMinimumRejectsBadHeaders(t *testing.T) {
	for _, b := range [][]byte{nil, []byte("12345678"), class("X", 44).data, class("X", 75).data, {0xca, 0xfe, 0xba, 0xbe, 0, 1, 0, 61}} {
		if got := classMinimum(b); got != 0 {
			t.Fatalf("invalid header minimum %d", got)
		}
	}
}

type actionContext struct {
	context.Context
	calls  int
	at     int
	action func()
}

func (c *actionContext) Err() error {
	c.calls++
	if c.calls == c.at {
		c.action()
	}
	return c.Context.Err()
}

func TestInspectReplacementDuringRead(t *testing.T) {
	dir := t.TempDir()
	name := filepath.Join(dir, "server.jar")
	data := jar(t, manifest("net.minecraft.server.Main", ""))
	if err := os.WriteFile(name, data, 0600); err != nil {
		t.Fatal(err)
	}
	ctx := &actionContext{Context: context.Background(), at: 3, action: func() {
		if err := os.Rename(name, filepath.Join(dir, "old.jar")); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(name, data, 0600); err != nil {
			t.Fatal(err)
		}
	}}
	info, err := Inspect(ctx, dir, "server.jar")
	if !errors.Is(err, errChanged) || info.SHA256 != "" || info.Software != "" {
		t.Fatalf("accepted replaced path: %+v %v", info, err)
	}
}

func TestCancellationDuringRead(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "server.jar"), jar(t), 0600); err != nil {
		t.Fatal(err)
	}
	base, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx := &actionContext{Context: base, at: 2, action: cancel}
	if _, err := Inspect(ctx, dir, "server.jar"); !errors.Is(err, context.Canceled) {
		t.Fatalf("read did not cancel: %v", err)
	}
}

func TestEOCDCommentAmbiguity(t *testing.T) {
	data := jar(t, text("README", "test"))
	// A second EOCD-like structure hidden inside the real EOCD's comment,
	// with trailing bytes, must not make preflight and archive/zip disagree.
	eocd := len(data) - 22
	comment := make([]byte, 30)
	binary.LittleEndian.PutUint32(comment, 0x06054b50)
	binary.LittleEndian.PutUint16(comment[10:], 60000)
	binary.LittleEndian.PutUint16(data[eocd+20:], uint16(len(comment)))
	data = append(data, comment...)
	if _, err := inspectBytes(t, data); err == nil {
		t.Fatal("accepted ambiguous EOCD")
	}
}

func TestFoldedManifestLinearAndExactVersionPath(t *testing.T) {
	m, err := mainManifest([]byte("Long: a\n" + strings.Repeat(" b\n", 30000) + "Main-Class: net.minecraft.server.Main\n"))
	if err != nil || len(m["long"]) != 30001 {
		t.Fatalf("folding: %v", err)
	}
	info, err := inspectBytes(t, jar(t, manifest("net.minecraft.server.Main", ""), text("fake/version.json", `{"id":"1.21.4","java_version":21}`)))
	if err != nil || info.MinecraftVersion != "" || info.RequiredJava != 0 {
		t.Fatalf("used non-exact metadata path: %+v %v", info, err)
	}
}
