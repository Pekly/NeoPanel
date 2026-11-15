package api

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
)

// Decode only bounded PNG-backed or uncompressed 24/32-bit Windows icons. The
// directory sizes/offsets and each image's actual dimensions are checked first.
func emailICOToPNG(data []byte) ([]byte, error) {
	if len(data) < 6 || !bytes.Equal(data[:4], []byte{0, 0, 1, 0}) {
		return nil, errEmailLogo
	}
	count := int(binary.LittleEndian.Uint16(data[4:6]))
	end := 6 + 16*count
	if count < 1 || count > 64 || end > len(data) {
		return nil, errEmailLogo
	}
	var best []byte
	area := 0
	for i := 0; i < count; i++ {
		entry := data[6+i*16 : 6+(i+1)*16]
		w, h := int(entry[0]), int(entry[1])
		if w == 0 {
			w = 256
		}
		if h == 0 {
			h = 256
		}
		size, offset := uint64(binary.LittleEndian.Uint32(entry[8:12])), uint64(binary.LittleEndian.Uint32(entry[12:16]))
		if offset < uint64(end) || size == 0 || offset+size > uint64(len(data)) {
			return nil, errEmailLogo
		}
		if w*h <= area {
			continue
		}
		payload := data[int(offset):int(offset+size)]
		var candidate []byte
		if bytes.HasPrefix(payload, []byte("\x89PNG\r\n\x1a\n")) {
			info, err := png.DecodeConfig(bytes.NewReader(payload))
			if err != nil || info.Width != w || info.Height != h {
				continue
			}
			candidate = payload
		} else {
			candidate, _ = emailICOBitmap(payload, w, h)
		}
		if len(candidate) > 0 {
			best = candidate
			area = w * h
		}
	}
	if len(best) == 0 {
		return nil, errEmailLogo
	}
	return best, nil
}

func emailICOBitmap(data []byte, w, h int) ([]byte, error) {
	if len(data) < 40 {
		return nil, errEmailLogo
	}
	header := int(binary.LittleEndian.Uint32(data[:4]))
	bits := int(binary.LittleEndian.Uint16(data[14:16]))
	if (header != 40 && header != 108 && header != 124) || header > len(data) || int(binary.LittleEndian.Uint32(data[4:8])) != w || int(binary.LittleEndian.Uint32(data[8:12])) != h*2 || binary.LittleEndian.Uint16(data[12:14]) != 1 || (bits != 24 && bits != 32) || binary.LittleEndian.Uint32(data[16:20]) != 0 || binary.LittleEndian.Uint32(data[32:36]) != 0 {
		return nil, errEmailLogo
	}
	stride := ((w*bits + 31) / 32) * 4
	pixelEnd := header + stride*h
	maskStride := ((w + 31) / 32) * 4
	if pixelEnd > len(data) {
		return nil, errEmailLogo
	}
	hasMask := pixelEnd+maskStride*h <= len(data)
	if bits == 24 && !hasMask {
		return nil, errEmailLogo
	}
	hasAlpha := false
	if bits == 32 {
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				if data[header+y*stride+x*4+3] != 0 {
					hasAlpha = true
				}
			}
		}
	}
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			index := header + y*stride + x*(bits/8)
			alpha := uint8(255)
			if bits == 32 && hasAlpha {
				alpha = data[index+3]
			}
			if hasMask && data[pixelEnd+y*maskStride+x/8]&(1<<uint(7-x%8)) != 0 {
				alpha = 0
			}
			img.SetNRGBA(x, h-y-1, color.NRGBA{R: data[index+2], G: data[index+1], B: data[index], A: alpha})
		}
	}
	var out bytes.Buffer
	if err := png.Encode(&out, img); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
