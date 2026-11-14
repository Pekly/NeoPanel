package api

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestSecurityUpdateRepositoryAndBranchValidation(t *testing.T) {
	for _, value := range []string{"https://github.com/Pekly/NeoPanel.git", "https://github.com/owner/repo-name"} {
		if _, _, ok := parseGithubRepo(value); !ok {
			t.Fatal("valid GitHub URL rejected", value)
		}
	}
	for _, value := range []string{"http://github.com/owner/repo", "https://evil.test/owner/repo", "https://github.com@evil.test/owner/repo", "https://github.com/owner/repo/extra", "https://github.com/owner/repo?command=x", "https://github.com/owner/$(id)"} {
		if _, _, ok := parseGithubRepo(value); ok {
			t.Fatal("unsafe update repository accepted", value)
		}
	}
	for _, value := range []string{"main", "v1.2.3", "feature/security-fix"} {
		if !validUpdateBranch(value) {
			t.Fatal(value)
		}
	}
	for _, value := range []string{"$(touch /tmp/pwned)", "`id`", "main; id", "../main", "-x", "refs//main", "main.lock", "main\nother"} {
		if validUpdateBranch(value) {
			t.Fatal("unsafe update branch accepted", value)
		}
	}
	f := newReliabilityFixture(t)
	for _, body := range []string{`{"repo_url":"http://github.com/owner/repo","branch":"main"}`, `{"repo_url":"https://github.com/owner/repo","branch":"$(id)"}`} {
		if w := f.request("POST", "/api/admin/updates/config", body, "admin"); w.Code != 400 {
			t.Fatal(w.Code, w.Body.String())
		}
	}
	if value, _ := f.db.GetPanelSetting("update_repo_url"); value != "" {
		t.Fatal("rejected config changed stored repository")
	}
}

func TestSecurityUpdateCommandNeverInterpolatesArguments(t *testing.T) {
	repo, branch := "https://github.com/owner/repo.git", "literal-$(id)"
	for _, helper := range []bool{false, true} {
		cmd := updateCommand(context.Background(), repo, branch, []byte("echo expected"), helper)
		if strings.Contains(strings.Join(cmd.Args, " "), " -c ") {
			t.Fatal("shell interpolation remains", cmd.Args)
		}
		if got := cmd.Args[len(cmd.Args)-4:]; !reflect.DeepEqual(got, []string{"--repo", repo, "--branch", branch}) {
			t.Fatal(got)
		}
		if !helper {
			input, _ := io.ReadAll(cmd.Stdin)
			if string(input) != "echo expected" {
				t.Fatal("installer not passed on stdin")
			}
		}
	}
}

func TestSecurityUpdateLogRejectsSymbolicLinks(t *testing.T) {
	root := t.TempDir()
	outside := filepath.Join(root, "outside.txt")
	if err := os.WriteFile(outside, []byte("unchanged"), 0600); err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(root, "state")
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatal(err)
	}
	log := filepath.Join(dir, "install.log")
	if err := os.Symlink(outside, log); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	if f, err := openUpdateLog(log); err == nil {
		f.Close()
		t.Fatal("followed update log symlink")
	}
	data, _ := os.ReadFile(outside)
	if string(data) != "unchanged" {
		t.Fatal("external file modified")
	}
	if err := os.Remove(log); err != nil {
		t.Fatal(err)
	}
	f, err := openUpdateLog(log)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
}
