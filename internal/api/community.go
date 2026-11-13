package api

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/Pekly/NeoPanel/internal/database"
)

const communitySettingKey = "community_settings_json"

var discordGuildID = regexp.MustCompile(`^[0-9]{17,20}$`)
var discordInviteCode = regexp.MustCompile(`^[A-Za-z0-9_-]{2,100}$`)

type DiscordSettings struct {
	Enabled bool   `json:"enabled"`
	GuildID string `json:"guild_id"`
	Title   string `json:"title"`
}
type AnnouncementSettings struct {
	Enabled     bool   `json:"enabled"`
	Text        string `json:"text"`
	Tone        string `json:"tone"`
	Dismissible bool   `json:"dismissible"`
}
type CommunitySettings struct {
	Discord      DiscordSettings      `json:"discord"`
	Announcement AnnouncementSettings `json:"announcement"`
}
type publicDiscord struct {
	DiscordSettings
	Name        string `json:"name"`
	InviteURL   string `json:"invite_url"`
	MemberCount int    `json:"member_count"`
	Status      string `json:"status"`
}
type CommunityHandler struct {
	db            *database.DB
	client        *http.Client
	mu            sync.Mutex
	cachedGuild   string
	cachedDiscord publicDiscord
	expires       time.Time
}

func NewCommunityHandler(db *database.DB) *CommunityHandler {
	return &CommunityHandler{db: db, client: &http.Client{Timeout: 5 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}
}
func defaultCommunitySettings() CommunitySettings {
	return CommunitySettings{Discord: DiscordSettings{Title: "Join our Discord"}, Announcement: AnnouncementSettings{Tone: "info", Dismissible: true}}
}
func (h *CommunityHandler) settings() (CommunitySettings, error) {
	result := defaultCommunitySettings()
	raw, err := h.db.GetPanelSetting(communitySettingKey)
	if errors.Is(err, sql.ErrNoRows) {
		return result, nil
	}
	if err != nil {
		return result, err
	}
	err = json.Unmarshal([]byte(raw), &result)
	if err != nil {
		return result, err
	}
	err = validateCommunitySettings(&result)
	return result, err
}
func validateCommunitySettings(s *CommunitySettings) error {
	s.Discord.GuildID = strings.TrimSpace(s.Discord.GuildID)
	s.Discord.Title = strings.TrimSpace(s.Discord.Title)
	if s.Discord.Title == "" {
		s.Discord.Title = "Join our Discord"
	}
	if utf8.RuneCountInString(s.Discord.Title) > 80 {
		return errors.New("Discord title must be 80 characters or fewer")
	}
	if (s.Discord.Enabled || s.Discord.GuildID != "") && !discordGuildID.MatchString(s.Discord.GuildID) {
		return errors.New("Enter a Discord server ID containing 17–20 digits")
	}
	s.Announcement.Text = strings.TrimSpace(s.Announcement.Text)
	if utf8.RuneCountInString(s.Announcement.Text) > 500 {
		return errors.New("Announcement must be 500 characters or fewer")
	}
	if s.Announcement.Enabled && s.Announcement.Text == "" {
		return errors.New("Enter announcement text before enabling the bar")
	}
	if s.Announcement.Tone == "" {
		s.Announcement.Tone = "info"
	}
	switch s.Announcement.Tone {
	case "info", "success", "warning":
	default:
		return errors.New("Choose info, success, or warning for the announcement tone")
	}
	return nil
}
func (h *CommunityHandler) GetAdmin(w http.ResponseWriter, r *http.Request) {
	if !communityAdmin(w, r) {
		return
	}
	settings, err := h.settings()
	if err != nil {
		writeError(w, 500, "Could not load community settings")
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, 200, settings)
}
func communityAdmin(w http.ResponseWriter, r *http.Request) bool {
	c, _ := r.Context().Value(contextUserKey).(*JWTClaims)
	if c == nil || c.Role != "admin" {
		writeError(w, 403, "Administrator privileges required")
		return false
	}
	return true
}
func (h *CommunityHandler) UpdateAdmin(w http.ResponseWriter, r *http.Request) {
	if !communityAdmin(w, r) {
		return
	}
	var input struct {
		Discord      *DiscordSettings      `json:"discord"`
		Announcement *AnnouncementSettings `json:"announcement"`
	}
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8192))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&input); err != nil {
		writeError(w, 400, "Invalid community settings")
		return
	}
	if err := dec.Decode(new(any)); err != io.EOF {
		writeError(w, 400, "Provide one settings object")
		return
	}
	if input.Discord == nil || input.Announcement == nil {
		writeError(w, 400, "Provide Discord and announcement settings")
		return
	}
	settings := CommunitySettings{Discord: *input.Discord, Announcement: *input.Announcement}
	if err := validateCommunitySettings(&settings); err != nil {
		writeError(w, 400, err.Error())
		return
	}
	raw, err := json.Marshal(settings)
	if err != nil {
		writeError(w, 500, "Could not encode community settings")
		return
	}
	if err := h.db.SetPanelSetting(communitySettingKey, string(raw)); err != nil {
		writeError(w, 500, "Could not save community settings")
		return
	}
	h.mu.Lock()
	h.expires = time.Time{}
	h.mu.Unlock()
	logAudit(h.db, r, "admin_community", "Updated Discord card and announcement settings")
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, 200, settings)
}
func safeDiscordInvite(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.User != nil || u.Port() != "" || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" || u.RawPath != "" {
		return ""
	}
	host := strings.ToLower(u.Host)
	code := ""
	switch host {
	case "discord.gg":
		code = strings.TrimPrefix(u.Path, "/")
	case "discord.com":
		if strings.HasPrefix(u.Path, "/invite/") {
			code = strings.TrimPrefix(u.Path, "/invite/")
		}
	}
	if !discordInviteCode.MatchString(code) {
		return ""
	}
	return "https://discord.gg/" + code
}
func (h *CommunityHandler) discord(s DiscordSettings) publicDiscord {
	result := publicDiscord{DiscordSettings: s, Status: "disabled"}
	if !s.Enabled {
		return result
	}
	result.Status = "unavailable"
	// Never interpolate arbitrary admin input into an outbound URL, even if the DB was edited manually.
	if !discordGuildID.MatchString(s.GuildID) {
		return result
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.cachedGuild == s.GuildID && time.Now().Before(h.expires) {
		result = h.cachedDiscord
		result.DiscordSettings = s
		return result
	}
	req, err := http.NewRequest(http.MethodGet, "https://discord.com/api/guilds/"+s.GuildID+"/widget.json", nil)
	if err == nil {
		req.Header.Set("User-Agent", "NeoPanel Community Widget")
		resp, e := h.client.Do(req)
		if e == nil {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				var widget struct {
					ID       string `json:"id"`
					Name     string `json:"name"`
					Invite   string `json:"instant_invite"`
					Presence int    `json:"presence_count"`
				}
				if e = json.NewDecoder(io.LimitReader(resp.Body, 256<<10)).Decode(&widget); e == nil && widget.ID == s.GuildID {
					invite := safeDiscordInvite(widget.Invite)
					if invite != "" {
						result.Status = "ready"
						result.InviteURL = invite
						result.Name = string([]rune(widget.Name)[:min(utf8.RuneCountInString(widget.Name), 100)])
						result.MemberCount = max(0, widget.Presence)
					}
				}
			}
		}
	}
	h.cachedGuild = s.GuildID
	h.cachedDiscord = result
	h.expires = time.Now().Add(time.Minute)
	return result
}
func (h *CommunityHandler) GetPublic(w http.ResponseWriter, r *http.Request) {
	s, err := h.settings()
	if err != nil {
		writeError(w, 503, "Community settings unavailable")
		return
	}
	encoded, _ := json.Marshal(s.Announcement)
	sum := sha256.Sum256(encoded)
	announcement := struct {
		AnnouncementSettings
		ID string `json:"id"`
	}{s.Announcement, hex.EncodeToString(sum[:8])}
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, 200, map[string]any{"discord": h.discord(s.Discord), "announcement": announcement})
}
