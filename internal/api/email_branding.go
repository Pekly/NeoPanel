package api

import (
	"bytes"
	"fmt"
	"html/template"
	"math"
	"net/url"
	"strconv"
	"strings"
)

// Normal image URLs work across mail clients without fetching administrator-selected
// URLs on the daemon. Inline data/SVG markup and credential-bearing URLs are not used.
func emailLogoURL(raw, panelURL string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || len(raw) > 4096 || strings.ContainsAny(raw, "\\\r\n\x00") {
		return ""
	}
	logo, err := url.Parse(raw)
	if err != nil || logo.User != nil {
		return ""
	}
	if !logo.IsAbs() {
		if logo.Path == "" && logo.Host == "" {
			return ""
		}
		base, err := url.Parse(panelURL)
		if err != nil || base.User != nil || base.Hostname() == "" || (base.Scheme != "https" && base.Scheme != "http") {
			return ""
		}
		// Account links may include invitation/reset tokens. Resolve against ONLY
		// their trusted origin so no token, query, or route is copied into an image URL.
		logo = (&url.URL{Scheme: base.Scheme, Host: base.Host, Path: "/"}).ResolveReference(logo)
	}
	logo.Scheme = strings.ToLower(logo.Scheme)
	if (logo.Scheme != "https" && logo.Scheme != "http") || logo.Hostname() == "" || logo.User != nil || logo.Opaque != "" {
		return ""
	}
	logo.Fragment, logo.RawFragment = "", ""
	return logo.String()
}

type emailBranding struct {
	Name, Accent, LogoURL                                                      string
	Background, Surface, Header, Text, Muted, Border, Inner, ButtonText, Theme string
	Radius                                                                     int
}

func emailColor(value, fallback string) string {
	if hexColorRe.MatchString(value) {
		return strings.ToLower(value)
	}
	return fallback
}
func emailRGB(value string) [3]float64 {
	n, _ := strconv.ParseUint(strings.TrimPrefix(value, "#"), 16, 32)
	return [3]float64{float64(n >> 16 & 255), float64(n >> 8 & 255), float64(n & 255)}
}
func emailMix(a, b string, weight float64) string {
	x, y := emailRGB(a), emailRGB(b)
	return fmt.Sprintf("#%02x%02x%02x", int(math.Round(x[0]*weight+y[0]*(1-weight))), int(math.Round(x[1]*weight+y[1]*(1-weight))), int(math.Round(x[2]*weight+y[2]*(1-weight))))
}
func emailButtonText(color string) string {
	rgb := emailRGB(color)
	for i, channel := range rgb {
		channel /= 255
		if channel <= .04045 {
			rgb[i] = channel / 12.92
		} else {
			rgb[i] = math.Pow((channel+.055)/1.055, 2.4)
		}
	}
	luminance := .2126*rgb[0] + .7152*rgb[1] + .0722*rgb[2]
	if (luminance+.05)/.05 > 1.05/(luminance+.05) {
		return "#000000"
	}
	return "#ffffff"
}

func emailBrand(brand PanelBrandingConfig, panelURL string) emailBranding {
	name := strings.Join(strings.Fields(brand.BrandName), " ")
	if name == "" {
		name = "NeoPanel"
	}
	b := emailBranding{Name: name, Accent: "#2f6feb", Background: "#101318", Surface: "#181d25", Header: "#13171e", Text: "#e9edf4", Inner: "#131820", Theme: "dark", Radius: 14}
	switch strings.ToLower(brand.DefaultTheme) {
	case "light":
		b.Accent, b.Background, b.Surface, b.Header, b.Text, b.Inner, b.Theme = "#2563eb", "#f5f6f8", "#ffffff", "#ffffff", "#0f172a", "#f1f5f9", "light"
	case "oled":
		b.Background, b.Surface, b.Header, b.Inner = "#000000", "#07090d", "#07090d", "#05070a"
	}
	// Mirrors applyBranding: the default obsidian accent uses the theme's base color.
	if !((brand.Preset == "" || brand.Preset == "obsidian") && strings.EqualFold(brand.AccentColor, "#387aff")) {
		b.Accent = emailColor(brand.AccentColor, b.Accent)
	}
	b.Background = emailColor(brand.BgDeep, b.Background)
	b.Surface = emailColor(brand.CardBg, b.Surface)
	b.Header = emailColor(brand.NavBg, b.Header)
	b.Inner = emailColor(brand.BgCanvas, b.Inner)
	b.Text = emailColor(brand.TextColor, b.Text)
	b.Muted, b.Border = emailMix(b.Text, b.Surface, .72), emailMix(b.Text, b.Surface, .16)
	b.ButtonText = emailButtonText(b.Accent)
	b.LogoURL = emailLogoURL(brand.LogoURL, panelURL)
	if brand.BorderRadiusPx >= 0 && brand.BorderRadiusPx <= 28 {
		b.Radius = brand.BorderRadiusPx
	}
	return b
}

const emailBrandHeaderTemplate = `{{define "emailBrand"}}<table role="presentation" cellspacing="0" cellpadding="0" class="np-email-brand"><tr>{{if .LogoURL}}<td width="24" style="width:24px;padding:0 10px 0 0;vertical-align:middle;"><img src="{{.LogoURL}}" alt="{{.Name}} logo" width="24" height="24" style="display:block;width:24px;height:24px;object-fit:contain;border:0;"></td>{{end}}<td style="padding:0;vertical-align:middle;color:{{.Text}};font-size:18px;font-weight:bold;line-height:24px;word-break:break-word;overflow-wrap:anywhere;">{{.Name}}</td></tr></table>{{end}}`

var emailBrandHeader = template.Must(template.New("branding").Parse(emailBrandHeaderTemplate))

func renderEmailBrand(brand emailBranding) string {
	var html bytes.Buffer
	// All fields are plain strings and this fixed template has no fallible calls.
	if err := emailBrandHeader.ExecuteTemplate(&html, "emailBrand", brand); err != nil {
		return ""
	}
	return html.String()
}
