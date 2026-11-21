package api

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Pekly/NeoPanel/internal/rooted"
	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
)

const avatarFileLimit int64 = 5 << 20
const avatarRequestLimit int64 = avatarFileLimit + (64 << 10)
const avatarPixelLimit = 4 << 20

var avatarProcessingSlots = make(chan struct{}, 2)
var errAvatarTooLarge = errors.New("Avatar exceeds the 5 MiB upload limit")
var errAvatarBusy = errors.New("Avatar processing is busy; please retry shortly")
var avatarFilename = regexp.MustCompile(`^avatar_[1-9][0-9]*\.(png|jpg|jpeg|webp|gif|svg)$`)

func writeAvatarError(w http.ResponseWriter, err error) {
	status := http.StatusBadRequest
	if errors.Is(err, errAvatarTooLarge) {
		status = http.StatusRequestEntityTooLarge
	}
	if errors.Is(err, errAvatarBusy) {
		status = http.StatusServiceUnavailable
		w.Header().Set("Retry-After", "2")
	}
	writeError(w, status, err.Error())
}

func saveUploadedAvatarFile(r *http.Request, userID int) (string, error) {
	if userID <= 0 {
		return "", fmt.Errorf("Invalid avatar owner")
	}
	if r.ContentLength > avatarRequestLimit {
		return "", errAvatarTooLarge
	}
	select {
	case avatarProcessingSlots <- struct{}{}:
		defer func() { <-avatarProcessingSlots }()
	default:
		return "", errAvatarBusy
	}
	r.Body = http.MaxBytesReader(nil, r.Body, avatarRequestLimit)
	err := r.ParseMultipartForm(512 << 10)
	if r.MultipartForm != nil {
		defer r.MultipartForm.RemoveAll()
	}
	if err != nil {
		return "", fmt.Errorf("Avatar request exceeds its limit or contains invalid multipart data")
	}
	file, header, err := r.FormFile("avatar")
	if err != nil {
		return "", fmt.Errorf("Avatar file field is required")
	}
	defer file.Close()
	if header.Size > avatarFileLimit {
		return "", errAvatarTooLarge
	}
	raw, err := rooted.ReadBounded(file, avatarFileLimit)
	if err != nil {
		return "", errAvatarTooLarge
	}
	cfg, format, err := image.DecodeConfig(bytes.NewReader(raw))
	if err != nil || (format != "png" && format != "jpeg" && format != "gif" && format != "webp") {
		return "", fmt.Errorf("Use a valid PNG, JPEG, GIF or WebP avatar; SVG and other active formats are not accepted")
	}
	if cfg.Width < 1 || cfg.Height < 1 || cfg.Width > 4096 || cfg.Height > 4096 || int64(cfg.Width)*int64(cfg.Height) > avatarPixelLimit {
		return "", fmt.Errorf("Avatar dimensions exceed 4096 per side or 4 megapixels")
	}
	img, _, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		return "", fmt.Errorf("Avatar image is incomplete or invalid")
	}
	width, height := cfg.Width, cfg.Height
	if width > 512 || height > 512 {
		if width >= height {
			height = max(1, height*512/width)
			width = 512
		} else {
			width = max(1, width*512/height)
			height = 512
		}
	}
	normalized := image.NewNRGBA(image.Rect(0, 0, width, height))
	draw.ApproxBiLinear.Scale(normalized, normalized.Bounds(), img, img.Bounds(), draw.Src, nil)
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, normalized); err != nil {
		return "", fmt.Errorf("Could not normalize avatar")
	}
	dir := filepath.Join("data", "avatars")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("Could not create avatar storage")
	}
	root, err := rooted.Open(dir)
	if err != nil {
		return "", fmt.Errorf("Could not open avatar storage")
	}
	defer root.Close()
	name := "avatar_" + strconv.Itoa(userID) + ".png"
	// Replacing an inode avoids truncating a tenant-planted hardlink. Root checks
	// reject symlink targets; only the normalized PNG is published after validation.
	if err := root.WriteFile(name, encoded.Bytes(), 0644); err != nil {
		return "", fmt.Errorf("Could not safely save avatar")
	}
	return fmt.Sprintf("/api/avatars/%s?t=%d", name, time.Now().UnixNano()), nil
}

func (h *AccountHandler) ServeAvatar(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/api/avatars/")
	if !avatarFilename.MatchString(name) {
		http.NotFound(w, r)
		return
	}
	root, err := rooted.Open(filepath.Join("data", "avatars"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer root.Close()
	f, err := root.Open(name)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || info.Size() > avatarFileLimit {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; sandbox")
	if strings.HasSuffix(name, ".svg") {
		// Do not silently delete older SVG uploads, but never offer them as an active
		// panel-origin document. Users can replace these with a new raster avatar.
		w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": name}))
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Cache-Control", "no-store")
	} else {
		types := map[string]string{".png": "image/png", ".jpg": "image/jpeg", ".jpeg": "image/jpeg", ".gif": "image/gif", ".webp": "image/webp"}
		w.Header().Set("Content-Type", types[filepath.Ext(name)])
		w.Header().Set("Cache-Control", "public, max-age=86400")
	}
	http.ServeContent(w, r, name, info.ModTime(), f)
}
