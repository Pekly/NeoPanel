package mcwake

import "strings"

// Horizon flattens semantic JSON styles in login disconnect reasons, but its
// HUD renderer recognizes literal section codes. Keep the valid JSON packet and
// map RGB to the legacy palette, without changing normal status/MOTD encoding.
func legacyDisconnectText(message string) string {
	component := formatText(message, 47) // Legacy palette, including hex approximation.
	parts := component.Extra
	if len(parts) == 0 {
		parts = []textComponent{component}
	}
	var out strings.Builder
	for _, part := range parts {
		prefix := "§r"
		for i, name := range legacyColors {
			if part.Color == name {
				prefix += "§" + string("0123456789abcdef"[i])
				break
			}
		}
		if part.Obfuscated {
			prefix += "§k"
		}
		if part.Bold {
			prefix += "§l"
		}
		if part.Strikethrough {
			prefix += "§m"
		}
		if part.Underlined {
			prefix += "§n"
		}
		if part.Italic {
			prefix += "§o"
		}
		out.WriteString(prefix)
		// Horizon wraps plain strings and renders each explicit line independently.
		// Reassert style at newlines; client-created automatic wraps remain client-controlled.
		out.WriteString(strings.ReplaceAll(part.Text, "\n", "\n"+prefix))
	}
	return out.String()
}
