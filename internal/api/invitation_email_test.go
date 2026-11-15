package api

import (
	"html"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Pekly/NeoPanel/internal/database"
)

func TestInvitationEmailExactPermissions(t *testing.T) {
	permissions := []struct{ key, label string }{
		{"console.view", "View console"},
		{"console.command", "Send console commands"},
		{"server.power", "Control server power"},
		{"files.view", "Browse files"},
		{"files.edit", "Create and edit files"},
		{"files.delete", "Delete files"},
		{"sftp.access", "Connect with SFTP (file permissions still apply)"},
		{"settings.view", "View settings"},
		{"settings.edit", "Edit settings"},
		{"settings.ram", "Change memory limits"},
		{"database.manage", "Manage databases"},
		{"backup.manage", "Manage backups"},
		{"subusers.manage", "Manage team"},
		{"schedules.manage", "Manage schedules"},
	}
	cases := []struct {
		name string
		keys []string
		want string
	}{
		{"none", nil, "No additional permissions"},
		{"restricted", []string{"console.view", "files.view"}, "View console, Browse files"},
		{"unknown", []string{"future.permission"}, "future.permission"},
	}
	var allKeys, allLabels []string
	for _, permission := range permissions {
		cases = append(cases, struct {
			name string
			keys []string
			want string
		}{permission.key, []string{permission.key}, permission.label})
		allKeys = append(allKeys, permission.key)
		allLabels = append(allLabels, permission.label)
	}
	cases = append(cases, struct {
		name string
		keys []string
		want string
	}{"full", allKeys, strings.Join(allLabels, ", ")})
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			invitation := &database.Invitation{Username: "preview-member", Permissions: tc.keys, ExpiresAt: time.Date(2030, 6, 12, 15, 30, 0, 0, time.UTC)}
			output, plain, err := invitationEmail(PanelBrandingConfig{BrandName: "Example Cloud", AccentColor: "#123456"}, invitation, "Survival server", "https://panel.example.test/#invite-token="+strings.Repeat("a", 64))
			if err != nil {
				t.Fatal(err)
			}
			// Match the entire permissions block, not just substrings: restricted
			// invitations must never imply an ungranted edit/delete/admin right.
			if !strings.Contains(output, ">"+html.EscapeString(tc.want)+"</p>\n</div>") || !strings.Contains(plain, "Permissions: "+tc.want+"\r\n") {
				t.Fatalf("incorrect permissions: HTML=%s plain=%s", output, plain)
			}
			if dir := os.Getenv("NEOPANEL_EMAIL_PREVIEW_DIR"); dir != "" && (tc.name == "full" || tc.name == "restricted") {
				if err := os.WriteFile(filepath.Join(dir, "invitation-"+tc.name+".html"), []byte(output), 0600); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

func TestInvitationEmailCompactWidthSafetyAndBranding(t *testing.T) {
	username := `<img src=x onerror="alert(1)">` + strings.Repeat("member", 20)
	server := `<script>alert("server")</script>`
	permission := `future.<b>permission</b>&extra`
	link := "https://panel.example.test/?a=1&b=2#invite-token=" + strings.Repeat("b", 64)
	invitation := &database.Invitation{Username: username, Permissions: []string{"files.view", permission}, ExpiresAt: time.Date(2030, 6, 12, 17, 30, 0, 0, time.FixedZone("preview", 2*60*60))}
	brand := PanelBrandingConfig{BrandName: "Example & Cloud", LogoURL: "/brand.svg", AccentColor: "#123456", DefaultTheme: "light"}
	output, plain, err := invitationEmail(brand, invitation, server, link)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`name="viewport" content="width=device-width, initial-scale=1"`,
		`width:100%;max-width:600px;table-layout:fixed;`,
		`padding:16px 8px;`, `padding:20px;font-size:14px;line-height:1.5;`,
		`overflow-wrap:anywhere;word-break:break-word;`,
		`class="np-invitation-details"`, `<h2`, `>Permissions</h2>`,
		`src="https://panel.example.test/brand.svg"`, `alt="Example &amp; Cloud logo"`,
		`width="24" height="24"`, "#123456", "#ffffff", "#0f172a",
		html.EscapeString(username), html.EscapeString(server), html.EscapeString(permission),
		`>Accept invitation</a>`, `>Open invitation link</a>`, "copy its link address",
	} {
		if !strings.Contains(output, want) {
			t.Errorf("HTML missing %q", want)
		}
	}
	for _, unwanted := range []string{"white-space:nowrap", "<script>", "<b>permission", "ZgotmplZ", "<strong>Browse files"} {
		if strings.Contains(output, unwanted) {
			t.Errorf("HTML contains %q", unwanted)
		}
	}
	if strings.Count(output, "<img ") != 1 || strings.Count(output, `href="`+html.EscapeString(link)+`"`) != 2 || strings.Contains(output, ">"+html.EscapeString(link)+"</a>") {
		t.Error("logo or compact CTA/fallback changed")
	}
	for _, want := range []string{"12 Jun 2030 at 15:30 UTC", "Single-use link; do not share it.", "Sign in with the account matching this invitation.", "Register using the requested username and invited email.", "Existing passwords stay unchanged.", "No access is granted until you accept."} {
		if !strings.Contains(output, want) || !strings.Contains(plain, want) {
			t.Errorf("missing security/identity copy %q", want)
		}
	}
	if !strings.Contains(plain, "Accept invitation:\r\n"+link+"\r\n") || !strings.Contains(plain, username) || !strings.Contains(plain, permission) {
		t.Error("plain alternative lost usable URL or literal details")
	}
}
