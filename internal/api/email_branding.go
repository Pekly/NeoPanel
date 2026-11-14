package api

import (
	"bytes"
	"html/template"
	"net/url"
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
	Name, Accent, LogoURL string
}

func emailBrand(brand PanelBrandingConfig, panelURL string) emailBranding {
	name := strings.TrimSpace(brand.BrandName)
	if name == "" {
		name = "NeoPanel"
	}
	accent := brand.AccentColor
	if !hexColorRe.MatchString(accent) {
		accent = "#387aff"
	}
	return emailBranding{Name: name, Accent: accent, LogoURL: emailLogoURL(brand.LogoURL, panelURL)}
}

const emailBrandHeaderTemplate = `{{define "emailBrand"}}{{if .LogoURL}}<img src="{{.LogoURL}}" alt="{{.Name}} logo" width="64" style="display:block;width:64px;height:auto;max-width:100%;border:0;margin:0 auto 12px;">{{end}}<div style="color:{{.Accent}};font-size:24px;font-weight:bold;line-height:1.3;text-align:center;">{{.Name}}</div>{{end}}`

var emailBrandHeader = template.Must(template.New("branding").Parse(emailBrandHeaderTemplate))

func renderEmailBrand(brand emailBranding) string {
	var html bytes.Buffer
	// All fields are plain strings and this fixed template has no fallible calls.
	if err := emailBrandHeader.ExecuteTemplate(&html, "emailBrand", brand); err != nil {
		return ""
	}
	return html.String()
}
