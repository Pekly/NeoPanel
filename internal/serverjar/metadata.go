package serverjar

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"io"
	"path"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

var className = regexp.MustCompile(`^[A-Za-z_$][A-Za-z0-9_$]*(\.[A-Za-z_$][A-Za-z0-9_$]*)*$`)
var versionID = regexp.MustCompile(`^[0-9][0-9A-Za-z.+_-]{0,79}$`)

func inspectArchive(ctx context.Context, r io.ReaderAt, size int64, budget *inspectionBudget, depth int) (Info, error) {
	info := Info{Status: "unknown", Source: "embedded"}
	z, err := boundedZip(ctx, r, size)
	if err != nil {
		return info, err
	}
	files := make(map[string]*zip.File, len(z.File))
	for _, f := range z.File {
		files[f.Name] = f
	}
	partial := false
	read := func(name string) []byte {
		f := files[name]
		if f == nil {
			return nil
		}
		data, err := budget.read(ctx, f, false)
		if err != nil {
			partial = true
			return nil
		}
		return data
	}
	manifest := map[string]string{}
	if data := read("META-INF/MANIFEST.MF"); data != nil {
		manifest, err = mainManifest(data)
		if err != nil {
			manifest = map[string]string{}
			partial = true
		}
	}
	main := manifest["main-class"]
	info.Software = softwareForMain(main)
	if main == "org.bukkit.craftbukkit.Main" && (manifest["brand-id"] == "papermc:paper" || manifest["implementation-title"] == "Paper") {
		info.Software = "paper"
	}
	if className.MatchString(main) {
		if f := files[strings.ReplaceAll(main, ".", "/")+".class"]; f != nil {
			// Only the actual base entry point. No broad class scan, and never
			// optional META-INF/versions/* classes in a multi-release archive.
			data, e := budget.read(ctx, f, false)
			if e != nil {
				partial = true
			} else {
				info.MinimumJava = classMinimum(data)
				if info.MinimumJava == 0 {
					partial = true
				}
			}
		}
	}
	if info.Software != "" && info.Software != "vanilla" {
		info.ProjectVersion = safeText(manifest["implementation-version"])
	}
	mcEngine := info.Software != "" && info.Software != "velocity" && info.Software != "bungeecord"
	if mcEngine {
		if v := manifest["minecraft-version"]; versionID.MatchString(v) {
			info.MinecraftVersion = v
		}
		if data := read("version.json"); data != nil {
			var v struct {
				ID   string          `json:"id"`
				Java json.RawMessage `json:"java_version"`
			}
			if strictJSON(data, &v) != nil || !versionID.MatchString(v.ID) {
				partial = true
			} else {
				if info.MinecraftVersion != "" && info.MinecraftVersion != v.ID {
					info.MinecraftVersion = ""
					partial = true
				} else {
					info.MinecraftVersion = v.ID
				}
				if len(v.Java) != 0 {
					info.RequiredJava = runtimeMajor(v.Java)
					if info.RequiredJava > 0 {
						info.JavaSource = "embedded"
					} else {
						partial = true
					}
				}
			}
		}
	}
	if info.Software == "fabric" {
		if data := read("fabric.mod.json"); data != nil {
			var mod struct {
				ID      string `json:"id"`
				Version string `json:"version"`
			}
			if strictJSON(data, &mod) != nil || mod.ID != "fabricloader" || safeText(mod.Version) == "" {
				partial = true
			} else {
				// This is the loader's release, never Minecraft's release.
				info.ProjectVersion = safeText(mod.Version)
			}
		}
	}
	if main == "net.minecraft.bundler.Main" || main == "io.papermc.paperclip.Main" || main == "io.papermc.paperclip.Paperclip" || main == "com.destroystokyo.paperclip.Paperclip" {
		if depth != 0 {
			partial = true
		} else {
			data := read("META-INF/versions.list")
			ref, digest, ok := bundleReference(data)
			if !ok || files[ref] == nil {
				partial = true
			} else {
				inner, e := budget.read(ctx, files[ref], true)
				if e != nil {
					partial = true
				} else {
					sum := sha256.Sum256(inner)
					if hex.EncodeToString(sum[:]) != digest {
						partial = true
					} else {
						nested, e := inspectArchive(ctx, contextAt{ctx, bytes.NewReader(inner)}, int64(len(inner)), budget, depth+1)
						if e != nil || (nested.Software != info.Software && !(info.Software == "paper" && nested.Software == "vanilla")) {
							partial = true
						} else {
							if nested.Status == "partial" {
								partial = true
							}
							if info.MinecraftVersion != "" && nested.MinecraftVersion != "" && info.MinecraftVersion != nested.MinecraftVersion {
								info.MinecraftVersion = ""
								partial = true
							} else if info.MinecraftVersion == "" {
								info.MinecraftVersion = nested.MinecraftVersion
							}
							if info.RequiredJava != 0 && nested.RequiredJava != 0 && info.RequiredJava != nested.RequiredJava {
								info.RequiredJava = 0
								info.JavaSource = ""
								partial = true
							} else if info.RequiredJava == 0 {
								info.RequiredJava, info.JavaSource = nested.RequiredJava, nested.JavaSource
							}
							info.MinimumJava = max(info.MinimumJava, nested.MinimumJava)
							if nested.Software == info.Software && info.ProjectVersion == "" {
								info.ProjectVersion = nested.ProjectVersion
							}
						}
					}
				}
			}
		}
	}
	if ctx.Err() != nil {
		return info, ctx.Err()
	}
	if info.Software != "" {
		info.Status = "identified"
		if info.MinecraftVersion == "" && info.ProjectVersion == "" {
			partial = true
		}
	}
	if partial {
		info.Status = "partial"
		info.Message = "Some embedded server metadata is unavailable, unsupported, or invalid"
	}
	return info, nil
}

func softwareForMain(main string) string {
	switch main {
	case "net.minecraft.server.Main", "net.minecraft.server.MinecraftServer", "net.minecraft.bundler.Main":
		return "vanilla"
	case "io.papermc.paperclip.Main", "io.papermc.paperclip.Paperclip", "com.destroystokyo.paperclip.Paperclip":
		return "paper"
	case "com.velocitypowered.proxy.Velocity":
		return "velocity"
	case "net.md_5.bungee.Bootstrap", "net.md_5.bungee.BungeeCord":
		return "bungeecord"
	case "net.fabricmc.loader.impl.launch.server.FabricServerLauncher", "net.fabricmc.loader.launch.server.FabricServerLauncher", "net.fabricmc.installer.ServerLauncher":
		return "fabric"
	case "net.minecraftforge.server.ServerMain", "net.minecraftforge.fml.relauncher.ServerLaunchWrapper", "cpw.mods.fml.relauncher.ServerLaunchWrapper":
		return "forge"
	case "net.neoforged.server.ServerMain":
		return "neoforge"
	}
	// CraftBukkit and generic ModLauncher/BootstrapLauncher are shared by
	// multiple engines: the entry point alone cannot identify Paper/Forge.
	return ""
}

// Main-Class in a later per-entry section is not the executable entry point.
func mainManifest(data []byte) (map[string]string, error) {
	if !utf8.Valid(data) || bytes.IndexByte(data, 0) >= 0 {
		return nil, errMetadata
	}
	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	m := make(map[string]string)
	key := ""
	var value strings.Builder
	for _, line := range lines {
		if line == "" {
			break
		}
		if strings.HasPrefix(line, " ") {
			if key == "" {
				return nil, errMetadata
			}
			value.WriteString(line[1:])
			continue
		}
		if key != "" {
			m[key] = value.String()
		}
		value.Reset()
		k, v, ok := strings.Cut(line, ": ")
		if !ok || k == "" || strings.ContainsAny(k, " \t\r:") {
			return nil, errMetadata
		}
		key = strings.ToLower(k)
		if _, exists := m[key]; exists {
			return nil, errMetadata
		}
		m[key] = ""
		value.WriteString(v)
	}
	if key != "" {
		m[key] = value.String()
	}
	return m, nil
}

func safeText(s string) string {
	if s == "" || len(s) > 160 || !utf8.ValidString(s) || strings.TrimSpace(s) != s || strings.IndexFunc(s, unicode.IsControl) >= 0 {
		return ""
	}
	return s
}

func classMinimum(data []byte) int {
	if len(data) < 8 || binary.BigEndian.Uint32(data) != 0xcafebabe {
		return 0
	}
	minor, major := binary.BigEndian.Uint16(data[4:]), binary.BigEndian.Uint16(data[6:])
	if major < 45 || major > 74 || !(minor == 0 || (major == 45 && minor <= 3) || (major >= 56 && minor == 65535)) {
		return 0
	}
	return int(major) - 44
}

func runtimeMajor(raw json.RawMessage) int {
	var major int
	if json.Unmarshal(raw, &major) != nil {
		var object struct {
			Major int `json:"majorVersion"`
		}
		if strictJSON(raw, &object) != nil {
			return 0
		}
		major = object.Major
	}
	if major < 1 || major > 30 {
		return 0
	}
	return major
}

func bundleReference(data []byte) (string, string, bool) {
	line := strings.TrimSuffix(strings.TrimSuffix(string(data), "\n"), "\r")
	parts := strings.Split(line, "\t")
	if len(parts) != 3 || strings.ContainsAny(line, "\r\n") || len(parts[0]) != 64 || safeText(parts[1]) == "" {
		return "", "", false
	}
	if _, err := hex.DecodeString(parts[0]); err != nil {
		return "", "", false
	}
	name := parts[2]
	if name == "" || strings.ContainsAny(name, "\\:\x00") || strings.HasPrefix(name, "/") || path.Clean(name) != name || name == ".." || strings.HasPrefix(name, "../") || !strings.HasSuffix(name, ".jar") {
		return "", "", false
	}
	return "META-INF/versions/" + name, strings.ToLower(parts[0]), true
}

// encoding/json normally silently accepts duplicate keys. Metadata is evidence,
// so reject duplicates (including case aliases used by struct field matching),
// trailing values, and excessive nesting before decoding the known schema.
func strictJSON(data []byte, dst any) error {
	d := json.NewDecoder(bytes.NewReader(data))
	d.UseNumber()
	var value func(int) error
	value = func(depth int) error {
		if depth > 32 {
			return errMetadata
		}
		t, err := d.Token()
		if err != nil {
			return errMetadata
		}
		delim, ok := t.(json.Delim)
		if !ok {
			return nil
		}
		switch delim {
		case '{':
			keys := map[string]bool{}
			for d.More() {
				k, e := d.Token()
				if e != nil {
					return errMetadata
				}
				key, ok := k.(string)
				if !ok {
					return errMetadata
				}
				key = strings.ToLower(key)
				if keys[key] {
					return errMetadata
				}
				keys[key] = true
				if e := value(depth + 1); e != nil {
					return e
				}
			}
		case '[':
			for d.More() {
				if e := value(depth + 1); e != nil {
					return e
				}
			}
		default:
			return errMetadata
		}
		end, e := d.Token()
		if e != nil || (delim == '{' && end != json.Delim('}')) || (delim == '[' && end != json.Delim(']')) {
			return errMetadata
		}
		return nil
	}
	if value(0) != nil {
		return errMetadata
	}
	if _, err := d.Token(); err != io.EOF {
		return errMetadata
	}
	if json.Unmarshal(data, dst) != nil {
		return errMetadata
	}
	return nil
}
