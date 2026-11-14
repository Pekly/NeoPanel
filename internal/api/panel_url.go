package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/Pekly/NeoPanel/internal/database"
)

type panelURLSettings struct {
	PublicURL string `json:"public_url"`
	Source    string `json:"source"`
}

// Database overrides are read on demand: saves apply immediately without
// mutating the shared startup config or trusting request/forwarded headers.
func (h *AuthHandler) configuredPanelURL() (panelURLSettings, error) {
	raw, err := h.db.GetPanelSetting(database.PanelPublicURLKey)
	source := "settings"
	if errors.Is(err, sql.ErrNoRows) {
		h.mu.RLock()
		if h.cfg != nil {
			raw = h.cfg.Server.PublicURL
		}
		h.mu.RUnlock()
		source = "config"
	} else if err != nil {
		return panelURLSettings{}, err
	}
	if strings.TrimSpace(raw) == "" {
		return panelURLSettings{Source: "unset"}, nil
	}
	return panelURLSettings{PublicURL: strings.TrimSpace(raw), Source: source}, nil
}

func (h *AuthHandler) panelURL() (panelURLSettings, error) {
	settings, err := h.configuredPanelURL()
	if err != nil || settings.PublicURL == "" {
		return settings, err
	}
	origin, _, err := canonicalPasskeyOrigin(settings.PublicURL)
	settings.PublicURL = origin
	return settings, err
}

func (h *AuthHandler) GetPanelURL(w http.ResponseWriter, r *http.Request) {
	// Display invalid legacy configuration so an administrator can repair it.
	settings, err := h.configuredPanelURL()
	if err != nil {
		writeError(w, 500, "Could not read public panel URL; check panel storage and retry")
		return
	}
	writeJSON(w, 200, settings)
}

func (h *AuthHandler) UpdatePanelURL(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PublicURL string `json:"public_url"`
	}
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, 400, "Provide a valid public_url")
		return
	}
	if err := dec.Decode(new(any)); err != io.EOF {
		writeError(w, 400, "Provide one settings object")
		return
	}
	if strings.TrimSpace(req.PublicURL) == "" {
		writeError(w, 400, "Enter the public panel URL, for example https://panel.example.com")
		return
	}
	origin, rpID, err := canonicalPasskeyOrigin(strings.TrimSpace(req.PublicURL))
	if err != nil {
		writeError(w, 400, err.Error())
		return
	}
	if err := h.db.SetPanelPublicURL(origin, rpID); err != nil {
		if errors.Is(err, database.ErrPublicURLPasskeys) {
			writeError(w, 400, err.Error())
		} else {
			writeError(w, 500, "Could not save public panel URL")
		}
		return
	}
	logAudit(h.db, r, "admin_panel_url", "Updated public panel URL to "+origin)
	writeJSON(w, 200, panelURLSettings{PublicURL: origin, Source: "settings"})
}

// Email links must use an administrator-configured origin, never request headers.
func (h *AuthHandler) panelLinkOrigin(_ *http.Request) (string, error) {
	settings, err := h.panelURL()
	if err != nil {
		return "", err
	}
	if settings.PublicURL != "" {
		return settings.PublicURL, nil
	}
	return "", errors.New("Configure the public panel URL before sending account emails")
}
