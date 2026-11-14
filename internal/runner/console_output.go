package runner

import (
	"regexp"
	"strings"
)

// The web console is a line viewer, not a terminal emulator. Retain SGR
// colors/styles, but discard cursor movement, erase, cursor shape and OSC codes.
var terminalControlRe = regexp.MustCompile("\x1b(?:\\[[0-?]*[ -/]*[@-~]|\\][^\x07\x1b]*(?:\x07|\x1b\\\\)|[ -/]*[@-Z\\\\-_])")
var terminalSGRRe = regexp.MustCompile("^\x1b\\[[0-9;]*m$")

func normalizeSteelConsoleLine(raw string) (string, bool) {
	clean := terminalControlRe.ReplaceAllStringFunc(raw, func(code string) string {
		if terminalSGRRe.MatchString(code) {
			return code
		}
		return ""
	})
	// Steel redraws its prompt using CR. Keep the last actual message rather
	// than presenting prompt-only redraws or cursor commands as log records.
	parts := strings.Split(clean, "\r")
	for i := len(parts) - 1; i >= 0; i-- {
		line := strings.TrimSpace(parts[i])
		visible := strings.TrimSpace(ansiEscapeRe.ReplaceAllString(line, ""))
		if visible == "" || visible == ">" {
			continue
		}
		return line, true
	}
	return "", false
}
