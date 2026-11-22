package chunker

import (
	"regexp"
	"strconv"
	"strings"
	"sync"
)

var percentRE = regexp.MustCompile(`(?:^|\s)(\d{1,3}(?:\.\d{1,2})?)%`)

// Log never blocks process pipes and retains only a bounded tail and line buffer.
type Log struct {
	mu       sync.Mutex
	tail     string
	pending  string
	complete bool
	percent  float64
}

func (l *Log) Write(b []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	n := len(b)
	for len(b) > 0 {
		take := len(b)
		if take > 4096 {
			take = 4096
		}
		s := strings.ReplaceAll(string(b[:take]), "\r", "\n")
		b = b[take:]
		l.tail += s
		if len(l.tail) > 32768 {
			l.tail = l.tail[len(l.tail)-32768:]
		}
		l.pending += s
		for {
			line, rest, ok := strings.Cut(l.pending, "\n")
			if !ok {
				break
			}
			l.pending = rest
			if strings.HasPrefix(strings.TrimSpace(line), "Conversion complete! Took ") {
				l.complete = true
			}
			if m := percentRE.FindStringSubmatch(line); len(m) == 2 {
				p, _ := strconv.ParseFloat(m[1], 64)
				if p >= 0 && p <= 100 {
					l.percent = p
				}
			}
		}
		if len(l.pending) > 4096 {
			l.pending = l.pending[len(l.pending)-4096:]
		}
	}
	return n, nil
}
func (l *Log) Snapshot() (string, float64, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.tail, l.percent, l.complete
}
