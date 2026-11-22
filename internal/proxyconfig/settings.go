// Package proxyconfig edits real proxy configuration files. Saves preserve unknown
// semantic settings, but serializers may reformat files and discard comments.
// Callers must serialize operations with their runner operation gate. The package
// also serializes its own calls; revision checks are not a cross-process lock.
package proxyconfig

import (
	"errors"
	"net"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

var (
	ErrConflict = errors.New("proxy configuration changed; reload before saving")
	ErrMissing  = errors.New("proxy configuration is missing; initialize it first")
	errConfig   = errors.New("Invalid proxy configuration; edit the configuration in Files")
	errSecret   = errors.New("Invalid or unsafe forwarding secret file")
)

type Backend struct {
	Name       string `json:"name"`
	Address    string `json:"address"`
	MOTD       string `json:"motd"`
	Restricted bool   `json:"restricted"`
}
type ForcedHost struct {
	Host    string   `json:"host"`
	Servers []string `json:"servers"`
}
type Settings struct {
	MOTD                   string       `json:"motd"`
	MaxPlayers             int          `json:"max_players"`
	OnlineMode             bool         `json:"online_mode"`
	ForwardingMode         string       `json:"forwarding_mode"`
	IPForward              bool         `json:"ip_forward"`
	ForceKeyAuthentication bool         `json:"force_key_authentication"`
	ForceDefaultServer     bool         `json:"force_default_server"`
	PingPassthrough        string       `json:"ping_passthrough"`
	Backends               []Backend    `json:"backends"`
	Try                    []string     `json:"try"`
	ForcedHosts            []ForcedHost `json:"forced_hosts"`
}
type Snapshot struct {
	Type     string   `json:"type"`
	File     string   `json:"file"`
	Exists   bool     `json:"exists"`
	Revision string   `json:"revision"`
	Config   Settings `json:"config"`
}

func Supported(kind string) bool { return FileName(kind) != "" }
func FileName(kind string) string {
	switch strings.ToLower(kind) {
	case "velocity":
		return "velocity.toml"
	case "bungeecord", "waterfall":
		return "config.yml"
	}
	return ""
}
func defaults(kind string) Settings {
	s := Settings{MOTD: "<aqua>A Minecraft Network", MaxPlayers: 500, OnlineMode: true, ForwardingMode: "MODERN", ForceKeyAuthentication: true, PingPassthrough: "DISABLED", Backends: []Backend{}, Try: []string{}, ForcedHosts: []ForcedHost{}}
	if kind != "velocity" {
		s.MOTD = "A Minecraft Network"
		s.MaxPlayers = 100
		s.ForwardingMode = ""
		s.ForceKeyAuthentication = false
		s.IPForward = true
	}
	return s
}

var safeName = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)
var dnsLabel = regexp.MustCompile(`^[A-Za-z0-9](?:[A-Za-z0-9-]{0,61}[A-Za-z0-9])?$`)

func validHost(s string) bool {
	if net.ParseIP(s) != nil {
		return true
	}
	if len(s) == 0 || len(s) > 253 {
		return false
	}
	s = strings.TrimSuffix(s, ".")
	for _, p := range strings.Split(s, ".") {
		if !dnsLabel.MatchString(p) {
			return false
		}
	}
	return true
}
func validPort(port int) bool { return port > 0 && port <= 65535 }
func address(s string) (string, bool) {
	host, p, err := net.SplitHostPort(s)
	if err != nil || !validHost(host) || p == "" {
		return "", false
	}
	for _, c := range p {
		if c < '0' || c > '9' {
			return "", false
		}
	}
	n, err := strconv.Atoi(p)
	return host, err == nil && validPort(n)
}
func validMOTD(s string) bool                { return len(s) <= 4096 && utf8.ValidString(s) }
func Validate(kind string, s Settings) error { return validate(strings.ToLower(kind), s, true) }
func validate(kind string, s Settings, usable bool) error {
	bad := func(s string) error { return errors.New(s) }
	if !Supported(kind) {
		return bad("Unsupported proxy type")
	}
	if !validMOTD(s.MOTD) || s.MaxPlayers < 0 || int64(s.MaxPlayers) > 2147483647 {
		return bad("Invalid MOTD or advertised player count")
	}
	if kind == "velocity" {
		switch s.ForwardingMode {
		case "NONE", "LEGACY", "MODERN", "BUNGEEGUARD":
		default:
			return bad("Invalid forwarding mode")
		}
		switch s.PingPassthrough {
		case "DISABLED", "MODS", "DESCRIPTION", "ALL":
		default:
			return bad("Invalid ping passthrough mode")
		}
		if s.IPForward || s.ForceDefaultServer {
			return bad("Bungee-only settings are not supported by Velocity")
		}
	} else {
		if s.ForwardingMode != "" || s.ForceKeyAuthentication {
			return bad("Velocity-only settings are not supported by BungeeCord")
		}
		if s.PingPassthrough != "DISABLED" && s.PingPassthrough != "ALL" {
			return bad("Invalid ping passthrough mode")
		}
	}
	if len(s.Backends) > 128 || len(s.Try) > 128 || len(s.ForcedHosts) > 128 {
		return bad("Proxy configuration exceeds 128-entry limit")
	}
	if usable && (len(s.Backends) == 0 || len(s.Try) == 0) {
		return bad("Configure at least one backend and fallback server before saving")
	}
	names, folded := map[string]bool{}, map[string]bool{}
	for _, b := range s.Backends {
		if !safeName.MatchString(b.Name) || strings.EqualFold(b.Name, "try") || folded[strings.ToLower(b.Name)] {
			return bad("Invalid or duplicate backend name")
		}
		if _, ok := address(b.Address); !ok {
			return bad("Backend address must be a valid host:port")
		}
		if !validMOTD(b.MOTD) {
			return bad("Invalid backend MOTD")
		}
		names[b.Name] = true
		folded[strings.ToLower(b.Name)] = true
	}
	refs := func(r []string) bool {
		if len(r) > 128 {
			return false
		}
		for _, n := range r {
			if !names[n] {
				return false
			}
		}
		return true
	}
	if !refs(s.Try) {
		return bad("Fallback references an unknown backend or exceeds the limit")
	}
	hosts := map[string]bool{}
	for _, h := range s.ForcedHosts {
		key := strings.ToLower(h.Host)
		if !validHost(h.Host) || hosts[key] {
			return bad("Invalid or duplicate forced host")
		}
		hosts[key] = true
		if len(h.Servers) == 0 || !refs(h.Servers) {
			return bad("Forced host requires valid backend references")
		}
		if kind != "velocity" && len(h.Servers) != 1 {
			return bad("BungeeCord forced hosts require exactly one backend")
		}
	}
	return nil
}
