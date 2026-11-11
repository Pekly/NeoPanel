package api

import (
	"bytes"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"strconv"
	"strings"
	"time"

	"github.com/Pekly/NeoPanel/internal/config"
	"github.com/Pekly/NeoPanel/internal/database"
	"golang.org/x/crypto/bcrypt"
)

type EmailHandler struct {
	db   *database.DB
	cfg  *config.Config
	auth *AuthHandler
}

func NewEmailHandler(db *database.DB, cfg *config.Config, auth *AuthHandler) *EmailHandler {
	return &EmailHandler{
		db:   db,
		cfg:  cfg,
		auth: auth,
	}
}

func (h *EmailHandler) GetSMTPSettings() (*database.SMTPSettings, error) {
	enabledStr, _ := h.db.GetPanelSetting("smtp_enabled")
	host, _ := h.db.GetPanelSetting("smtp_host")
	portStr, _ := h.db.GetPanelSetting("smtp_port")
	username, _ := h.db.GetPanelSetting("smtp_user")
	password, _ := h.db.GetPanelSetting("smtp_pass")
	fromEmail, _ := h.db.GetPanelSetting("smtp_from_email")
	fromName, _ := h.db.GetPanelSetting("smtp_from_name")
	encryption, _ := h.db.GetPanelSetting("smtp_encryption")

	port := 587
	if p, err := strconv.Atoi(portStr); err == nil && p > 0 {
		port = p
	}
	if encryption == "" {
		if port == 465 {
			encryption = "tls"
		} else {
			encryption = "starttls"
		}
	}

	return &database.SMTPSettings{
		Enabled:    enabledStr == "true" || enabledStr == "1",
		Host:       strings.TrimSpace(host),
		Port:       port,
		Username:   strings.TrimSpace(username),
		Password:   strings.TrimSpace(password),
		FromEmail:  strings.TrimSpace(fromEmail),
		FromName:   strings.TrimSpace(fromName),
		Encryption: strings.ToLower(strings.TrimSpace(encryption)),
	}, nil
}

func (h *EmailHandler) SaveSMTPSettings(s database.SMTPSettings) error {
	enabledVal := "false"
	if s.Enabled {
		enabledVal = "true"
	}
	if err := h.db.SetPanelSetting("smtp_enabled", enabledVal); err != nil {
		return err
	}
	if err := h.db.SetPanelSetting("smtp_host", strings.TrimSpace(s.Host)); err != nil {
		return err
	}
	portVal := 587
	if s.Port > 0 {
		portVal = s.Port
	}
	if err := h.db.SetPanelSetting("smtp_port", strconv.Itoa(portVal)); err != nil {
		return err
	}
	if err := h.db.SetPanelSetting("smtp_user", strings.TrimSpace(s.Username)); err != nil {
		return err
	}
	// Only overwrite password if non-empty and not masked placeholder
	if strings.TrimSpace(s.Password) != "" && strings.TrimSpace(s.Password) != "••••••••" {
		if err := h.db.SetPanelSetting("smtp_pass", strings.TrimSpace(s.Password)); err != nil {
			return err
		}
	}
	if err := h.db.SetPanelSetting("smtp_from_email", strings.TrimSpace(s.FromEmail)); err != nil {
		return err
	}
	fromName := strings.TrimSpace(s.FromName)
	if fromName == "" {
		fromName = "NeoPanel"
	}
	if err := h.db.SetPanelSetting("smtp_from_name", fromName); err != nil {
		return err
	}
	enc := strings.ToLower(strings.TrimSpace(s.Encryption))
	if enc == "" {
		enc = "starttls"
	}
	return h.db.SetPanelSetting("smtp_encryption", enc)
}

func (h *EmailHandler) SendEmail(cfg *database.SMTPSettings, toEmail, toName, subject, bodyHTML, bodyPlain string) error {
	if cfg == nil || !cfg.Enabled || cfg.Host == "" {
		return fmt.Errorf("SMTP is not configured or disabled")
	}

	from := cfg.FromEmail
	if from == "" {
		from = cfg.Username
	}
	fromName := cfg.FromName
	if fromName == "" {
		fromName = "NeoPanel"
	}

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	// Build MIME multipart email
	var msg bytes.Buffer
	msg.WriteString(fmt.Sprintf("From: %s <%s>\r\n", fromName, from))
	if toName != "" {
		msg.WriteString(fmt.Sprintf("To: %s <%s>\r\n", toName, toEmail))
	} else {
		msg.WriteString(fmt.Sprintf("To: %s\r\n", toEmail))
	}
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("MIME-Version: 1.0\r\n")

	boundary := fmt.Sprintf("np_boundary_%d", time.Now().UnixNano())
	msg.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=%q\r\n\r\n", boundary))

	// Plain text body
	msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	msg.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	msg.WriteString("Content-Transfer-Encoding: 7bit\r\n\r\n")
	msg.WriteString(bodyPlain)
	msg.WriteString("\r\n\r\n")

	// HTML body
	if bodyHTML != "" {
		msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		msg.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
		msg.WriteString("Content-Transfer-Encoding: 7bit\r\n\r\n")
		msg.WriteString(bodyHTML)
		msg.WriteString("\r\n\r\n")
	}
	msg.WriteString(fmt.Sprintf("--%s--\r\n", boundary))

	var auth smtp.Auth
	if cfg.Username != "" {
		auth = smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
	}

	tlsConfig := &tls.Config{
		ServerName:         cfg.Host,
		InsecureSkipVerify: false,
	}

	// Port 465 or direct SSL/TLS
	if cfg.Port == 465 || strings.ToLower(cfg.Encryption) == "tls" || strings.ToLower(cfg.Encryption) == "ssl" {
		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return fmt.Errorf("SSL/TLS connection to %s failed: %w", addr, err)
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, cfg.Host)
		if err != nil {
			return fmt.Errorf("SMTP client initialization failed: %w", err)
		}
		defer client.Close()

		if auth != nil {
			if ok, _ := client.Extension("AUTH"); ok {
				if err = client.Auth(auth); err != nil {
					return fmt.Errorf("SMTP authentication failed: %w", err)
				}
			}
		}
		if err = client.Mail(from); err != nil {
			return fmt.Errorf("MAIL FROM failed: %w", err)
		}
		if err = client.Rcpt(toEmail); err != nil {
			return fmt.Errorf("RCPT TO failed: %w", err)
		}
		w, err := client.Data()
		if err != nil {
			return fmt.Errorf("DATA command failed: %w", err)
		}
		if _, err = w.Write(msg.Bytes()); err != nil {
			return fmt.Errorf("failed to write email body: %w", err)
		}
		return w.Close()
	}

	// Port 587 or STARTTLS / Plain
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP host %s: %w", addr, err)
	}
	defer client.Close()

	if strings.ToLower(cfg.Encryption) != "none" {
		if ok, _ := client.Extension("STARTTLS"); ok {
			if err = client.StartTLS(tlsConfig); err != nil {
				return fmt.Errorf("STARTTLS negotiation failed: %w", err)
			}
		}
	}

	if auth != nil {
		if ok, _ := client.Extension("AUTH"); ok {
			if err = client.Auth(auth); err != nil {
				return fmt.Errorf("SMTP authentication failed: %w", err)
			}
		}
	}

	if err = client.Mail(from); err != nil {
		return fmt.Errorf("MAIL FROM failed: %w", err)
	}
	if err = client.Rcpt(toEmail); err != nil {
		return fmt.Errorf("RCPT TO failed: %w", err)
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("DATA command failed: %w", err)
	}
	if _, err = w.Write(msg.Bytes()); err != nil {
		return fmt.Errorf("failed to write email body: %w", err)
	}
	return w.Close()
}

func (h *EmailHandler) HandleGetSettings(w http.ResponseWriter, r *http.Request) {
	s, err := h.GetSMTPSettings()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	hasPass := s.Password != ""
	s.Password = ""
	if hasPass {
		s.Password = "••••••••"
	}
	writeJSON(w, http.StatusOK, s)
}

func (h *EmailHandler) HandleSaveSettings(w http.ResponseWriter, r *http.Request) {
	var s database.SMTPSettings
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	if err := h.SaveSMTPSettings(s); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to save SMTP settings: " + err.Error()})
		return
	}

	logAudit(h.db, r, "admin_smtp_config", fmt.Sprintf("Updated SMTP server configuration (%s:%d)", s.Host, s.Port))
	writeJSON(w, http.StatusOK, map[string]string{"message": "SMTP email configuration saved successfully"})
}

func (h *EmailHandler) HandleTestEmail(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ToEmail string `json:"to_email"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	s, err := h.GetSMTPSettings()
	if err != nil || !s.Enabled || s.Host == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "SMTP is disabled or incomplete. Please enter host, port, and credentials."})
		return
	}

	toEmail := strings.TrimSpace(req.ToEmail)
	if toEmail == "" {
		toEmail = s.FromEmail
	}
	if toEmail == "" {
		toEmail = s.Username
	}
	if toEmail == "" || !strings.Contains(toEmail, "@") {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Please specify a valid recipient email address for testing"})
		return
	}

	subject := "[NeoPanel] SMTP Delivery Test"
	bodyPlain := fmt.Sprintf("Success! Your SMTP email server at %s:%d successfully delivered this message.\nSent: %s\n", s.Host, s.Port, time.Now().Format(time.RFC1123))
	bodyHTML := fmt.Sprintf(`<!DOCTYPE html><html><body style="font-family:sans-serif;background:#0f172a;color:#f8fafc;padding:30px;"><div style="max-width:480px;margin:0 auto;background:#1e293b;padding:24px;border-radius:10px;"><h2 style="color:#10b981;margin-top:0;">✔ SMTP Test Successful</h2><p>Your NeoPanel email server is functioning properly.</p><p style="color:#94a3b8;font-size:12px;">Host: %s:%d &bull; Sent: %s</p></div></body></html>`, s.Host, s.Port, time.Now().Format(time.RFC1123))

	if err := h.SendEmail(s, toEmail, "NeoPanel Administrator", subject, bodyHTML, bodyPlain); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "SMTP delivery test failed: " + err.Error()})
		return
	}

	logAudit(h.db, r, "admin_smtp_test", fmt.Sprintf("Sent test email to %s", toEmail))
	writeJSON(w, http.StatusOK, map[string]string{"message": fmt.Sprintf("Test email sent successfully to %s!", toEmail)})
}

func (h *EmailHandler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Identity       string `json:"identity"`
		TurnstileResp  string `json:"cf-turnstile-response"`
		TurnstileToken string `json:"turnstile_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	identity := strings.TrimSpace(req.Identity)
	if identity == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Please enter your account username or email address"})
		return
	}

	if h.auth != nil {
		if tsActive, _, tsSecret := h.auth.isTurnstileActive(); tsActive {
			token := req.TurnstileResp
			if token == "" {
				token = req.TurnstileToken
			}
			if err := VerifyTurnstileToken(tsSecret, token, r); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
				return
			}
		}
	}

	user, err := h.db.GetUserByEmailOrUsername(identity)
	if err != nil || user == nil {
		// Prevent user enumeration by returning standard positive message
		writeJSON(w, http.StatusOK, map[string]string{
			"message": "If an account matches that username or email address, password reset instructions have been sent.",
		})
		return
	}

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to generate security token"})
		return
	}
	resetToken := hex.EncodeToString(tokenBytes)
	expiresAt := time.Now().UTC().Add(1 * time.Hour)

	if err := h.db.CreatePasswordResetToken(user.ID, resetToken, expiresAt); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to register password reset request"})
		return
	}

	smtpCfg, _ := h.GetSMTPSettings()
	smtpReady := smtpCfg != nil && smtpCfg.Enabled && smtpCfg.Host != "" && user.Email != ""

	if smtpReady {
		host := r.Host
		scheme := "https"
		if r.TLS == nil {
			if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
				scheme = proto
			}
		}
		resetLink := fmt.Sprintf("%s://%s/#reset-token=%s", scheme, host, resetToken)

		subject := fmt.Sprintf("[NeoPanel] Reset Password for %s", user.Username)
		bodyPlain := fmt.Sprintf("Hello %s,\n\nA request was received to reset your NeoPanel password.\n\nTo choose a new password, click the link below (valid for 1 hour):\n%s\n\nReset Code: %s\n\nIf you did not request this, you can safely ignore this email.\n", user.Username, resetLink, resetToken)
		bodyHTML := fmt.Sprintf(`<!DOCTYPE html>
<html>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; background:#0b0f19; color:#f8fafc; padding:32px 16px; margin:0;">
  <div style="max-width:520px; margin:0 auto; background:#121827; border:1px solid #1e293b; border-radius:12px; padding:32px 28px;">
    <div style="margin-bottom:24px; text-align:center;">
      <h1 style="color:#387aff; margin:0; font-size:24px; font-weight:800;">NeoPanel</h1>
      <p style="color:#94a3b8; font-size:13px; margin:4px 0 0;">Password Recovery Assistance</p>
    </div>
    <h2 style="color:#ffffff; font-size:18px; margin:0 0 12px;">Password Reset Request</h2>
    <p style="color:#cbd5e1; font-size:14px; line-height:1.6; margin:0 0 20px;">
      Hello <b>%s</b>,<br><br>
      We received a request to reset the password for your NeoPanel account. Click the button below to choose a new password. This link is valid for 1 hour.
    </p>
    <div style="text-align:center; margin:28px 0;">
      <a href="%s" style="background:#387aff; color:#ffffff; padding:12px 28px; border-radius:8px; text-decoration:none; font-weight:700; font-size:14px; display:inline-block;">Reset Password</a>
    </div>
    <p style="color:#94a3b8; font-size:12.5px; line-height:1.5; margin:0 0 16px;">
      Or enter this verification code directly in the panel:<br>
      <code style="background:#1e293b; color:#38bdf8; padding:5px 10px; border-radius:6px; font-family:monospace; font-size:13px; font-weight:bold; letter-spacing:1px; display:inline-block; margin-top:6px;">%s</code>
    </p>
    <p style="color:#64748b; font-size:12px; line-height:1.5; margin:24px 0 0; border-top:1px solid #1e293b; padding-top:16px;">
      If you did not request this, you can safely ignore this email. Your current password remains unchanged.
    </p>
  </div>
</body>
</html>`, user.Username, resetLink, resetToken)

		if err := h.SendEmail(smtpCfg, user.Email, user.Username, subject, bodyHTML, bodyPlain); err != nil {
			logAuditUser(h.db, user.ID, r, "auth_password_reset_email_fail", fmt.Sprintf("Failed to send reset email: %v", err))
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"message": "Password reset token generated, but failed to deliver email via SMTP: " + err.Error(),
				"token":   resetToken,
			})
			return
		}

		logAuditUser(h.db, user.ID, r, "auth_password_reset_email_sent", fmt.Sprintf("Sent password reset link to %s", user.Email))
		writeJSON(w, http.StatusOK, map[string]string{
			"message": "Password reset instructions have been sent to your registered email address.",
		})
		return
	}

	logAuditUser(h.db, user.ID, r, "auth_password_reset_nosmtp", "Password reset requested without active SMTP")
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Password reset token generated. Note: SMTP email delivery is not configured on this server.",
		"token":   resetToken,
	})
}

func (h *EmailHandler) ConfirmPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	token := strings.TrimSpace(req.Token)
	newPassword := req.NewPassword
	if token == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Reset token is required"})
		return
	}
	if len(newPassword) < 6 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "New password must be at least 6 characters long"})
		return
	}

	pr, err := h.db.GetPasswordResetToken(token)
	if err != nil || pr == nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid or expired password reset token"})
		return
	}
	if pr.Used {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "This password reset token has already been used"})
		return
	}
	if pr.ExpiresAt.Before(time.Now().UTC()) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "This password reset token has expired. Please request a new one."})
		return
	}

	user, err := h.db.GetUserByID(pr.UserID)
	if err != nil || user == nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "User account was not found"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to encrypt new password"})
		return
	}

	if err := h.db.UpdateUserPassword(user.ID, string(hash)); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to update password in database"})
		return
	}

	_ = h.db.MarkPasswordResetUsed(token)
	logAuditUser(h.db, user.ID, r, "auth_password_reset_success", fmt.Sprintf("Password was successfully reset for user %s", user.Username))

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":  "Your password has been reset successfully! You can now log in with your new password.",
		"username": user.Username,
	})
}
