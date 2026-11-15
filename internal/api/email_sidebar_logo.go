package api

import (
	"context"
	"encoding/base64"
	"fmt"
	"html"
	"io"
	"io/fs"
	"net/url"
	"os"
	"path"
	"strings"
	"time"
)

func (h *EmailHandler) prepareEmailLogo(body string) (string, []byte) {
	body, logo, _ := h.prepareSidebarEmailLogo(body)
	return body, logo
}

// The marker carries the exact saved sidebar source, not a favicon or an image
// URL reconstructed from the email link. Remove source metadata before delivery.
func (h *EmailHandler) prepareSidebarEmailLogo(body string) (string, []byte, bool) {
	match := emailLogoCell.FindStringSubmatch(body)
	if match == nil {
		return body, nil, false
	}
	source := html.UnescapeString(match[1])
	var data []byte
	var err error
	if source != "" {
		data, err = h.loadEmailLogo(source)
	}
	warning := source != "" && err != nil
	if len(data) == 0 {
		data, err = sidebarCubePNG(html.UnescapeString(match[2]))
	}
	if err != nil {
		return strings.Replace(body, match[0], "", 1), nil, true
	}
	cell := strings.Replace(match[0], `src="`+match[3]+`"`, `src="cid:`+emailLogoCID+`"`, 1)
	for _, attr := range []string{` data-neopanel-logo="true"`, ` data-logo-source="` + match[1] + `"`, ` data-logo-accent="` + match[2] + `"`} {
		cell = strings.Replace(cell, attr, "", 1)
	}
	return strings.Replace(body, match[0], cell, 1), data, warning
}

func sidebarCubePNG(accent string) ([]byte, error) {
	// Same square/diamond, corner radius and accent as .brand .cube in the UI.
	accent = emailColor(accent, "#387aff")
	svg := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 22 22"><rect width="22" height="22" rx="6" fill="%s"/><rect x="5" y="5" width="12" height="12" rx="2" fill="none" stroke="#ffffff" stroke-opacity="0.85" stroke-width="1.5" transform="translate(11,11) rotate(45) scale(0.78,0.78) translate(-11,-11)"/></svg>`, accent)
	return emailLogoPNG([]byte(svg), true)
}

func (h *EmailHandler) loadEmailLogo(raw string) ([]byte, error) {
	raw = strings.TrimSpace(raw)
	if strings.HasPrefix(strings.ToLower(raw), "data:image/") {
		data, svg, err := emailLogoDataURL(raw)
		if err != nil {
			return nil, err
		}
		return emailLogoPNG(data, svg)
	}
	if raw == "" || len(raw) > 4096 || strings.ContainsAny(raw, "\\\r\n\x00") {
		return nil, errEmailLogo
	}
	u, err := url.Parse(raw)
	if err != nil || u.User != nil || u.Opaque != "" || (u.Scheme != "" && u.Scheme != "http" && u.Scheme != "https") {
		return nil, errEmailLogo
	}
	origin := ""
	if h.auth != nil {
		if settings, e := h.auth.panelURL(); e == nil {
			origin = settings.PublicURL
		}
	} else if h.cfg != nil {
		origin = h.cfg.Server.PublicURL
	}
	base, _ := url.Parse(origin)
	local := u.Scheme == "" && u.Host == ""
	if base != nil && strings.EqualFold(u.Host, base.Host) && (u.Scheme == "" || strings.EqualFold(u.Scheme, base.Scheme)) {
		local = true
	}
	var data []byte
	if local {
		name := strings.TrimPrefix(u.Path, "/")
		if !fs.ValidPath(name) {
			return nil, errEmailLogo
		}
		var file fs.File
		if strings.HasPrefix(name, "branding/") {
			name = strings.TrimPrefix(name, "branding/")
			if !safeAssetRe.MatchString(name) || strings.ContainsAny(name, `/\`) {
				return nil, errEmailLogo
			}
			root, e := os.OpenRoot((&BrandingHandler{cfg: h.cfg}).brandingDir())
			if e != nil {
				return nil, errEmailLogo
			}
			defer root.Close()
			file, err = root.Open(name)
		} else if h.assets != nil {
			file, err = h.assets.Open(name)
			// The public static handler accepts both /asset and /static/asset.
			if err != nil && strings.HasPrefix(name, "static/") {
				file, err = h.assets.Open(strings.TrimPrefix(name, "static/"))
			}
		} else {
			return nil, errEmailLogo
		}
		if err != nil {
			return nil, errEmailLogo
		}
		defer file.Close()
		info, e := file.Stat()
		if e != nil || !info.Mode().IsRegular() || info.Size() > emailLogoMaxBytes {
			return nil, errEmailLogo
		}
		data, err = io.ReadAll(io.LimitReader(file, emailLogoMaxBytes+1))
	} else {
		resolved := emailLogoURL(raw, origin)
		u, err = url.Parse(resolved)
		if err != nil || u.Hostname() == "" {
			return nil, errEmailLogo
		}
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		data, err = fetchEmailLogo(ctx, u)
	}
	if err != nil || len(data) == 0 || len(data) > emailLogoMaxBytes {
		return nil, errEmailLogo
	}
	return emailLogoPNG(data, strings.EqualFold(path.Ext(u.Path), ".svg"))
}

func emailLogoDataURL(raw string) ([]byte, bool, error) {
	if len(raw) > emailLogoMaxBytes*4/3+1024 {
		return nil, false, errEmailLogo
	}
	header, payload, ok := strings.Cut(raw, ",")
	if !ok {
		return nil, false, errEmailLogo
	}
	fields := strings.Split(strings.ToLower(header), ";")
	kind := strings.TrimPrefix(fields[0], "data:")
	switch kind {
	case "image/png", "image/jpeg", "image/jpg", "image/gif", "image/webp", "image/svg+xml", "image/x-icon", "image/vnd.microsoft.icon":
	default:
		return nil, false, errEmailLogo
	}
	encoded := false
	for _, field := range fields[1:] {
		switch field {
		case "base64":
			encoded = true
		case "utf8", "charset=utf-8":
		default:
			return nil, false, errEmailLogo
		}
	}
	var data []byte
	var err error
	if encoded {
		data, err = base64.StdEncoding.DecodeString(payload)
	} else {
		var text string
		text, err = url.PathUnescape(payload)
		data = []byte(text)
	}
	if err != nil || len(data) > emailLogoMaxBytes {
		return nil, false, errEmailLogo
	}
	return data, kind == "image/svg+xml", nil
}
