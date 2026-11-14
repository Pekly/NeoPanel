package api

import (
	"fmt"
	"strings"

	"github.com/Pekly/NeoPanel/internal/database"
)

func invitationEmail(brand PanelBrandingConfig, i *database.Invitation, serverName, link string) (string, string, error) {
	branding := emailBrand(brand, link)
	labels := map[string]string{
		"console.view": "View console", "console.command": "Send console commands", "server.power": "Control server power",
		"files.view": "Browse files", "files.edit": "Create and edit files", "files.delete": "Delete files",
		"sftp.access":   "Connect with SFTP (file permissions still apply)",
		"settings.view": "View settings", "settings.edit": "Edit settings", "settings.ram": "Change memory limits",
		"database.manage": "Manage databases", "backup.manage": "Manage backups", "subusers.manage": "Manage team", "schedules.manage": "Manage schedules",
	}
	permissionNames := make([]string, 0, len(i.Permissions))
	for _, key := range i.Permissions {
		if label, ok := labels[key]; ok {
			permissionNames = append(permissionNames, label)
		} else {
			permissionNames = append(permissionNames, key)
		}
	}
	permissions := strings.Join(permissionNames, ", ")
	if permissions == "" {
		permissions = "No additional permissions"
	}
	return renderPanelEmail(panelEmail{
		Branding: branding, Title: "Team invitation", Heading: "You're invited to the team",
		Intro:      fmt.Sprintf("You have been invited to collaborate on %s through %s.", serverName, branding.Name),
		Details:    []emailDetail{{"Requested username", i.Username}, {"Permissions", permissions}},
		Paragraphs: []string{"Already have an account? Sign in with the account matching this invitation. New here? Register using the requested username and invited email. This invitation does not change an existing password."},
		ActionText: "Accept invitation", ActionURL: link,
		Notice: "Expires: " + i.ExpiresAt.UTC().Format("02 Jan 2006 at 15:04 UTC") + ". This invitation link is single-use. Do not share it.",
		Footer: "If you were not expecting this invitation, you can safely ignore this email. No access is granted until you accept.",
	})
}
