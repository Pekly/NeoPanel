package downloader

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/Pekly/NeoPanel/internal/serverjar"
)

const mcjarsAPI = "https://mcjars.app/api/v3/builds/"
const mcjarsBinaryLimit int64 = 1 << 30

var ErrMCJarsNotFound = errors.New("MCJars build not found")
var mcjarsUUID = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
var mcjarsSHA256 = regexp.MustCompile(`^[0-9a-fA-F]{64}$`)

// MCJarsBuild contains provider evidence and a validated, non-executable install plan.
// Info does not describe installed bytes until DownloadServerJarWithInfo succeeds.
type MCJarsBuild struct {
	Info     serverjar.Info
	download mcjarsStep
	prepared bool
}

type mcjarsStep struct {
	Type     string `json:"type"`
	URL      string `json:"url,omitempty"`
	File     string `json:"file,omitempty"`
	Size     int64  `json:"size,omitempty"`
	Location string `json:"location,omitempty"`
}

func (s *mcjarsStep) UnmarshalJSON(data []byte) error {
	type plain mcjarsStep
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	return d.Decode((*plain)(s))
}

type mcjarsWireBuild struct {
	UUID         string         `json:"uuid"`
	Type         string         `json:"type"`
	Version      *string        `json:"version_id"`
	Project      *string        `json:"project_version_id"`
	Name         string         `json:"name"`
	Java         *int           `json:"java"`
	Installation [][]mcjarsStep `json:"installation"`
}
type mcjarsResponse struct {
	Build *mcjarsWireBuild `json:"build"`
	// Deliberately do not decode "latest": hash lookup must identify the matched build.
	Version *struct {
		ID   string `json:"id"`
		Java *int   `json:"java"`
	} `json:"version"`
}

func mcjarsType(flavor string) (string, error) {
	if !softwareIdentifier.MatchString(flavor) {
		return "", fmt.Errorf("invalid software identifier")
	}
	switch strings.ToLower(flavor) {
	case "leaf", "leafmc":
		return "LEAF", nil
	case "paper", "purpur", "folia", "fabric", "forge", "neoforge", "vanilla", "velocity", "waterfall", "bungeecord":
		return strings.ToUpper(flavor), nil
	}
	return "", fmt.Errorf("unsupported MCJars software %q", flavor)
}
func mcjarsProxy(kind string) bool {
	return kind == "VELOCITY" || kind == "WATERFALL" || kind == "BUNGEECORD"
}
func mcjarsString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// ResolveMCJarsBuild uses the project's exact version, or the literal latest selector.
// Metadata has one shared 30-second deadline and at most two requests (no redirects).
func ResolveMCJarsBuild(ctx context.Context, flavor, version string) (*MCJarsBuild, error) {
	kind, err := mcjarsType(flavor)
	if err != nil {
		return nil, err
	}
	if !softwareIdentifier.MatchString(version) {
		return nil, fmt.Errorf("invalid version identifier")
	}
	if strings.EqualFold(version, "latest") {
		version = "latest"
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	var payload mcjarsResponse
	if err = mcjarsJSON(ctx, "types/"+kind+"/versions/"+version+"/latest", &payload); err != nil {
		return nil, err
	}
	build, err := validateMCJars(payload, kind, version)
	if err != nil {
		return nil, err
	}
	if payload.Build.Java == nil {
		var detail mcjarsResponse
		if err = mcjarsJSON(ctx, payload.Build.UUID, &detail); err != nil {
			return nil, err
		}
		checked, err := validateMCJars(detail, kind, version)
		if err != nil {
			return nil, err
		}
		if checked.Info.BuildID != build.Info.BuildID || checked.Info.MinecraftVersion != build.Info.MinecraftVersion || checked.Info.ProjectVersion != build.Info.ProjectVersion || checked.download != build.download || checked.prepared != build.prepared {
			return nil, fmt.Errorf("MCJars detail build identity mismatch")
		}
		build = checked
	}
	return build, nil
}

// LookupMCJarsBuild uses the documented exact SHA256 route, not the latest build.
// SHA256 here is provider lookup evidence, not a locally verified download digest.
func LookupMCJarsBuild(ctx context.Context, sha256 string) (*MCJarsBuild, error) {
	if !mcjarsSHA256.MatchString(sha256) {
		return nil, fmt.Errorf("invalid SHA256")
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	var payload mcjarsResponse
	if err := mcjarsJSON(ctx, strings.ToLower(sha256), &payload); err != nil {
		return nil, err
	}
	build, err := validateMCJars(payload, "", "latest")
	if err != nil {
		return nil, err
	}
	if h, expected := mcjarsPathHash(build.download.URL); !build.prepared && h != nil && h.Size() == 32 && expected != strings.ToLower(sha256) {
		return nil, fmt.Errorf("MCJars hash lookup contradicts official artifact hash")
	}
	build.Info.SHA256 = strings.ToLower(sha256)
	return build, nil
}

func validateMCJars(p mcjarsResponse, kind, version string) (*MCJarsBuild, error) {
	b := p.Build
	if b == nil || !mcjarsUUID.MatchString(b.UUID) {
		return nil, fmt.Errorf("MCJars missing or invalid build")
	}
	canonical, err := mcjarsType(b.Type)
	if err != nil || canonical != b.Type || (kind != "" && b.Type != kind) {
		return nil, fmt.Errorf("MCJars engine mismatch")
	}
	mc, project := mcjarsString(b.Version), mcjarsString(b.Project)
	for _, id := range []*string{b.Version, b.Project} {
		if id != nil && (!softwareIdentifier.MatchString(*id) || strings.EqualFold(*id, "latest")) {
			return nil, fmt.Errorf("MCJars invalid returned version")
		}
	}
	selected := mc
	if mcjarsProxy(b.Type) {
		if b.Version != nil {
			return nil, fmt.Errorf("MCJars proxy must not have a Minecraft version")
		}
		selected = project
	} else if mc == "" {
		return nil, fmt.Errorf("MCJars missing Minecraft version")
	}
	if version != "latest" && selected != version {
		return nil, fmt.Errorf("MCJars version mismatch: requested %s, got %s", version, selected)
	}
	if len(b.Name) > 256 || strings.ContainsAny(b.Name, "\r\n\x00") {
		return nil, fmt.Errorf("MCJars invalid build name")
	}
	result := &MCJarsBuild{Info: serverjar.Info{Status: "identified", Source: "mcjars", Software: strings.ToLower(b.Type), MinecraftVersion: mc, ProjectVersion: project, BuildID: b.UUID, BuildName: b.Name}}
	java, source := b.Java, "mcjars-build"
	if java == nil && p.Version != nil {
		// The version object is a Minecraft version, not a proxy/loader version.
		if mc == "" || p.Version.ID != mc {
			return nil, fmt.Errorf("MCJars Java version identity mismatch")
		}
		java, source = p.Version.Java, "mcjars-version"
	}
	if java != nil {
		switch *java {
		case 7, 8, 11, 16, 17, 21, 25:
		default:
			return nil, fmt.Errorf("MCJars unsupported Java requirement %d", *java)
		}
		result.Info.RequiredJava, result.Info.JavaSource = *java, source
	}
	steps := b.Installation
	switch {
	case len(steps) == 1 && len(steps[0]) == 1 && steps[0][0].File == "server.jar":
		result.download = steps[0][0]
	case (b.Type == "FORGE" || b.Type == "NEOFORGE") && len(steps) == 3 && len(steps[0]) == 2 && len(steps[1]) == 1 && len(steps[2]) == 1:
		result.download = steps[0][0]
		const archive = "mcvapi.server.jar.zip"
		if result.download.File != archive || steps[0][1] != (mcjarsStep{Type: "remove", Location: "libraries"}) || steps[1][0] != (mcjarsStep{Type: "unzip", File: archive, Location: "."}) || steps[2][0] != (mcjarsStep{Type: "remove", Location: archive}) {
			return nil, fmt.Errorf("unsupported MCJars prepared installation shape")
		}
		result.prepared = true
	default:
		return nil, fmt.Errorf("unsupported MCJars installation shape")
	}
	d := result.download
	if d.Type != "download" || d.Location != "" || d.Size <= 0 || d.Size > mcjarsBinaryLimit {
		return nil, fmt.Errorf("invalid MCJars download size or step")
	}
	if _, err := mcjarsDownloadURL(d.URL); err != nil {
		return nil, err
	}
	return result, nil
}

func mcjarsJSON(ctx context.Context, route string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, mcjarsAPI+route, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", userAgent)
	client := mcjarsClient(false)
	defer client.CloseIdleConnections()
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return ErrMCJarsNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("MCJars metadata HTTP %d", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, (4<<20)+1))
	if err != nil {
		return err
	}
	if len(data) > 4<<20 {
		return fmt.Errorf("MCJars metadata exceeds 4 MiB")
	}
	return json.Unmarshal(data, out)
}

// Exact hosts only: no wildcard suffix grants and no arbitrary API-provided origins.
func mcjarsDownloadURL(raw string) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}
	if u.Scheme != "https" || u.User != nil || u.Opaque != "" || u.Fragment != "" || (u.Port() != "" && u.Port() != "443") {
		return nil, fmt.Errorf("unsafe MCJars download URL")
	}
	switch u.Hostname() {
	case "mcjars.app", "files.mcjars.app", "versions.mcjars.app", "cdn.mcjars.app",
		"fill-data.papermc.io", "fill.papermc.io", "api.papermc.io",
		"api.purpurmc.org", "download.purpurmc.org", "meta.fabricmc.net", "maven.fabricmc.net",
		"piston-data.mojang.com", "launcher.mojang.com", "launchermeta.mojang.com",
		"maven.minecraftforge.net", "maven.neoforged.net", "ci.md-5.net",
		"api.leafmc.one", "download.leafmc.one", "github.com", "objects.githubusercontent.com", "release-assets.githubusercontent.com":
		return u, nil
	}
	return nil, fmt.Errorf("untrusted MCJars download host %q", u.Hostname())
}

func mcjarsPublicIP(ip net.IP) bool {
	if !ip.IsGlobalUnicast() || ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsUnspecified() {
		return false
	}
	if v4 := ip.To4(); v4 != nil {
		return v4[0] != 0 && !(v4[0] == 100 && v4[1] >= 64 && v4[1] <= 127) && !(v4[0] == 198 && (v4[1] == 18 || v4[1] == 19))
	}
	return true
}

// Pin each connection to a checked public DNS result; disabling environment proxies
// prevents proxy-side DNS from bypassing the check. Custom transports support tests.
func mcjarsClient(binary bool) *http.Client {
	c := *httpClient
	c.Timeout = 30 * time.Second
	if binary {
		c.Timeout = 10 * time.Minute
	}
	transport := c.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}
	if base, ok := transport.(*http.Transport); ok {
		tr := base.Clone()
		tr.Proxy = nil
		tr.DialTLSContext = nil
		tr.DialTLS = nil
		tr.DialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(address)
			if err != nil {
				return nil, err
			}
			ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
			if err != nil {
				return nil, err
			}
			if len(ips) == 0 {
				return nil, fmt.Errorf("empty DNS result")
			}
			for _, addr := range ips {
				if !mcjarsPublicIP(addr.IP) {
					return nil, fmt.Errorf("non-public download address")
				}
			}
			return (&net.Dialer{Timeout: 15 * time.Second}).DialContext(ctx, network, net.JoinHostPort(ips[0].IP.String(), port))
		}
		c.Transport = tr
	}
	c.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if !binary {
			return fmt.Errorf("MCJars metadata redirects are forbidden")
		}
		if len(via) >= 5 {
			return fmt.Errorf("too many download redirects")
		}
		_, err := mcjarsDownloadURL(req.URL.String())
		return err
	}
	return &c
}
