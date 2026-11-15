package runner

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

// CanAutoInstallJava reports platform support, not directory writability or
// upstream build availability (in particular, older Windows ARM64 releases).
func CanAutoInstallJava() bool {
	return (runtime.GOOS == "linux" || runtime.GOOS == "windows") &&
		(runtime.GOARCH == "amd64" || runtime.GOARCH == "arm64")
}

func windowsJavaInstallRoot() (string, error) {
	root := strings.TrimSpace(os.Getenv("NEOPANEL_JAVA_DIR"))
	if root == "" {
		local := os.Getenv("LOCALAPPDATA")
		if local == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", fmt.Errorf("find per-user Java directory: %w", err)
			}
			local = filepath.Join(home, "AppData", "Local")
		}
		root = filepath.Join(local, "NeoPanel", "runtimes")
	}
	if !filepath.IsAbs(root) {
		return "", fmt.Errorf("Java runtime directory must be absolute: %q", root)
	}
	return filepath.Clean(root), nil
}

func windowsJavaCandidates() []string {
	var candidates []string
	seen := map[string]bool{}
	add := func(bin string) {
		if !filepath.IsAbs(bin) || !strings.EqualFold(filepath.Ext(bin), ".exe") {
			return
		}
		bin = filepath.Clean(bin)
		key := strings.ToLower(bin)
		if !seen[key] {
			if st, err := os.Stat(bin); err == nil && st.Mode().IsRegular() {
				seen[key] = true
				candidates = append(candidates, bin)
			}
		}
	}
	addHome := func(home string) {
		if home != "" && filepath.IsAbs(home) {
			add(filepath.Join(home, "bin", "java.exe"))
			add(filepath.Join(home, "jre", "bin", "java.exe"))
		}
	}
	addHome(strings.Trim(strings.TrimSpace(os.Getenv("JAVA_HOME")), `"`))
	var roots []string
	if managed, err := windowsJavaInstallRoot(); err == nil {
		// Explicit canonical order makes default selection prefer newer LTS.
		for _, def := range canonicalPremadeJavas {
			addHome(filepath.Join(managed, "java-"+def.Version+"-temurin"))
		}
		roots = append(roots, managed)
	}
	for _, env := range []string{"ProgramFiles", "ProgramW6432", "ProgramFiles(x86)", "LOCALAPPDATA"} {
		base := os.Getenv(env)
		if base == "" || !filepath.IsAbs(base) {
			continue
		}
		for _, vendor := range []string{"Eclipse Adoptium", "Java", "Microsoft", "Eclipse Foundation", "AdoptOpenJDK"} {
			roots = append(roots, filepath.Join(base, vendor))
			if env == "LOCALAPPDATA" {
				roots = append(roots, filepath.Join(base, "Programs", vendor))
			}
		}
	}
	if home, err := os.UserHomeDir(); err == nil {
		roots = append(roots, filepath.Join(home, ".jdks"), filepath.Join(home, ".jvms"))
	}
	visited := map[string]bool{}
	for _, root := range roots {
		key := strings.ToLower(root)
		if visited[key] {
			continue
		}
		visited[key] = true
		addHome(root)
		f, err := os.Open(root)
		if err != nil {
			continue
		}
		// Discovery is bounded and never enters installer staging/backup trees.
		entries, _ := f.ReadDir(512)
		f.Close()
		sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
		for _, e := range entries {
			if !strings.HasPrefix(e.Name(), ".") {
				addHome(filepath.Join(root, e.Name()))
			}
		}
	}
	if lp, err := exec.LookPath("java.exe"); err == nil {
		add(lp)
	}
	return candidates
}

func windowsDefaultJava() string {
	// Preserve PATH as the system default; JAVA_HOME/managed runtimes are a
	// fallback when the operator has deliberately made no PATH changes.
	if lp, err := exec.LookPath("java.exe"); err == nil && detectWindowsJavaMajor(lp) != "" {
		return lp
	}
	for _, bin := range windowsJavaCandidates() {
		if detectWindowsJavaMajor(bin) != "" {
			return bin
		}
	}
	return ""
}

type javaVersionBuffer struct{ bytes.Buffer }

func (b *javaVersionBuffer) Write(p []byte) (int, error) {
	if len(p) > (64<<10)-b.Len() {
		return 0, fmt.Errorf("Java version output exceeds limit")
	}
	return b.Buffer.Write(p)
}

func detectWindowsJavaMajor(bin string) string {
	if !filepath.IsAbs(bin) || !strings.EqualFold(filepath.Ext(bin), ".exe") || !isExecutableBinary(bin) {
		return ""
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2500*time.Millisecond)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, "-version")
	var output javaVersionBuffer
	cmd.Stdout, cmd.Stderr = &output, &output
	cmd.WaitDelay = 500 * time.Millisecond
	if err := cmd.Run(); err != nil {
		return ""
	}
	return parseJavaVersionOutput(output.String())
}
