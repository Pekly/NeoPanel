package sftp

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	sftpclient "github.com/pkg/sftp"
)

func portableJunction(t *testing.T, target, link string) {
	t.Helper()
	out, err := exec.Command("cmd.exe", "/c", "mklink", "/J", link, target).CombinedOutput()
	if err != nil {
		t.Fatalf("create test junction: %v: %s", err, out)
	}
	t.Cleanup(func() { os.Remove(link) })
}
func TestPortableJunctionAndAncestorReplacement(t *testing.T) {
	for _, kind := range []string{"member", "root", "ancestor"} {
		t.Run(kind, func(t *testing.T) {
			h, dir := portableTestHandler(t)
			outside := t.TempDir()
			victim := filepath.Join(outside, "file")
			must(t, os.WriteFile(victim, []byte("outside-secret"), 0600))
			p := "/file"
			switch kind {
			case "member":
				portableJunction(t, outside, filepath.Join(dir, "junction"))
				p = "/junction/file"
			case "root":
				must(t, os.Rename(dir, dir+"-old"))
				portableJunction(t, outside, dir)
			case "ancestor":
				parent := filepath.Dir(dir)
				must(t, os.Rename(parent, parent+"-old"))
				replacement := t.TempDir()
				must(t, os.Mkdir(filepath.Join(replacement, "alpha"), 0755))
				victim = filepath.Join(replacement, "alpha", "file")
				must(t, os.WriteFile(victim, []byte("outside-secret"), 0600))
				portableJunction(t, replacement, parent)
			}
			if f, err := h.fsOpen(p, os.O_WRONLY|os.O_TRUNC); err == nil {
				f.Close()
				t.Fatal("junction followed")
			}
			_, err := h.fsStat(p)
			denied(t, err)
			_, err = h.fsList(filepath.ToSlash(filepath.Dir(p)))
			denied(t, err)
			denied(t, h.fsCommand(portableSizeRequest(p, 0)))
			denied(t, h.fsCommand(portableTimesRequest(p)))
			denied(t, h.fsCommand(&sftpclient.Request{Method: "Remove", Filepath: p}))
			denied(t, h.fsCommand(&sftpclient.Request{Method: "Rename", Filepath: p, Target: "/stolen"}))
			portableSentinel(t, victim)
		})
	}
}
