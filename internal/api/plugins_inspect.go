package api

import (
	"archive/zip"
	"context"
	"crypto/sha512"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"unicode"

	"github.com/Pekly/NeoPanel/internal/rooted"
	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

const pluginFileLimit int64 = 200 << 20
const pluginBatchLimit int64 = 512 << 20
const pluginDescriptorLimit = 128 << 10

func strictPluginFilename(name string) bool {
	if len(name) > 255 || !validInstalledPluginName(name) || !rooted.Basename(name) {
		return false
	}
	for _, c := range name {
		if unicode.IsControl(c) {
			return false
		}
	}
	return true
}

func pluginDigest(ctx context.Context, dir *rooted.Root, name string, remaining *int64) (string, os.FileInfo, error) {
	if !strictPluginFilename(name) {
		return "", nil, fmt.Errorf("invalid plugin filename")
	}
	f, err := dir.Open(name)
	if err != nil {
		return "", nil, fmt.Errorf("cannot read regular plugin file")
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		return "", nil, fmt.Errorf("cannot inspect plugin file")
	}
	if st.Size() > pluginFileLimit || st.Size() > *remaining {
		return "", st, fmt.Errorf("plugin exceeds file or page byte limit")
	}
	h := sha512.New()
	n, err := rooted.Copy(ctx, h, f, st.Size())
	*remaining -= n
	if err != nil || n != st.Size() {
		return "", st, fmt.Errorf("plugin changed or could not be read")
	}
	after, err := f.Stat()
	if err != nil || after.Size() != st.Size() || !after.ModTime().Equal(st.ModTime()) {
		return "", st, fmt.Errorf("plugin changed while reading")
	}
	return hex.EncodeToString(h.Sum(nil)), st, nil
}

// Check the bounded non-ZIP64 central directory before archive/zip allocates its
// file index. JARs needing ZIP64 or more than 10,000 entries are not inspected.
func pluginZip(f *os.File, size int64) (*zip.Reader, error) {
	n := min(size, int64(65557))
	tail := make([]byte, n)
	if _, err := f.ReadAt(tail, size-n); err != nil {
		return nil, fmt.Errorf("invalid JAR archive")
	}
	found := false
	for i := len(tail) - 22; i >= 0; i-- {
		if binary.LittleEndian.Uint32(tail[i:]) != 0x06054b50 {
			continue
		}
		if i+22+int(binary.LittleEndian.Uint16(tail[i+20:])) != len(tail) {
			continue
		}
		count := binary.LittleEndian.Uint16(tail[i+10:])
		central := binary.LittleEndian.Uint32(tail[i+12:])
		if count > 10000 || central > 4<<20 || binary.LittleEndian.Uint16(tail[i+4:]) != 0 || binary.LittleEndian.Uint32(tail[i+16:]) == 0xffffffff {
			return nil, fmt.Errorf("JAR archive resource limit exceeded")
		}
		start := int64(binary.LittleEndian.Uint32(tail[i+16:]))
		end := start + int64(central)
		if end != size-n+int64(i) {
			return nil, fmt.Errorf("invalid JAR central directory bounds")
		}
		// Do not trust a forged small entry count/size to constrain archive/zip.
		pos := start
		var header [46]byte
		for j := 0; j < int(count); j++ {
			if pos+46 > end {
				return nil, fmt.Errorf("invalid JAR central directory")
			}
			if _, e := f.ReadAt(header[:], pos); e != nil || binary.LittleEndian.Uint32(header[:]) != 0x02014b50 {
				return nil, fmt.Errorf("invalid JAR central directory")
			}
			pos += 46 + int64(binary.LittleEndian.Uint16(header[28:])) + int64(binary.LittleEndian.Uint16(header[30:])) + int64(binary.LittleEndian.Uint16(header[32:]))
		}
		if pos != end {
			return nil, fmt.Errorf("invalid JAR central directory entry count")
		}
		found = true
		break
	}
	if !found {
		return nil, fmt.Errorf("invalid JAR archive")
	}
	z, err := zip.NewReader(f, size)
	if err != nil || len(z.File) > 10000 {
		return nil, fmt.Errorf("invalid or oversized JAR directory")
	}
	return z, nil
}

func pluginDescriptor(dir *rooted.Root, filename string) (string, string, error) {
	f, err := dir.Open(filename)
	if err != nil {
		return "", "", fmt.Errorf("cannot inspect JAR")
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil || st.Size() > pluginFileLimit {
		return "", "", fmt.Errorf("JAR exceeds resource limit")
	}
	z, err := pluginZip(f, st.Size())
	if err != nil {
		return "", "", err
	}
	remaining := int64(pluginDescriptorLimit)
	for _, name := range []string{"paper-plugin.yml", "plugin.yml", "bungee.yml", "velocity-plugin.json", "fabric.mod.json", "META-INF/mods.toml", "META-INF/neoforge.mods.toml"} {
		for _, entry := range z.File {
			if entry.Name != name {
				continue
			}
			if entry.UncompressedSize64 > uint64(remaining) || !entry.Mode().IsRegular() {
				return "", "", fmt.Errorf("plugin descriptor exceeds resource limit")
			}
			r, err := entry.Open()
			if err != nil {
				return "", "", fmt.Errorf("invalid plugin descriptor")
			}
			b, err := io.ReadAll(io.LimitReader(r, remaining+1))
			r.Close()
			remaining -= int64(len(b))
			if err != nil || remaining < 0 {
				return "", "", fmt.Errorf("invalid or oversized plugin descriptor")
			}
			var title, version string
			switch {
			case strings.HasSuffix(name, ".yml"):
				var doc yaml.Node
				if yaml.Unmarshal(b, &doc) != nil {
					return "", "", fmt.Errorf("invalid YAML plugin descriptor")
				}
				if len(doc.Content) != 1 || doc.Content[0].Kind != yaml.MappingNode {
					return "", "", fmt.Errorf("invalid YAML plugin descriptor")
				}
				nodes := doc.Content[0].Content
				for i := 0; i+1 < len(nodes); i += 2 {
					if nodes[i+1].Kind != yaml.ScalarNode {
						continue
					}
					switch nodes[i].Value {
					case "name":
						title = nodes[i+1].Value
					case "version":
						version = nodes[i+1].Value
					}
				}
			case strings.HasSuffix(name, ".json"):
				var v struct {
					Name    string `json:"name"`
					ID      string `json:"id"`
					Version string `json:"version"`
				}
				if json.Unmarshal(b, &v) != nil {
					return "", "", fmt.Errorf("invalid JSON plugin descriptor")
				}
				title, version = v.Name, v.Version
				if title == "" {
					title = v.ID
				}
			default:
				var v struct {
					Mods []struct {
						DisplayName string `toml:"displayName"`
						ModID       string `toml:"modId"`
						Version     string `toml:"version"`
					} `toml:"mods"`
				}
				if toml.Unmarshal(b, &v) != nil {
					return "", "", fmt.Errorf("invalid TOML plugin descriptor")
				}
				if len(v.Mods) > 0 {
					title, version = v.Mods[0].DisplayName, v.Mods[0].Version
					if title == "" {
						title = v.Mods[0].ModID
					}
				}
			}
			title, version = cleanPluginText(title), cleanPluginText(version)
			if title != "" {
				return title, version, nil
			}
		}
	}
	return "", "", nil
}

func cleanPluginText(s string) string {
	var out strings.Builder
	for _, c := range s {
		if unicode.IsLetter(c) || unicode.IsDigit(c) || strings.ContainsRune(" ._+-", c) {
			out.WriteRune(c)
		}
		if out.Len() >= 80 {
			break
		}
	}
	return strings.TrimSpace(out.String())
}

var pluginVersionSuffix = regexp.MustCompile(`(?i)[\s_-]+(?:v?\d|build[\s_-]*\d|snapshot|release).*$`)

func pluginFilenameQuery(filename string) string {
	s := filename
	if strings.HasSuffix(strings.ToLower(s), ".disabled") {
		s = s[:len(s)-9]
	}
	s = s[:len(s)-4]
	return cleanPluginText(pluginVersionSuffix.ReplaceAllString(s, ""))
}
func normalizedPluginName(s string) string {
	var b strings.Builder
	for _, c := range strings.ToLower(s) {
		if unicode.IsLetter(c) || unicode.IsDigit(c) {
			b.WriteRune(c)
		}
	}
	return b.String()
}
func crediblePluginName(query, title, slug string) bool {
	q := normalizedPluginName(query)
	switch q {
	case "", "plugin", "mod", "server", "download", "unknown":
		return false
	}
	return len(q) >= 3 && (q == normalizedPluginName(title) || q == normalizedPluginName(slug))
}
