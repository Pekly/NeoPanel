package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"image/png"
	"net/url"
	"strings"
	"testing"
)

func TestTOTPEnrollmentReturnsLocalQRCode(t *testing.T) {
	f := newReliabilityFixture(t)
	seen := map[string]bool{}
	for i := 0; i < 2; i++ {
		w := f.request("POST", "/api/account/2fa/generate", "{}", "owner")
		if w.Code != 200 || w.Header().Get("Cache-Control") != "no-store" {
			t.Fatalf("%d %s", w.Code, w.Body.String())
		}
		var result struct {
			Secret string `json:"secret"`
			URI    string `json:"otpauth_url"`
			QR     string `json:"qr_data_url"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		if len(result.Secret) != 32 || seen[result.Secret] {
			t.Fatal("invalid or reused secret")
		}
		seen[result.Secret] = true
		u, err := url.Parse(result.URI)
		if err != nil || u.Scheme != "otpauth" || u.Query().Get("secret") != result.Secret {
			t.Fatal("bad enrollment URI")
		}
		const prefix = "data:image/png;base64,"
		if !strings.HasPrefix(result.QR, prefix) || len(result.QR) > 65536 {
			t.Fatal("QR is not a bounded inline PNG")
		}
		data, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(result.QR, prefix))
		if err != nil {
			t.Fatal(err)
		}
		image, err := png.Decode(bytes.NewReader(data))
		if err != nil || image.Bounds().Dx() != 256 || image.Bounds().Dy() != 256 {
			t.Fatalf("invalid QR image: %v", err)
		}
	}
}

func TestTOTPQRCodeInputBound(t *testing.T) {
	if _, err := totpQRDataURL(strings.Repeat("x", 4097)); err == nil {
		t.Fatal("oversized enrollment accepted")
	}
}
