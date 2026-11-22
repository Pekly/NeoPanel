package runner

import (
	"bufio"
	"strings"
)

const maxConsoleLineBytes = 16 << 10
const maxConsoleHistoryBytes = 1 << 20
const consoleTruncated = " [NeoPanel: line truncated]"

func limitConsoleText(text string) string {
	if len(text) <= maxConsoleLineBytes {
		return text
	}
	return strings.ToValidUTF8(text[:maxConsoleLineBytes-len(consoleTruncated)], "") + consoleTruncated
}

// readConsoleLine drains oversized lines in bounded chunks. Unlike Scanner's
// token limit, this keeps later readiness and player events readable.
func readConsoleLine(reader *bufio.Reader) (string, bool) {
	var line []byte
	truncated := false
	for {
		fragment, err := reader.ReadSlice('\n')
		if err != bufio.ErrBufferFull {
			if len(fragment) > 0 && fragment[len(fragment)-1] == '\n' {
				fragment = fragment[:len(fragment)-1]
			}
			if len(fragment) > 0 && fragment[len(fragment)-1] == '\r' {
				fragment = fragment[:len(fragment)-1]
			}
		}
		remaining := maxConsoleLineBytes - len(line)
		if len(fragment) > remaining {
			line = append(line, fragment[:remaining]...)
			truncated = true
		} else {
			line = append(line, fragment...)
		}
		if err == bufio.ErrBufferFull {
			continue
		}
		if err != nil && len(line) == 0 && !truncated {
			return "", false
		}
		if truncated {
			return strings.ToValidUTF8(string(line[:maxConsoleLineBytes-len(consoleTruncated)]), "") + consoleTruncated, true
		}
		return string(line), true
	}
}
