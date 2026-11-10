package api

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/Pekly/NeoPanel/internal/config"
	"github.com/Pekly/NeoPanel/internal/database"
)

const (
	panelSettingBrandingKey = "branding_json"
	panelSettingDaemonKey   = "daemon_settings_json"
)

var hexColorRe = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

type PanelBrandingConfig struct {
	BrandName        string `json:"brand_name"`
	BrandSubtitle    string `json:"brand_subtitle"`
	LogoURL          string `json:"logo_url"`
	FaviconURL       string `json:"favicon_url"`
	LoginBgURL       string `json:"login_bg_url"`
	ServerHeroURL    string `json:"server_hero_url"`
	CardCoverURL     string `json:"card_cover_url"`
	DefaultAvatarURL string `json:"default_avatar_url"`
	FooterText       string `json:"footer_text"`
	Preset           string `json:"preset"`
	AccentColor      string `json:"accent_color"`
	AccentHover      string `json:"accent_hover"`
	BgDeep           string `json:"bg_deep"`
	BgCanvas         string `json:"bg_canvas"`
	CardBg           string `json:"card_bg"`
	NavBg            string `json:"nav_bg"`
	TextColor        string `json:"text_color"`
	CardOpacity      int    `json:"card_opacity"`
	BorderRadiusPx   int    `json:"border_radius_px"`
	GlassBlurPx      int    `json:"glass_blur_px"`
	DefaultTheme     string `json:"default_theme"`
	CustomCSS        string `json:"custom_css"`
}

type DaemonSettingsConfig struct {
	FQDN              string `json:"fqdn"`
	SFTPPort          int    `json:"sftp_port"`
	MaxUploadMB       int    `json:"max_upload_mb"`
	DefaultRAMMB      int    `json:"default_ram_mb"`
	AutoRestartPolicy bool   `json:"auto_restart_policy"`
}

type BrandingHandler struct {
	db  *database.DB
	cfg *config.Config
	mu  sync.RWMutex
}

func NewBrandingHandler(db *database.DB, cfg *config.Config) *BrandingHandler {
	return &BrandingHandler{
		db:  db,
		cfg: cfg,
	}
}

func defaultPanelBranding() PanelBrandingConfig {
	return PanelBrandingConfig{
		BrandName:        "NeoPanel",
		BrandSubtitle:    "Ultra-Lightweight Minecraft Server Manager",
		LogoURL:          "",
		FaviconURL:       "",
		LoginBgURL:       "",
		ServerHeroURL:    "https://m.media-amazon.com/images/I/71rzlQ2XMOL.jpg",
		CardCoverURL:     "https://www.minecraft.net/content/dam/minecraftnet/games/minecraft/key-art/MC-Downloadables_Image-0_15th-Anniversary_1110x624.jpg",
		DefaultAvatarURL: "https://www.minecraft.net/content/dam/minecraftnet/games/minecraft/key-art/MCV_SummerDrop_Hero_DotNet_Downloadable_Wallpaper_r414x414.png",
		FooterText:       "Powered by NeoPanel Daemon",
		Preset:           "obsidian",
		AccentColor:      "#387aff",
		AccentHover:      "#4f8aff",
		BgDeep:           "",
		BgCanvas:         "",
		CardBg:           "",
		NavBg:            "",
		TextColor:        "",
		CardOpacity:      72,
		BorderRadiusPx:   14,
		GlassBlurPx:      16,
		DefaultTheme:     "dark",
		CustomCSS:        "",
	}
}

func (h *BrandingHandler) defaultDaemonSettings() DaemonSettingsConfig {
	sftpPort := 2022
	if h.cfg != nil && h.cfg.SFTP.Port > 0 {
		sftpPort = h.cfg.SFTP.Port
	}
	return DaemonSettingsConfig{
		FQDN:              "",
		SFTPPort:          sftpPort,
		MaxUploadMB:       500,
		DefaultRAMMB:      2048,
		AutoRestartPolicy: true,
	}
}

func (h *BrandingHandler) loadBranding() PanelBrandingConfig {
	def := defaultPanelBranding()
	if h.db == nil {
		return def
	}
	raw, err := h.db.GetPanelSetting(panelSettingBrandingKey)
	if err != nil || strings.TrimSpace(raw) == "" {
		return def
	}
	var stored PanelBrandingConfig
	if err := json.Unmarshal([]byte(raw), &stored); err != nil {
		return def
	}
	return sanitizeBrandingConfig(stored, def)
}

func (h *BrandingHandler) loadDaemonSettings() DaemonSettingsConfig {
	def := h.defaultDaemonSettings()
	if h.db == nil {
		return def
	}
	raw, err := h.db.GetPanelSetting(panelSettingDaemonKey)
	if err != nil || strings.TrimSpace(raw) == "" {
		return def
	}
	var stored DaemonSettingsConfig
	if err := json.Unmarshal([]byte(raw), &stored); err != nil {
		return def
	}
	return sanitizeDaemonSettings(stored, def)
}

func sanitizeHexOrEmpty(val, fallback string) string {
	val = strings.TrimSpace(val)
	if val == "" {
		return ""
	}
	if hexColorRe.MatchString(val) {
		return strings.ToLower(val)
	}
	return fallback
}

func sanitizeMediaURL(val, fallback string) string {
	val = strings.TrimSpace(val)
	if val == "" {
		return ""
	}
	lower := strings.ToLower(val)
	if strings.HasPrefix(lower, "https://") ||
		strings.HasPrefix(lower, "http://") ||
		strings.HasPrefix(lower, "/") ||
		strings.HasPrefix(lower, "data:image/") {
		if len(val) > 1024 {
			return val[:1024]
		}
		return val
	}
	return fallback
}

func sanitizeBrandingConfig(in, def PanelBrandingConfig) PanelBrandingConfig {
	out := def
	if name := strings.TrimSpace(in.BrandName); name != "" {
		if len(name) > 48 {
			name = name[:48]
		}
		out.BrandName = name
	}
	if sub := strings.TrimSpace(in.BrandSubtitle); sub != "" {
		if len(sub) > 120 {
			sub = sub[:120]
		}
		out.BrandSubtitle = sub
	} else {
		out.BrandSubtitle = ""
	}

	out.LogoURL = sanitizeMediaURL(in.LogoURL, def.LogoURL)
	out.FaviconURL = sanitizeMediaURL(in.FaviconURL, def.FaviconURL)
	out.LoginBgURL = sanitizeMediaURL(in.LoginBgURL, def.LoginBgURL)
	if strings.TrimSpace(in.ServerHeroURL) == "" {
		out.ServerHeroURL = def.ServerHeroURL
	} else {
		out.ServerHeroURL = sanitizeMediaURL(in.ServerHeroURL, def.ServerHeroURL)
	}
	if strings.TrimSpace(in.CardCoverURL) == "" {
		out.CardCoverURL = def.CardCoverURL
	} else {
		out.CardCoverURL = sanitizeMediaURL(in.CardCoverURL, def.CardCoverURL)
	}
	if strings.TrimSpace(in.DefaultAvatarURL) == "" {
		out.DefaultAvatarURL = def.DefaultAvatarURL
	} else {
		out.DefaultAvatarURL = sanitizeMediaURL(in.DefaultAvatarURL, def.DefaultAvatarURL)
	}

	footer := strings.TrimSpace(in.FooterText)
	if len(footer) > 120 {
		footer = footer[:120]
	}
	out.FooterText = footer

	if preset := strings.TrimSpace(strings.ToLower(in.Preset)); preset != "" {
		out.Preset = preset
	}
	if ac := sanitizeHexOrEmpty(in.AccentColor, def.AccentColor); ac != "" {
		out.AccentColor = ac
	}
	if ah := sanitizeHexOrEmpty(in.AccentHover, def.AccentHover); ah != "" {
		out.AccentHover = ah
	}
	out.BgDeep = sanitizeHexOrEmpty(in.BgDeep, "")
	out.BgCanvas = sanitizeHexOrEmpty(in.BgCanvas, "")
	out.CardBg = sanitizeHexOrEmpty(in.CardBg, "")
	out.NavBg = sanitizeHexOrEmpty(in.NavBg, "")
	out.TextColor = sanitizeHexOrEmpty(in.TextColor, "")

	if in.CardOpacity >= 35 && in.CardOpacity <= 100 {
		out.CardOpacity = in.CardOpacity
	}
	if in.BorderRadiusPx >= 0 && in.BorderRadiusPx <= 28 {
		out.BorderRadiusPx = in.BorderRadiusPx
	}
	if in.GlassBlurPx >= 0 && in.GlassBlurPx <= 32 {
		out.GlassBlurPx = in.GlassBlurPx
	}
	switch strings.ToLower(strings.TrimSpace(in.DefaultTheme)) {
	case "light", "dark", "oled", "auto":
		out.DefaultTheme = strings.ToLower(strings.TrimSpace(in.DefaultTheme))
	}

	css := in.CustomCSS
	if len(css) > 16384 {
		css = css[:16384]
	}
	out.CustomCSS = css
	return out
}

func sanitizeDaemonSettings(in, def DaemonSettingsConfig) DaemonSettingsConfig {
	out := def
	out.FQDN = strings.TrimSpace(in.FQDN)
	if len(out.FQDN) > 128 {
		out.FQDN = out.FQDN[:128]
	}
	if in.SFTPPort >= 1 && in.SFTPPort <= 65535 {
		out.SFTPPort = in.SFTPPort
	}
	if in.MaxUploadMB >= 10 && in.MaxUploadMB <= 25000 {
		out.MaxUploadMB = in.MaxUploadMB
	}
	if in.DefaultRAMMB >= 256 && in.DefaultRAMMB <= 131072 {
		out.DefaultRAMMB = in.DefaultRAMMB
	}
	out.AutoRestartPolicy = in.AutoRestartPolicy
	return out
}

func (h *BrandingHandler) GetPublicBranding(w http.ResponseWriter, r *http.Request) {
	h.mu.RLock()
	branding := h.loadBranding()
	daemon := h.loadDaemonSettings()
	h.mu.RUnlock()

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"branding": branding,
		"daemon": map[string]interface{}{
			"fqdn":           daemon.FQDN,
			"sftp_port":      daemon.SFTPPort,
			"default_ram_mb": daemon.DefaultRAMMB,
		},
	})
}

func (h *BrandingHandler) GetAdminBranding(w http.ResponseWriter, r *http.Request) {
	claims, _ := r.Context().Value(contextUserKey).(*JWTClaims)
	if claims == nil || claims.Role != "admin" {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "Admin privileges are required"})
		return
	}

	h.mu.RLock()
	branding := h.loadBranding()
	daemon := h.loadDaemonSettings()
	h.mu.RUnlock()

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"branding": branding,
		"daemon":   daemon,
	})
}

func (h *BrandingHandler) UpdateAdminBranding(w http.ResponseWriter, r *http.Request) {
	claims, _ := r.Context().Value(contextUserKey).(*JWTClaims)
	if claims == nil || claims.Role != "admin" {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "Admin privileges are required"})
		return
	}

	var req struct {
		Branding *PanelBrandingConfig  `json:"branding"`
		Daemon   *DaemonSettingsConfig `json:"daemon"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON payload"})
		return
	}

	h.mu.Lock()
	curBranding := h.loadBranding()
	curDaemon := h.loadDaemonSettings()

	if req.Branding != nil {
		nextBranding := sanitizeBrandingConfig(*req.Branding, defaultPanelBranding())
		bBytes, err := json.Marshal(nextBranding)
		if err != nil {
			h.mu.Unlock()
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to encode branding settings"})
			return
		}
		if err := h.db.SetPanelSetting(panelSettingBrandingKey, string(bBytes)); err != nil {
			h.mu.Unlock()
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to persist branding settings"})
			return
		}
		curBranding = nextBranding
	}

	if req.Daemon != nil {
		nextDaemon := sanitizeDaemonSettings(*req.Daemon, h.defaultDaemonSettings())
		dBytes, err := json.Marshal(nextDaemon)
		if err != nil {
			h.mu.Unlock()
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to encode daemon settings"})
			return
		}
		if err := h.db.SetPanelSetting(panelSettingDaemonKey, string(dBytes)); err != nil {
			h.mu.Unlock()
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to persist daemon settings"})
			return
		}
		curDaemon = nextDaemon
		if h.cfg != nil && nextDaemon.SFTPPort > 0 && h.cfg.SFTP.Port != nextDaemon.SFTPPort {
			h.cfg.SFTP.Port = nextDaemon.SFTPPort
			if h.cfg.FilePath != "" {
				_ = config.SaveConfig(h.cfg.FilePath, h.cfg)
			}
		}
	}
	h.mu.Unlock()

	logAudit(h.db, r, "admin_branding", "Updated panel visual theme, branding & daemon configuration")

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":  "Panel customization saved",
		"branding": curBranding,
		"daemon":   curDaemon,
	})
}

func (h *BrandingHandler) ResetAdminBranding(w http.ResponseWriter, r *http.Request) {
	claims, _ := r.Context().Value(contextUserKey).(*JWTClaims)
	if claims == nil || claims.Role != "admin" {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "Admin privileges are required"})
		return
	}

	h.mu.Lock()
	_ = h.db.DeletePanelSetting(panelSettingBrandingKey)
	def := defaultPanelBranding()
	daemon := h.loadDaemonSettings()
	h.mu.Unlock()

	logAudit(h.db, r, "admin_branding", "Reset panel branding and visual look to factory defaults")

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":  "Panel branding reset to defaults",
		"branding": def,
		"daemon":   daemon,
	})
}

