package api

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"image/color"
	"os"
	"path/filepath"
	"testing"

	"github.com/Pekly/NeoPanel/internal/config"
)

func TestEmailPNGMetadataDoesNotSelectSVGRenderer(t *testing.T) {
	original := sidebarTestPNG(t, color.RGBA{R: 255, A: 255})
	// Reproduce a PNG with a C2PA ancillary chunk containing an SVG thumbnail.
	// No user artwork is stored in the repository. The source also exceeds the
	// former 2 MiB limit, while remaining within the uploaded-logo allowance.
	metadata := make([]byte, (2<<20)+64)
	copy(metadata, []byte(`c2pa thumbnail <svg viewBox="0 0 4 4"><script>metadata, not an image</script></svg>`))
	chunk := make([]byte, len(metadata)+12)
	binary.BigEndian.PutUint32(chunk[:4], uint32(len(metadata)))
	copy(chunk[4:8], "caBX")
	copy(chunk[8:], metadata)
	binary.BigEndian.PutUint32(chunk[len(chunk)-4:], crc32.ChecksumIEEE(chunk[4:len(chunk)-4]))
	annotated := append(append(append([]byte{}, original[:33]...), chunk...), original[33:]...)
	if !bytes.Contains(annotated[:512], []byte("<svg")) {
		t.Fatal("fixture does not reproduce detection bug")
	}
	expected, err := emailLogoPNG(original, false)
	if err != nil {
		t.Fatal(err)
	}
	for _, hint := range []bool{false, true} {
		actual, err := emailLogoPNG(annotated, hint)
		if err != nil {
			t.Fatalf("PNG metadata selected SVG renderer: %v", err)
		}
		if !bytes.Equal(actual, expected) {
			t.Fatal("metadata changed the rendered logo")
		}
	}
	cfg := config.DefaultConfig()
	cfg.Paths.DataDir = t.TempDir()
	dir := filepath.Join(cfg.Paths.DataDir, "branding")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "logo_url-regression.png"), annotated, 0600); err != nil {
		t.Fatal(err)
	}
	h := &EmailHandler{cfg: cfg}
	body := renderEmailBrand(emailBrand(PanelBrandingConfig{LogoURL: "/branding/logo_url-regression.png"}, "https://panel.example"))
	_, actual, warning := h.prepareSidebarEmailLogo(body)
	if warning || !bytes.Equal(actual, expected) {
		t.Fatal("uploaded PNG replaced by fallback")
	}
}
