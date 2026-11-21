package mcwake

import (
	"strconv"
	"strings"
)

type textComponent struct {
	Text          string          `json:"text"`
	Color         string          `json:"color,omitempty"`
	Bold          bool            `json:"bold,omitempty"`
	Italic        bool            `json:"italic,omitempty"`
	Underlined    bool            `json:"underlined,omitempty"`
	Strikethrough bool            `json:"strikethrough,omitempty"`
	Obfuscated    bool            `json:"obfuscated,omitempty"`
	Extra         []textComponent `json:"extra,omitempty"`
}

var legacyColors = []string{"black", "dark_blue", "dark_green", "dark_aqua", "dark_red", "dark_purple", "gold", "gray", "dark_gray", "blue", "green", "aqua", "red", "light_purple", "yellow", "white"}
var legacyRGB = []int{0x000000, 0x0000aa, 0x00aa00, 0x00aaaa, 0xaa0000, 0xaa00aa, 0xffaa00, 0xaaaaaa, 0x555555, 0x5555ff, 0x55ff55, 0x55ffff, 0xff5555, 0xff55ff, 0xffff55, 0xffffff}

// JSON text components do not rely on clients interpreting literal section
// signs. Parse only formatting, never user-supplied JSON/HTML/click actions.
func formatText(text string, protocol int32) textComponent {
	chars := []rune(text)
	style := textComponent{}
	parts := make([]textComponent, 0, 8)
	var run strings.Builder
	flush := func() {
		if run.Len() > 0 {
			part := style
			part.Text = run.String()
			parts = append(parts, part)
			run.Reset()
		}
	}
	for i := 0; i < len(chars); i++ {
		if (chars[i] == '&' || chars[i] == '§') && i+1 < len(chars) {
			code := strings.ToLower(string(chars[i+1]))
			if code == "#" && i+7 < len(chars) {
				hex := string(chars[i+2 : i+8])
				if rgb, err := strconv.ParseUint(hex, 16, 24); err == nil {
					flush()
					style = textComponent{Color: "#" + strings.ToLower(hex)}
					if protocol < 735 {
						style.Color = nearestLegacyColor(int(rgb))
					}
					i += 7
					continue
				}
			}
			if index := strings.Index("0123456789abcdef", code); index >= 0 {
				flush()
				style = textComponent{Color: legacyColors[index]}
				i++
				continue
			}
			if strings.Contains("klmnor", code) {
				flush()
				switch code {
				case "k":
					style.Obfuscated = true
				case "l":
					style.Bold = true
				case "m":
					style.Strikethrough = true
				case "n":
					style.Underlined = true
				case "o":
					style.Italic = true
				case "r":
					style = textComponent{}
				}
				i++
				continue
			}
		}
		run.WriteRune(chars[i])
		// Bound component amplification even for callers outside validated settings.
		if len(parts) > 1024 {
			return textComponent{Text: text}
		}
	}
	flush()
	if len(parts) == 0 {
		return textComponent{Text: ""}
	}
	if len(parts) == 1 {
		return parts[0]
	}
	return textComponent{Text: "", Extra: parts}
}

func nearestLegacyColor(rgb int) string {
	best, distance := 0, int(^uint(0)>>1)
	for i, color := range legacyRGB {
		r, g, b := (rgb>>16)-(color>>16), ((rgb>>8)&255)-((color>>8)&255), (rgb&255)-(color&255)
		d := r*r + g*g + b*b
		if d < distance {
			best, distance = i, d
		}
	}
	return legacyColors[best]
}
