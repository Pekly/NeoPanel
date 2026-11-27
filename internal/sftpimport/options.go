// Package sftpimport downloads an SFTP tree into an unpublished private staging
// directory. It never modifies the source. SFTP cannot provide a snapshot: stop
// the source before importing. Stat and size checks detect mutations best effort,
// but cannot detect every replacement or a dishonest server. The caller owns
// staging cleanup on failure and must keep the staging directory private until
// the transfer succeeds; this package never publishes or overwrites files.
package sftpimport

import (
	"encoding/base64"
	"errors"
	"net/netip"
	"path"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/crypto/ssh"
)

type Options struct {
	Protocol       string   `json:"protocol,omitempty"` // "sftp" (default), "ftp", or "ftps"
	Host           string   `json:"host"`
	Port           int      `json:"port"`
	Username       string   `json:"username"`
	Password       string   `json:"password"`
	PrivateKey     string   `json:"private_key"`
	Passphrase     string   `json:"passphrase"`
	Fingerprint    string   `json:"fingerprint"`
	RemotePath     string   `json:"remote_path"`
	IncludePaths   []string `json:"include_paths"`
	ExcludePaths   []string `json:"exclude_paths"`
	SkipHidden     bool     `json:"skip_hidden"`
	PreserveTimes  bool     `json:"preserve_times"`
	ConflictMode   string   `json:"conflict_mode,omitempty"` // "", "empty", "preserve", or "replace"
	MaxBytes       int64    `json:"max_bytes"`
	BandwidthKiB   int      `json:"bandwidth_kib"`
	TimeoutMinutes int      `json:"timeout_minutes"`
}

type HostKeyInfo struct {
	Fingerprint string `json:"fingerprint"`
	Algorithm   string `json:"algorithm"`
}
type Entry struct {
	Name  string `json:"name"`
	Path  string `json:"path"`
	IsDir bool   `json:"is_dir"`
	Size  int64  `json:"size"`
}
type Preview struct {
	RemotePath string  `json:"remote_path"`
	Entries    []Entry `json:"entries"`
	Truncated  bool    `json:"truncated"`
}
type Progress struct {
	State       string `json:"state"`
	CurrentPath string `json:"current_path"`
	FilesTotal  int    `json:"files_total"`
	FilesDone   int    `json:"files_done"`
	BytesTotal  int64  `json:"bytes_total"`
	BytesDone   int64  `json:"bytes_done"`
	SpeedBPS    int64  `json:"speed_bps"`
}

func problem(s string) error { return errors.New("sftp import: " + s) }

// Validate returns a normalized copy. Port zero defaults to 22 for SFTP or 21 for FTP/FTPS.
func Validate(o Options) (Options, error) {
	o.Protocol = strings.ToLower(strings.TrimSpace(o.Protocol))
	if o.Protocol == "" {
		o.Protocol = "sftp"
	}
	if o.Protocol != "sftp" && o.Protocol != "ftp" && o.Protocol != "ftps" {
		return o, problem("unsupported protocol")
	}
	if o.Port == 0 {
		if o.Protocol == "ftp" || o.Protocol == "ftps" {
			o.Port = 21
		} else {
			o.Port = 22
		}
	}
	var err error
	o.Host, o.Port, err = hostPort(o.Host, o.Port)
	if err != nil {
		return o, err
	}
	if o.Username == "" || len(o.Username) > 256 || strings.IndexFunc(o.Username, unicode.IsControl) >= 0 {
		return o, problem("invalid username")
	}
	if len(o.Password) > 4096 || len(o.Passphrase) > 4096 || len(o.PrivateKey) > 65536 {
		return o, problem("credential too long")
	}
	if o.Protocol == "ftp" || o.Protocol == "ftps" {
		if o.Password == "" || o.PrivateKey != "" || o.Passphrase != "" {
			return o, problem("password authentication required for FTP/FTPS")
		}
	} else {
		if (o.Password == "") == (o.PrivateKey == "") || (o.PrivateKey == "" && o.Passphrase != "") {
			return o, problem("provide exactly one password or private key")
		}
		if o.PrivateKey != "" {
			if _, err := signer(o); err != nil {
				return o, problem("invalid private key or passphrase")
			}
		}
		if len(o.Fingerprint) != 50 {
			return o, problem("valid SHA256 host fingerprint required")
		}
		b, err := base64.RawStdEncoding.DecodeString(strings.TrimPrefix(o.Fingerprint, "SHA256:"))
		if !strings.HasPrefix(o.Fingerprint, "SHA256:") || err != nil || len(b) != 32 || "SHA256:"+base64.RawStdEncoding.EncodeToString(b) != o.Fingerprint {
			return o, problem("valid SHA256 host fingerprint required")
		}
	}
	if o.RemotePath == "" {
		o.RemotePath = "."
	}
	if !safePath(o.RemotePath, true) {
		return o, problem("invalid remote path")
	}
	o.RemotePath = path.Clean(o.RemotePath)
	if len(o.IncludePaths) > 128 || len(o.ExcludePaths) > 128 {
		return o, problem("too many path filters")
	}
	o.IncludePaths = append([]string(nil), o.IncludePaths...)
	o.ExcludePaths = append([]string(nil), o.ExcludePaths...)
	for _, list := range [][]string{o.IncludePaths, o.ExcludePaths} {
		for i, p := range list {
			if !safePath(p, false) {
				return o, problem("invalid path filter")
			}
			list[i] = path.Clean(p)
		}
	}
	o.ConflictMode = strings.ToLower(strings.TrimSpace(o.ConflictMode))
	switch o.ConflictMode {
	case "", "empty", "preserve", "replace":
	default:
		return o, problem("invalid conflict mode")
	}
	if o.MaxBytes == 0 {
		o.MaxBytes = 10 << 30
	}
	if o.MaxBytes < 1 || o.MaxBytes > 100<<30 {
		return o, problem("invalid byte limit")
	}
	if o.BandwidthKiB < 0 || o.BandwidthKiB > 1048576 {
		return o, problem("invalid bandwidth limit")
	}
	if o.TimeoutMinutes == 0 {
		o.TimeoutMinutes = 60
	}
	if o.TimeoutMinutes < 1 || o.TimeoutMinutes > 240 {
		return o, problem("invalid timeout")
	}
	return o, nil
}
func signer(o Options) (ssh.Signer, error) {
	if o.Passphrase != "" {
		return ssh.ParsePrivateKeyWithPassphrase([]byte(o.PrivateKey), []byte(o.Passphrase))
	}
	return ssh.ParsePrivateKey([]byte(o.PrivateKey))
}
func hostPort(host string, port int) (string, int, error) {
	if len(host) > 253 {
		return "", 0, problem("invalid host")
	}
	host = strings.ToLower(strings.TrimSpace(host))
	if port == 0 {
		port = 22
	}
	if len(host) == 0 || len(host) > 253 || port < 1 || port > 65535 || strings.ContainsAny(host, "/%\\@[]") {
		return "", 0, problem("invalid host or port")
	}
	if ip, e := netip.ParseAddr(host); e == nil {
		if ip.Zone() != "" {
			return "", 0, problem("invalid host")
		}
		return ip.String(), port, nil
	}
	for _, label := range strings.Split(host, ".") {
		if len(label) == 0 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return "", 0, problem("invalid host")
		}
		for _, c := range label {
			if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '-') {
				return "", 0, problem("invalid host")
			}
		}
	}
	return host, port, nil
}
func safeComponent(s string) bool {
	if s == "" || s == "." || s == ".." || len(s) > 255 || !utf8.ValidString(s) || strings.ContainsAny(s, "/\\:<>\"|?*") || strings.HasSuffix(s, ".") || strings.HasSuffix(s, " ") || strings.IndexFunc(s, unicode.IsControl) >= 0 {
		return false
	}
	base := strings.ToUpper(strings.TrimRight(strings.Split(s, ".")[0], " "))
	switch base {
	case "CON", "PRN", "AUX", "NUL", "CLOCK$", "CONIN$", "CONOUT$":
		return false
	}
	if strings.HasPrefix(base, "COM") || strings.HasPrefix(base, "LPT") {
		suffix := base[3:]
		if len(suffix) == 1 && suffix[0] >= '0' && suffix[0] <= '9' || suffix == "¹" || suffix == "²" || suffix == "³" {
			return false
		}
	}
	return true
}
func safePath(p string, remote bool) bool {
	if len(p) == 0 || len(p) > 4096 {
		return false
	}
	if remote {
		if p == "." || p == "/" {
			return true
		}
		p = strings.TrimPrefix(p, "/")
	} else if strings.HasPrefix(p, "/") {
		return false
	}
	p = strings.TrimSuffix(p, "/")
	for _, s := range strings.Split(p, "/") {
		if !safeComponent(s) {
			return false
		}
	}
	return true
}
