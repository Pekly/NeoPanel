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
		fmt.Fprintln(stdout, "Installing/verifying standard Java families: "+strings.Join(runner.RequiredJavaMajors(), ", "))
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
			opt := ""
			if runner.IsOptionalJavaMajor(ver) {
				opt = " (optional • on-demand)"
			}
			fmt.Fprintf(stdout, "Java %s: MISSING%s\n", ver, opt)
		}
	}
	fmt.Fprintf(stdout, "Verified %d/%d Java families.\n", verified, len(runner.SupportedJavaMajors()))
	if installErr != nil {
		fmt.Fprintf(stderr, "Java provisioning incomplete: %v\n", installErr)
		return 1
	}
	if args[0] == "install-all" {
		for _, ver := range runner.RequiredJavaMajors() {
			info := statuses[ver]
			if !info.Installed || info.Path == "" {
				fmt.Fprintf(stderr, "Java provisioning incomplete: required Java %s is still missing.\n", ver)
				return 1
			}
		}
	}
	return 0
}
