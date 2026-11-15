package api

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
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
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
)

const emailLogoMaxBytes = 12 << 20 // Matches uploaded branding assets.
const remoteEmailLogoMaxBytes = 2 << 20
const emailLogoCID = "neopanel-logo"

var emailLogoCell = regexp.MustCompile(`<td data-neopanel-logo="true" data-logo-source="([^"]*)" data-logo-accent="([^"]*)"[^>]*><img src="([^"]+)"[^>]*></td>`)
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

// Never use environment proxies, credentials, or private/link-local
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
	if !validEmailLogoRemoteURL(u) {
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
	client := emailLogoHTTPClient(transport)
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
	if response.StatusCode != http.StatusOK || response.ContentLength > remoteEmailLogoMaxBytes {
		return nil, errEmailLogo
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, remoteEmailLogoMaxBytes+1))
	if err != nil || len(data) > remoteEmailLogoMaxBytes {
		return nil, errEmailLogo
	}
	return data, nil
}

func emailLogoHTTPClient(transport http.RoundTripper) *http.Client {
	return &http.Client{Transport: transport, CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) == 0 || len(via) > 3 || !validEmailLogoRemoteURL(req.URL) || (via[0].URL.Scheme == "https" && req.URL.Scheme != "https") {
			return errEmailLogo
		}
		return nil // Production transport also validates and pins every DNS result.
	}}
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
	if len(data) >= 6 && bytes.Equal(data[:4], []byte{0, 0, 1, 0}) {
		data, err = emailICOToPNG(data)
		if err != nil {
			return nil, err
		}
		svg = false
	}
	const size = 128
	dst := image.NewRGBA(image.Rect(0, 0, size, size))
	// PNG/JPEG metadata can contain SVG thumbnails (for example C2PA data).
	// A recognized raster must win over filename hints or embedded text.
	info, _, rasterErr := image.DecodeConfig(bytes.NewReader(data))
	if rasterErr != nil && (svg || bytes.Contains(bytes.TrimSpace(data[:min(len(data), 512)]), []byte("<svg"))) {
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
		if rasterErr != nil || info.Width <= 0 || info.Height <= 0 || info.Width > 8192 || info.Height > 8192 || int64(info.Width)*int64(info.Height) > 16<<20 {
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
	count, depth, styleDepth, cssCost := 0, 0, 0, 0
	var css strings.Builder
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
			if styleDepth != 0 {
				return false
			}
			count++
			depth++
			if count > 512 || depth > 32 {
				return false
			}
			switch strings.ToLower(value.Name.Local) {
			case "use", "image", "script", "foreignobject", "filter", "pattern":
				return false
			case "style":
				styleDepth = depth
				css.Reset()
			}
			for _, attr := range value.Attr {
				if len(attr.Value) > 64<<10 {
					return false
				}
				if strings.EqualFold(attr.Name.Local, "href") && !strings.HasPrefix(attr.Value, "#") {
					return false
				}
				if strings.EqualFold(attr.Name.Local, "style") {
					cost, ok := emailSVGStyleCost(".inline{" + attr.Value + "}")
					if !ok {
						return false
					}
					cssCost += cost
				}
				switch strings.ToLower(attr.Name.Local) {
				case "d", "points", "transform", "gradienttransform", "viewbox", "cx", "cy", "r", "rx", "ry", "x", "y", "x1", "x2", "y1", "y2", "width", "height", "stroke-width":
					if !boundedSVGNumbers(attr.Value) {
						return false
					}
				}
			}
		case xml.CharData:
			if styleDepth != 0 {
				if css.Len()+len(value) > 8192 {
					return false
				}
				css.Write(value)
			}
		case xml.EndElement:
			if depth == styleDepth {
				cost, ok := emailSVGStyleCost(css.String())
				if !ok {
					return false
				}
				cssCost += cost
				styleDepth = 0
			}
			depth--
		}
		if cssCost > 4096 {
			return false
		}
	}
}

func boundedSVGNumbers(text string) bool {
	lower := strings.ToLower(text)
	if strings.Contains(lower, "nan") || strings.Contains(lower, "inf") {
		return false
	}
	for _, number := range emailSVGNumber.FindAllString(text, -1) {
		value, err := strconv.ParseFloat(number, 64)
		if err != nil || math.Abs(value) > 1e6 {
			return false
		}
	}
	return true
}

// Bound selector × declaration expansion before handing CSS to oksvg.
func emailSVGStyleCost(css string) (int, bool) {
	if len(css) > 8192 || strings.Contains(css, "@") {
		return 0, false
	}
	cost := 0
	for _, rule := range strings.Split(css, "}") {
		if strings.TrimSpace(rule) == "" {
			continue
		}
		selectors, declarations, ok := strings.Cut(rule, "{")
		if !ok {
			return 0, false
		}
		n, m := strings.Count(selectors, ",")+1, strings.Count(declarations, ";")+1
		if n > 64 || m > 64 {
			return 0, false
		}
		cost += n * m
		for _, declaration := range strings.Split(declarations, ";") {
			key, value, _ := strings.Cut(declaration, ":")
			switch strings.TrimSpace(strings.ToLower(key)) {
			case "stroke-width", "stroke-miterlimit", "stroke-dasharray", "stroke-dashoffset", "opacity", "fill-opacity", "stroke-opacity":
				if !boundedSVGNumbers(value) {
					return 0, false
				}
			}
		}
	}
	return cost, cost <= 4096
}

func validEmailLogoRemoteURL(u *url.URL) bool {
	if u == nil || u.User != nil || u.Hostname() == "" || u.Opaque != "" {
		return false
	}
	if ip, err := netip.ParseAddr(u.Hostname()); err == nil && !publicEmailLogoIP(ip) {
		return false
	}
	return (u.Scheme == "https" && (u.Port() == "" || u.Port() == "443")) || (u.Scheme == "http" && (u.Port() == "" || u.Port() == "80"))
}
