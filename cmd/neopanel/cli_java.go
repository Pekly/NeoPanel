package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/Pekly/NeoPanel/internal/runner"
)

type javaCLI struct {
	installAll func() error
	statuses   func() []runner.JavaRuntimeInfo
}

func defaultJavaCLI() javaCLI {
	return javaCLI{installAll: runner.InstallRequiredJavaRuntimes, statuses: runner.GetPremadeJavaStatuses}
}

// Java commands are local host operations; they never load config or open a DB.
func cliJava(args []string, stdout, stderr io.Writer, deps javaCLI) int {
	if len(args) != 1 || (args[0] != "list" && args[0] != "install-all") {
		fmt.Fprintln(stderr, "Usage: neopanel java <install-all|list>")
		return 1
	}
	var installErr error
	if args[0] == "install-all" {
		fmt.Fprintln(stdout, "Installing/verifying required Java families: "+strings.Join(runner.SupportedJavaMajors(), ", "))
		installErr = deps.installAll()
	}
	statuses := make(map[string]runner.JavaRuntimeInfo)
	for _, info := range deps.statuses() {
		statuses[info.Version] = info
	}
	verified := 0
	for _, ver := range runner.SupportedJavaMajors() {
		info := statuses[ver]
		if info.Installed && info.Path != "" {
			fmt.Fprintf(stdout, "Java %s: verified (%s)\n", ver, info.Path)
			verified++
		} else {
			fmt.Fprintf(stdout, "Java %s: MISSING\n", ver)
		}
	}
	fmt.Fprintf(stdout, "Verified %d/%d required Java families.\n", verified, len(runner.SupportedJavaMajors()))
	if installErr != nil {
		fmt.Fprintf(stderr, "Java provisioning incomplete: %v\n", installErr)
		return 1
	}
	if args[0] == "install-all" && verified != len(runner.SupportedJavaMajors()) {
		fmt.Fprintln(stderr, "Java provisioning incomplete: required runtimes are still missing.")
		return 1
	}
	return 0
}
