package api

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"html"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"io"
	"math"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
)

const emailLogoMaxBytes = 2 << 20
const emailLogoCID = "neopanel-logo"

var emailLogoCell = regexp.MustCompile(`<td data-neopanel-logo="true"[^>]*><img src="([^"]+)"[^>]*></td>`)
var errEmailLogo = errors.New("email logo unavailable")

var emailSVGTransform = regexp.MustCompile(`(?i)\b(?:transform|gradientTransform)\s*=\s*(?:"[^"]*"|'[^']*')`)
var emailSVGSingleScale = regexp.MustCompile(`(?i)\bscale\(\s*([-+]?(?:\d+\.?\d*|\.\d+)(?:e[-+]?\d+)?)\s*\)`)
var emailSVGNumber = regexp.MustCompile(`[-+]?(?:\d+\.?\d*|\.\d+)(?:[eE][-+]?\d+)?`)

func normalizeEmailSVGScale(data []byte) []byte {
	// oksvg's one-argument scale applies a zero Y scale. Expand SVG's implicit
	// second argument explicitly, only inside transform attributes.
	return emailSVGTransform.ReplaceAllFunc(data, func(attr []byte) []byte {
		return emailSVGSingleScale.ReplaceAll(attr, []byte(`scale($1,$1)`))
	})
}

// Only the fixed template's marked brand image is replaced. Invalid or missing
// assets leave the visible brand name, not a broken-image placeholder.
func (h *EmailHandler) prepareEmailLogo(body string) (string, []byte) {
	match := emailLogoCell.FindStringSubmatch(body)
	if match == nil {
		return body, nil
	}
	data, err := h.loadEmailLogo(html.UnescapeString(match[1]))
	if err != nil {
		return strings.Replace(body, match[0], "", 1), nil
	}
	cell := strings.Replace(match[0], `src="`+match[1]+`"`, `src="cid:`+emailLogoCID+`"`, 1)
	cell = strings.Replace(cell, ` data-neopanel-logo="true"`, "", 1)
	return strings.Replace(body, match[0], cell, 1), data
}

func (h *EmailHandler) loadEmailLogo(raw string) ([]byte, error) {
	u, err := url.Parse(emailLogoURL(raw, ""))
	if err != nil || u.Hostname() == "" {
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
	var data []byte
	if base != nil && strings.EqualFold(u.Scheme, base.Scheme) && strings.EqualFold(u.Host, base.Host) && strings.HasPrefix(u.Path, "/branding/") {
		name := strings.TrimPrefix(u.Path, "/branding/")
		if !safeAssetRe.MatchString(name) || strings.ContainsAny(name, `/\`) {
			return nil, errEmailLogo
		}
		dir := (&BrandingHandler{cfg: h.cfg}).brandingDir()
		root, e := os.OpenRoot(dir)
		if e != nil {
			return nil, errEmailLogo
		}
		defer root.Close()
		file, e := root.Open(name)
		if e != nil {
			return nil, errEmailLogo
		}
		defer file.Close()
		info, e := file.Stat()
		if e != nil || !info.Mode().IsRegular() || info.Size() > emailLogoMaxBytes {
			return nil, errEmailLogo
		}
		data, err = io.ReadAll(io.LimitReader(file, emailLogoMaxBytes+1))
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		data, err = fetchEmailLogo(ctx, u)
	}
	if err != nil || len(data) == 0 || len(data) > emailLogoMaxBytes {
		return nil, errEmailLogo
	}
	return emailLogoPNG(data, strings.EqualFold(filepath.Ext(u.Path), ".svg"))
}

// Never use environment proxies, credentials, redirects, or private/link-local
// addresses for administrator-selected remote logos. Dial the checked IP itself
// to prevent DNS rebinding between validation and connection.
func publicEmailLogoIP(ip netip.Addr) bool {
	ip = ip.Unmap()
	if !ip.IsGlobalUnicast() || ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() {
		return false
	}
	for _, block := range []string{"0.0.0.0/8", "100.64.0.0/10", "192.0.0.0/24", "192.0.2.0/24", "198.18.0.0/15", "198.51.100.0/24", "203.0.113.0/24", "240.0.0.0/4", "2001:db8::/32", "64:ff9b::/96", "64:ff9b:1::/48", "2002::/16"} {
		if netip.MustParsePrefix(block).Contains(ip) {
			return false
		}
	}
	return true
}
func fetchEmailLogo(ctx context.Context, u *url.URL) ([]byte, error) {
	if u.Scheme != "https" || u.User != nil || u.Hostname() == "" || (u.Port() != "" && u.Port() != "443") {
		return nil, errEmailLogo
	}
	transport := &http.Transport{TLSHandshakeTimeout: 2 * time.Second, MaxResponseHeaderBytes: 32 << 10, DisableKeepAlives: true,
		DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(address)
			if err != nil {
				return nil, errEmailLogo
			}
			ips, err := net.DefaultResolver.LookupNetIP(ctx, "ip", host)
			if err != nil || len(ips) == 0 {
				return nil, errEmailLogo
			}
			for _, ip := range ips {
				if !publicEmailLogoIP(ip) {
					return nil, errEmailLogo
				}
			}
			return (&net.Dialer{Timeout: 2 * time.Second}).DialContext(ctx, network, net.JoinHostPort(ips[0].String(), port))
		},
	}
	defer transport.CloseIdleConnections()
	client := &http.Client{Transport: transport, CheckRedirect: func(*http.Request, []*http.Request) error { return errEmailLogo }}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, errEmailLogo
	}
	req.Header.Set("Accept", "image/png,image/jpeg,image/gif,image/webp,image/svg+xml")
	response, err := client.Do(req)
	if err != nil {
		return nil, errEmailLogo
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK || response.ContentLength > emailLogoMaxBytes {
		return nil, errEmailLogo
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, emailLogoMaxBytes+1))
	if err != nil || len(data) > emailLogoMaxBytes {
		return nil, errEmailLogo
	}
	return data, nil
}

// Rasterize to a small square PNG: email clients need neither SVG/WebP support
// nor object-fit. Bounds are checked before allocating the decoded raster.
func emailLogoPNG(data []byte, svg bool) (out []byte, err error) {
	defer func() {
		if recover() != nil {
			out = nil
			err = errEmailLogo
		}
	}()
	if len(data) == 0 || len(data) > emailLogoMaxBytes {
		return nil, errEmailLogo
	}
	const size = 128
	dst := image.NewRGBA(image.Rect(0, 0, size, size))
	if svg || bytes.Contains(bytes.TrimSpace(data[:min(len(data), 512)]), []byte("<svg")) {
		if !safeEmailSVG(data) {
			return nil, errEmailLogo
		}
		icon, e := oksvg.ReadIconStream(bytes.NewReader(normalizeEmailSVGScale(data)), oksvg.StrictErrorMode)
		if e != nil {
			return nil, errEmailLogo
		}
		w, h := icon.ViewBox.W, icon.ViewBox.H
		if w <= 0 || h <= 0 || w > 8192 || h > 8192 || math.IsNaN(w) || math.IsNaN(h) || len(icon.SVGPaths) > 512 {
			return nil, errEmailLogo
		}
		scale := float64(size) / math.Max(w, h)
		icon.SetTarget((size-w*scale)/2, (size-h*scale)/2, w*scale, h*scale)
		icon.Draw(rasterx.NewDasher(size, size, rasterx.NewScannerGV(size, size, dst, dst.Bounds())), 1)
	} else {
		info, _, e := image.DecodeConfig(bytes.NewReader(data))
		if e != nil || info.Width <= 0 || info.Height <= 0 || info.Width > 8192 || info.Height > 8192 || int64(info.Width)*int64(info.Height) > 16<<20 {
			return nil, errEmailLogo
		}
		src, _, e := image.Decode(bytes.NewReader(data))
		if e != nil {
			return nil, errEmailLogo
		}
		scale := float64(size) / float64(max(info.Width, info.Height))
		w, h := max(1, int(float64(info.Width)*scale)), max(1, int(float64(info.Height)*scale))
		bounds := image.Rect((size-w)/2, (size-h)/2, (size-w)/2+w, (size-h)/2+h)
		draw.CatmullRom.Scale(dst, bounds, src, src.Bounds(), draw.Over, nil)
	}
	visible := false
	for i := 3; i < len(dst.Pix); i += 4 {
		if dst.Pix[i] != 0 {
			visible = true
			break
		}
	}
	if !visible {
		return nil, errEmailLogo
	}
	var buf bytes.Buffer
	if e := png.Encode(&buf, dst); e != nil {
		return nil, e
	}
	return buf.Bytes(), nil
}

func safeEmailSVG(data []byte) bool {
	if len(data) > 128<<10 {
		return false
	}
	decoder := xml.NewDecoder(bytes.NewReader(data))
	count, depth := 0, 0
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			return count > 0 && depth == 0
		}
		if err != nil {
			return false
		}
		switch value := token.(type) {
		case xml.Directive:
			return false
		case xml.StartElement:
			count++
			depth++
			if count > 512 || depth > 32 {
				return false
			}
			switch strings.ToLower(value.Name.Local) {
			// CSS selector lists expand multiplicatively in oksvg; disallow style
			// elements before parsing rather than relying on post-parse limits.
			case "style", "use", "image", "script", "foreignobject", "filter", "pattern":
				return false
			}
			for _, attr := range value.Attr {
				if len(attr.Value) > 64<<10 {
					return false
				}
				if strings.EqualFold(attr.Name.Local, "href") && !strings.HasPrefix(attr.Value, "#") {
					return false
				}
				switch strings.ToLower(attr.Name.Local) {
				case "d", "points", "transform", "gradienttransform", "viewbox", "cx", "cy", "r", "rx", "ry", "x", "y", "x1", "x2", "y1", "y2", "width", "height", "stroke-width":
					lower := strings.ToLower(attr.Value)
					if strings.Contains(lower, "nan") || strings.Contains(lower, "inf") {
						return false
					}
					for _, number := range emailSVGNumber.FindAllString(attr.Value, -1) {
						value, err := strconv.ParseFloat(number, 64)
						if err != nil || math.Abs(value) > 1e6 {
							return false
						}
					}
				}
			}
		case xml.EndElement:
			depth--
		}
	}
}
