package api

import (
	"bytes"
	"fmt"
	"image/png"
	"strings"
	"testing"
)

func TestEmailLogoSVGSingleArgumentScale(t *testing.T) {
	for _, transform := range []string{`transform="scale(2)"`, `transform='scale(2)'`, `transform="translate(0,0) scale(2e0)"`} {
		svg := `<svg viewBox="0 0 20 20"><g ` + transform + `><rect width="10" height="10" fill="#ff0000"/></g></svg>`
		data, err := emailLogoPNG([]byte(svg), true)
		if err != nil {
			t.Fatal(transform, err)
		}
		image, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			t.Fatal(err)
		}
		_, _, _, alpha := image.At(100, 100).RGBA()
		if alpha == 0 {
			t.Fatal("single scale did not apply to both axes", transform)
		}
	}
}

func TestEmailLogoSVGStyleExpansionRejectedBeforeParser(t *testing.T) {
	selectors := make([]string, 8000)
	declarations := make([]string, 6000)
	for i := range selectors {
		selectors[i] = fmt.Sprintf(".s%d", i)
	}
	for i := range declarations {
		declarations[i] = fmt.Sprintf("p%d:0", i)
	}
	svg := `<svg viewBox="0 0 20 20"><style>` + strings.Join(selectors, ",") + `{` + strings.Join(declarations, ";") + `}</style><rect width="10" height="10"/></svg>`
	if len(svg) > 128<<10 {
		t.Fatal("fixture exceeds byte limit instead of exercising CSS rejection")
	}
	if safeEmailSVG([]byte(svg)) {
		t.Fatal("CSS selector expansion admitted")
	}
	if _, err := emailLogoPNG([]byte(svg), true); err == nil {
		t.Fatal("CSS expansion reached renderer")
	}
	for _, value := range []string{"1e100", "NaN", "Infinity"} {
		if safeEmailSVG([]byte(`<svg viewBox="0 0 20 20"><path d="M0 0 L` + value + ` 1"/></svg>`)) {
			t.Fatal("unbounded coordinate", value)
		}
	}
}
