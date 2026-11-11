package api

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"net/url"
	"strings"
	"time"
)

const (
	totpStepSeconds = 30
	totpDigits      = 6
)

// GenerateTOTPSecret creates a new random 20-byte base32 encoded secret key.
func GenerateTOTPSecret() (string, error) {
	buf := make([]byte, 20)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	// RFC 4648 base32 without padding
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(buf), nil
}

// GenerateTOTPURL formats an otpauth:// URI recognized by all authenticator apps.
func GenerateTOTPURL(issuer, accountName, secret string) string {
	if issuer == "" {
		issuer = "NeoPanel"
	}
	label := fmt.Sprintf("%s:%s", issuer, accountName)
	v := url.Values{}
	v.Set("secret", secret)
	v.Set("issuer", issuer)
	v.Set("algorithm", "SHA1")
	v.Set("digits", "6")
	v.Set("period", "30")

	return fmt.Sprintf("otpauth://totp/%s?%s", url.PathEscape(label), v.Encode())
}

// calculateTOTP generates the 6-digit numeric string for the specified time interval counter.
func calculateTOTP(secret string, counter int64) string {
	cleanedSecret := strings.ToUpper(strings.ReplaceAll(secret, " ", ""))
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(cleanedSecret)
	if err != nil {
		// try with padding
		key, err = base32.StdEncoding.DecodeString(cleanedSecret)
		if err != nil {
			return ""
		}
	}

	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(counter))

	mac := hmac.New(sha1.New, key)
	mac.Write(buf)
	h := mac.Sum(nil)

	offset := h[len(h)-1] & 0x0f
	code := binary.BigEndian.Uint32(h[offset : offset+4])
	code &= 0x7fffffff
	code %= 1000000

	return fmt.Sprintf("%06d", code)
}

// ValidateTOTPCode checks if the provided 6-digit code matches within a +/- 1 step window (60s total tolerance).
func ValidateTOTPCode(secret, code string) bool {
	code = strings.TrimSpace(code)
	if len(code) != 6 {
		return false
	}

	now := time.Now().Unix()
	step := now / totpStepSeconds

	for _, s := range []int64{step, step - 1, step + 1} {
		expected := calculateTOTP(secret, s)
		if expected != "" && subtle.ConstantTimeCompare([]byte(expected), []byte(code)) == 1 {
			return true
		}
	}
	return false
}

// GenerateRecoveryCodes produces count random 9-character recovery codes (e.g. 1a2b-3c4d).
func GenerateRecoveryCodes(count int) []string {
	codes := make([]string, count)
	const chars = "23456789abcdefghjkmnpqrstuvwxyz"
	for i := 0; i < count; i++ {
		bytes := make([]byte, 8)
		_, _ = rand.Read(bytes)
		p1 := make([]byte, 4)
		p2 := make([]byte, 4)
		for j := 0; j < 4; j++ {
			p1[j] = chars[int(bytes[j])%len(chars)]
			p2[j] = chars[int(bytes[j+4])%len(chars)]
		}
		codes[i] = fmt.Sprintf("%s-%s", string(p1), string(p2))
	}
	return codes
}

// CheckAndConsumeRecoveryCode validates a recovery code against stored comma-separated codes and consumes it.
func CheckAndConsumeRecoveryCode(storedCodesCSV, code string) (bool, string) {
	code = strings.ToLower(strings.TrimSpace(code))
	if code == "" || storedCodesCSV == "" {
		return false, storedCodesCSV
	}

	parts := strings.Split(storedCodesCSV, ",")
	newParts := make([]string, 0, len(parts))
	found := false

	for _, p := range parts {
		cleaned := strings.ToLower(strings.TrimSpace(p))
		if cleaned == "" {
			continue
		}
		if !found && subtle.ConstantTimeCompare([]byte(cleaned), []byte(code)) == 1 {
			found = true
			continue // consume this code
		}
		newParts = append(newParts, cleaned)
	}

	return found, strings.Join(newParts, ",")
}

