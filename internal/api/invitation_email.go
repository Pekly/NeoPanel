package api

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"

	"github.com/Pekly/NeoPanel/internal/database"
)

var invitationEmailTemplate = template.Must(template.New("invitation").Parse(`<!doctype html>
<html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>Team invitation</title></head>
<body style="margin:0;padding:0;background:#f3f5f9;color:#202b3c;font-family:Arial,Helvetica,sans-serif;">
<table role="presentation" width="100%" cellspacing="0" cellpadding="0" style="background:#f3f5f9;"><tr><td align="center" style="padding:32px 12px;">
<table role="presentation" width="600" cellspacing="0" cellpadding="0" style="width:100%;max-width:600px;background:#ffffff;border:1px solid #e0e5ed;border-radius:12px;">
<tr><td style="padding:24px 28px;border-top:4px solid {{.Accent}};font-size:20px;font-weight:bold;">{{.Brand}}</td></tr>
<tr><td style="padding:0 28px 28px;font-size:16px;line-height:1.6;">
<h1 style="margin:0 0 16px;font-size:26px;line-height:1.3;">You're invited to the team</h1>
<p>You have been invited to collaborate on <strong>{{.Server}}</strong> through {{.Brand}}.</p>
<p>Requested username: <strong>{{.Username}}</strong><br>Permissions: {{.Permissions}}</p>
<p>Already have an account? Sign in with the account matching this invitation. New here? Register using the requested username and invited email. This invitation does not change an existing password.</p>
<table role="presentation" cellspacing="0" cellpadding="0"><tr><td style="border-radius:6px;background:#202b3c;"><a href="{{.Link}}" style="display:inline-block;padding:14px 24px;color:#ffffff;text-decoration:none;font-weight:bold;">Accept invitation</a></td></tr></table>
<p style="font-size:14px;"><strong>Expires: {{.Expires}}</strong><br>This invitation link is single-use. Do not share it.</p>
<p style="font-size:14px;color:#536176;">If the button does not work, copy and paste this link into your browser:<br><a href="{{.Link}}" style="color:#2457b2;word-break:break-all;">{{.Link}}</a></p>
</td></tr><tr><td style="padding:20px 28px;border-top:1px solid #e0e5ed;font-size:13px;line-height:1.5;color:#536176;">If you were not expecting this invitation, you can safely ignore this email. No access is granted until you accept.</td></tr>
</table></td></tr></table></body></html>`))

func invitationEmail(brand PanelBrandingConfig, i *database.Invitation, serverName, link string) (string, string, error) {
	accent := brand.AccentColor
	if !hexColorRe.MatchString(accent) {
		accent = "#387aff"
	}
	name := strings.TrimSpace(brand.BrandName)
	if name == "" {
		name = "NeoPanel"
	}
	permissions := strings.Join(i.Permissions, ", ")
	if permissions == "" {
		permissions = "No additional permissions"
	}
	data := struct {
		Brand, Accent, Server, Username, Permissions, Link, Expires string
	}{name, accent, serverName, i.Username, permissions, link, i.ExpiresAt.UTC().Format("02 Jan 2006 at 15:04 UTC")}
	var html bytes.Buffer
	if err := invitationEmailTemplate.Execute(&html, data); err != nil {
		return "", "", err
	}
	plain := fmt.Sprintf("%s — Team invitation\r\n\r\nYou have been invited to collaborate on %s.\r\nRequested username: %s\r\nPermissions: %s\r\n\r\nAlready have an account? Sign in with the account matching this invitation. New here? Register using the requested username and invited email. This invitation does not change an existing password.\r\n\r\nAccept invitation:\r\n%s\r\n\r\nExpires: %s\r\nThis invitation link is single-use. Do not share it.\r\n\r\nIf you were not expecting this invitation, you can safely ignore this email. No access is granted until you accept.\r\n", data.Brand, data.Server, data.Username, data.Permissions, data.Link, data.Expires)
	return html.String(), plain, nil
}
