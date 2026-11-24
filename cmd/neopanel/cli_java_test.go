package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Pekly/NeoPanel/internal/runner"
)

func TestCLIJavaProvisionAndList(t *testing.T) {
	for _, scenario := range []string{"list", "success", "partial", "missing"} {
		t.Run(scenario, func(t *testing.T) {
			calls := 0
			var stdout, stderr bytes.Buffer
			deps := javaCLI{
				installAll: func() error {
					calls++
					if scenario == "partial" {
						return fmt.Errorf("required Java 21 unavailable")
					}
					return nil
				},
				statuses: func() []runner.JavaRuntimeInfo {
					var out []runner.JavaRuntimeInfo
					for _, ver := range runner.SupportedJavaMajors() {
						if ver == "21" && scenario != "success" {
							continue
						}
						out = append(out, runner.JavaRuntimeInfo{Version: ver, Path: "fixture/" + ver, Installed: true})
					}
					return out
				},
			}
			action := "install-all"
			if scenario == "list" {
				action = "list"
			}
			code := cliJava([]string{action}, &stdout, &stderr, deps)
			if (code == 0) != (scenario == "success" || scenario == "list") {
				t.Fatalf("exit %d: %s", code, stderr.String())
			}
			if calls != 1 && scenario != "list" || calls != 0 && scenario == "list" {
				t.Fatal("unexpected installation")
			}
			if scenario == "success" && !strings.Contains(stdout.String(), "Verified 6/6") {
				t.Fatal(stdout.String())
			}
			if scenario != "success" && (!strings.Contains(stdout.String(), "Java 21: MISSING") || strings.Contains(stdout.String(), "Verified 6/6")) {
				t.Fatal(stdout.String())
			}
		})
	}
}

func TestCLIJavaInvalidArgumentsDoNotProvision(t *testing.T) {
	for _, args := range [][]string{nil, {"install"}, {"install-all", "25"}, {"list", "--remote"}} {
		var out bytes.Buffer
		deps := javaCLI{installAll: func() error { t.Fatal("unexpected install"); return nil }, statuses: func() []runner.JavaRuntimeInfo { t.Fatal("unexpected discovery"); return nil }}
		if cliJava(args, &out, &out, deps) == 0 {
			t.Fatalf("accepted %v", args)
		}
	}
}

func TestCLIJavaDispatcherDoesNotLoadConfig(t *testing.T) {
	// Invalid Java arguments still dispatch to the Java parser without loading
	// configuration, initializing a database, or touching runtimes/network.
	config := filepath.Join(t.TempDir(), "missing", "config.yaml")
	if code := runCLI(config, []string{"java", "bad-action"}); code != 1 {
		t.Fatal(code)
	}
	if _, err := os.Stat(filepath.Dir(config)); !os.IsNotExist(err) {
		t.Fatalf("created config/data: %v", err)
	}
	source, err := os.ReadFile("cli.go")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(source), `case "java":`) || !strings.Contains(string(source), "return cliJava(rest,") {
		t.Fatal("Java dispatcher missing")
	}
}
