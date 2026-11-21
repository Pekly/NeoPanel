package api

import (
	"net/http"
	"strings"
	"unicode"

	"github.com/Pekly/NeoPanel/internal/database"
)

// GetNodeWakePreset only proposes editable text. It never saves configuration,
// enables wake, changes an icon, or contacts a worker/image URL.
func (h *NodeHandler) GetNodeWakePreset(w http.ResponseWriter, r *http.Request) {
	if !wakeAdmin(w, r) {
		return
	}
	if _, ok := h.wakeNode(w, r.PathValue("id")); !ok {
		return
	}
	brand := LoadBrandingConfig(h.db)
	name := strings.Map(func(c rune) rune {
		if unicode.IsControl(c) {
			return -1
		}
		if c == '&' {
			return '＆'
		}
		if c == '§' {
			return -1
		}
		if c == '{' || c == '}' {
			return -1
		}
		return c
	}, brand.BrandName)
	if len([]rune(name)) > 64 {
		name = string([]rune(name)[:64])
	}
	if strings.TrimSpace(name) == "" {
		name = "NeoPanel"
	}
	accent := brand.AccentColor
	if len(accent) != 7 || accent[0] != '#' || strings.Trim(accent[1:], "0123456789abcdefABCDEF") != "" {
		accent = "#387aff"
	}
	c := database.DefaultNodeWakeConfig()
	header := "&" + accent + "&l" + name + " &8» &f{server}"
	c.OfflineMOTD = header + "\n&a&lJOIN TO START &8• &7Your adventure is one click away"
	c.StartingMessage = header + "\n\n&a&lWAKE REQUEST ACCEPTED\n&7Your server is starting up.\n&ePlease reconnect in a moment!"
	c.UnavailableMessage = header + "\n\n&6&lNOT READY YET\n&7The server cannot accept a wake request right now.\n&ePlease wait a moment and try again."
	writeJSON(w, http.StatusOK, map[string]any{"config": c, "brand_name": name})
}
