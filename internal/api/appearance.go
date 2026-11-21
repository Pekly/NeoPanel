package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// AppearanceBootstrap sends only public theme values before the first paint.
// No user state, URLs, custom CSS, or executable administrator content is included.
func (h *BrandingHandler) AppearanceBootstrap(w http.ResponseWriter, r *http.Request) {
	h.mu.RLock()
	b := h.loadBranding()
	h.mu.RUnlock()
	values := map[string]any{
		"preset": b.Preset, "accent_color": b.AccentColor, "accent_hover": b.AccentHover,
		"bg_deep": b.BgDeep, "bg_canvas": b.BgCanvas, "card_bg": b.CardBg,
		"nav_bg": b.NavBg, "text_color": b.TextColor, "card_opacity": b.CardOpacity,
		"border_radius_px": b.BorderRadiusPx, "glass_blur_px": b.GlassBlurPx, "default_theme": b.DefaultTheme,
	}
	data, err := json.Marshal(values)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Could not load appearance")
		return
	}
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	fmt.Fprintf(w, "window.NPInitialAppearance=%s;\n", data)
}
