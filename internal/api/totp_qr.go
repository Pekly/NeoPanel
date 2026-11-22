package api

import (
	"encoding/base64"
	"fmt"

	qrcode "github.com/skip2/go-qrcode"
)

// TOTP enrollment contains a long-lived authentication secret. Encode locally;
// never send this URI to a third-party image or QR generation service.
func totpQRDataURL(uri string) (string, error) {
	if len(uri) > 4096 {
		return "", fmt.Errorf("authenticator enrollment URI is too long")
	}
	image, err := qrcode.Encode(uri, qrcode.Medium, 256)
	if err != nil {
		return "", err
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(image), nil
}
