package api

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"
)

type emailDetail struct{ Label, Value string }
type panelEmail struct {
	Branding                                    emailBranding
	Title, Heading, Greeting, Intro             string
	Details                                     []emailDetail
	Paragraphs                                  []string
	ActionText, ActionURL, Code, Notice, Footer string
}

// A single mail-client-friendly frame: no browser CSS, JavaScript, external fonts,
// raw HTML content, or recipient-specific theme detection is required.
var panelEmailTemplate = template.Must(template.New("panelEmail").Parse(emailBrandHeaderTemplate + `<!doctype html>
<html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><meta name="color-scheme" content="{{.Branding.Theme}}"><title>{{.Title}}</title></head>
<body style="margin:0;padding:0;background:{{.Branding.Background}};color:{{.Branding.Text}};font-family:Arial,Helvetica,sans-serif;">
<table role="presentation" width="100%" cellspacing="0" cellpadding="0" bgcolor="{{.Branding.Background}}"><tr><td align="center" style="padding:32px 12px;">
<table role="presentation" class="np-email-shell" width="600" cellspacing="0" cellpadding="0" bgcolor="{{.Branding.Surface}}" style="width:100%;max-width:600px;background:{{.Branding.Surface}};color:{{.Branding.Text}};border:1px solid {{.Branding.Border}};border-radius:{{.Branding.Radius}}px;overflow:hidden;">
<tr><td style="padding:22px 28px;background:{{.Branding.Header}};border-top:3px solid {{.Branding.Accent}};border-bottom:1px solid {{.Branding.Border}};">{{template "emailBrand" .Branding}}</td></tr>
<tr><td style="padding:28px;font-size:15px;line-height:1.65;color:{{.Branding.Text}};">
<h1 style="margin:0 0 20px;font-size:24px;line-height:1.3;color:{{.Branding.Text}};">{{.Heading}}</h1>
{{if .Greeting}}<p style="margin:0 0 12px;">{{.Greeting}}</p>{{end}}
<p style="margin:0 0 20px;">{{.Intro}}</p>
{{if .Details}}<table role="presentation" width="100%" cellspacing="0" cellpadding="0" style="margin:0 0 20px;">{{range .Details}}<tr><td style="padding:5px 12px 5px 0;vertical-align:top;white-space:nowrap;font-size:13px;">{{.Label}}</td><td style="padding:5px 0;overflow-wrap:anywhere;word-break:break-word;"><strong>{{.Value}}</strong></td></tr>{{end}}</table>{{end}}
{{range .Paragraphs}}<p style="margin:0 0 18px;">{{.}}</p>{{end}}
{{if .ActionURL}}<table role="presentation" cellspacing="0" cellpadding="0" style="margin:24px 0;"><tr><td bgcolor="{{.Branding.Accent}}" style="background:{{.Branding.Accent}};border-radius:7px;"><a href="{{.ActionURL}}" style="display:inline-block;padding:12px 22px;color:{{.Branding.ButtonText}};text-decoration:none;font-weight:bold;font-size:14px;">{{.ActionText}}</a></td></tr></table>{{end}}
{{if .Code}}<p style="margin:0 0 8px;color:{{.Branding.Muted}};font-size:13px;">Or enter this reset code in the panel:</p><div style="padding:12px 14px;background:{{.Branding.Inner}};border:1px solid {{.Branding.Border}};border-radius:7px;word-break:break-all;overflow-wrap:anywhere;"><code style="font:13px/1.6 Consolas,monospace;color:{{.Branding.Text}};">{{.Code}}</code></div>{{end}}
{{if .Notice}}<p style="margin:20px 0 0;font-size:13px;color:{{.Branding.Muted}};">{{.Notice}}</p>{{end}}
{{if .ActionURL}}<p style="margin:20px 0 0;font-size:12px;line-height:1.6;color:{{.Branding.Muted}};">If the button does not work, copy and paste this link into your browser:<br><a href="{{.ActionURL}}" style="color:{{.Branding.Text}};word-break:break-all;overflow-wrap:anywhere;">{{.ActionURL}}</a></p>{{end}}
</td></tr><tr><td style="padding:18px 28px;border-top:1px solid {{.Branding.Border}};font-size:12px;line-height:1.6;color:{{.Branding.Muted}};">{{.Footer}}</td></tr>
</table></td></tr></table></body></html>`))

func renderPanelEmail(message panelEmail) (string, string, error) {
	var html bytes.Buffer
	if err := panelEmailTemplate.Execute(&html, message); err != nil {
		return "", "", err
	}
	var plain strings.Builder
	fmt.Fprintf(&plain, "%s — %s\r\n\r\n", message.Branding.Name, message.Title)
	if message.Greeting != "" {
		fmt.Fprintf(&plain, "%s\r\n\r\n", message.Greeting)
	}
	fmt.Fprintf(&plain, "%s\r\n\r\n", message.Intro)
	for _, detail := range message.Details {
		fmt.Fprintf(&plain, "%s: %s\r\n", detail.Label, detail.Value)
	}
	for _, paragraph := range message.Paragraphs {
		fmt.Fprintf(&plain, "\r\n%s\r\n", paragraph)
	}
	if message.ActionURL != "" {
		fmt.Fprintf(&plain, "\r\n%s:\r\n%s\r\n", message.ActionText, message.ActionURL)
	}
	if message.Code != "" {
		fmt.Fprintf(&plain, "\r\nReset Code: %s\r\n", message.Code)
	}
	if message.Notice != "" {
		fmt.Fprintf(&plain, "\r\n%s\r\n", message.Notice)
	}
	fmt.Fprintf(&plain, "\r\n%s\r\n", message.Footer)
	return html.String(), plain.String(), nil
}
