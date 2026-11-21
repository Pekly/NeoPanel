package runner

import (
	"bufio"
	"fmt"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/Pekly/NeoPanel/internal/database"
)

func TestConsoleHistoryAndLinesAreByteBounded(t *testing.T) {
	p := NewProcessInstance(&database.Server{ID: "logs", Type: "paper"}, t.TempDir(), time.Second, nil, nil)
	for i := 0; i < 200; i++ {
		p.broadcast(fmt.Sprintf("%04d %s", i, strings.Repeat("x", 32<<10)))
	}
	total := 0
	for _, line := range p.GetLogs() {
		total += len(line)
		if len(line) > maxConsoleLineBytes {
			t.Fatal("oversized retained line")
		}
	}
	if total > maxConsoleHistoryBytes || p.logBufferBytes != total {
		t.Fatal("history byte limit", total, p.logBufferBytes)
	}
	if !strings.HasPrefix(p.GetLogs()[len(p.GetLogs())-1], "0199 ") {
		t.Fatal("lost latest output")
	}
	if got := limitConsoleText(strings.Repeat("界", maxConsoleLineBytes)); len(got) > maxConsoleLineBytes || !utf8.ValidString(got) || !strings.HasSuffix(got, consoleTruncated) {
		t.Fatal("bad UTF-8 truncation")
	}
}
func TestHugeConsoleLineDoesNotStopReadinessOrPlayerTracking(t *testing.T) {
	p := NewProcessInstance(&database.Server{ID: "logs", Type: "paper"}, t.TempDir(), time.Second, nil, nil)
	p.status = "starting"
	p.runID = 1
	text := strings.Repeat("x", (1<<20)+128) + "\nDone (0.01s)! For help, type help\nPlayer joined the game\n"
	p.streamOutput(strings.NewReader(text), 1)
	if p.GetStatus() != "online" || p.GetPlayerCount() != 1 {
		t.Fatal("oversized line disabled later stream tracking", p.GetStatus(), p.GetPlayerCount())
	}
	if logs := p.GetLogs(); len(logs) != 3 || !strings.HasSuffix(logs[0], consoleTruncated) {
		t.Fatal("missing truncation marker", len(logs))
	}
}
func TestBoundedConsoleLineEndingsAndEOF(t *testing.T) {
	reader := bufio.NewReaderSize(strings.NewReader("one\r\n\nlast"), 16)
	for _, want := range []string{"one", "", "last"} {
		got, ok := readConsoleLine(reader)
		if !ok || got != want {
			t.Fatal(got, ok, want)
		}
	}
	if _, ok := readConsoleLine(reader); ok {
		t.Fatal("EOF produced an extra line")
	}
}
