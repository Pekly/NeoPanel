package runner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Re-exec this test binary as java.exe: no installed Java/compiler or shell is
// required, and paths/argument quoting exercise the actual Windows launcher.
func init() {
	if os.Getenv("NEOPANEL_TEST_JAVA_HELPER") != "1" || len(os.Args) != 2 || os.Args[1] != "-version" {
		return
	}
	bin, _ := os.Executable()
	major, err := os.ReadFile(filepath.Join(filepath.Dir(bin), "java.major"))
	if err != nil {
		os.Exit(1)
	}
	if string(major) == "fail" {
		fmt.Fprintln(os.Stderr, `openjdk version "25.0.1"`)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "openjdk version %q\n", string(major))
	os.Exit(0)
}

func isolateJavaWindowsEnvironment(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for _, env := range []string{"ProgramFiles", "ProgramW6432", "ProgramFiles(x86)", "LOCALAPPDATA", "USERPROFILE", "JAVA_HOME", "NEOPANEL_JAVA_DIR"} {
		t.Setenv(env, filepath.Join(root, env))
	}
	t.Setenv("PATH", root)
	t.Setenv("NEOPANEL_TEST_JAVA_HELPER", "1")
	return root
}

func copyJavaHelper(t *testing.T, home, major string) string {
	t.Helper()
	bin := filepath.Join(home, "bin", "java.exe")
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(exe)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(bin), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(bin, data, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(filepath.Dir(bin), "java.major"), []byte(major), 0600); err != nil {
		t.Fatal(err)
	}
	return bin
}

func TestJavaWindowsDiscoveryUsesActualMajor(t *testing.T) {
	root := isolateJavaWindowsEnvironment(t)
	home := filepath.Join(root, "Custom Java Home With Spaces", "jdk-25")
	bin := copyJavaHelper(t, home, "21.0.7")
	t.Setenv("JAVA_HOME", `"`+home+`"`)
	if got := detectJavaBinaryMajor(bin); got != "21" {
		t.Fatalf("major = %q", got)
	}
	if got, _ := ResolveJavaBinary("21"); got != bin {
		t.Fatalf("resolved %q, want %q", got, bin)
	}
	if got, _ := ResolveJavaBinary("default"); got != bin {
		t.Fatalf("default %q", got)
	}
	if got, _ := ResolveJavaBinary(`"` + bin + `"`); got != bin {
		t.Fatalf("absolute quoted path %q", got)
	}
	for _, major := range []string{"25", "17", "8", "11"} {
		if got, _ := ResolveJavaBinary(major); got != "" {
			t.Fatalf("selected wrong version for %s", major)
		}
	}
	for _, info := range GetPremadeJavaStatuses() {
		if info.Installed != (info.Version == "21") {
			t.Fatalf("folder name trusted as version: %+v", info)
		}
	}
	foundDefault := false
	for _, info := range DetectInstalledJavas() {
		foundDefault = foundDefault || (info.IsDefault && info.Path == bin)
	}
	if !foundDefault {
		t.Fatal("working fallback default missing from status")
	}
	// No download is needed when a verified external runtime is available.
	info, err := InstallJavaRuntime("21")
	if err != nil || info.Path != bin {
		t.Fatalf("external reuse: %+v %v", info, err)
	}
}

func TestJavaWindowsManagedAndPerUserDiscovery(t *testing.T) {
	isolateJavaWindowsEnvironment(t)
	managed := copyJavaHelper(t, filepath.Join(os.Getenv("NEOPANEL_JAVA_DIR"), "java-8-temurin"), "1.8.0_452")
	user := copyJavaHelper(t, filepath.Join(os.Getenv("LOCALAPPDATA"), "Programs", "Eclipse Adoptium", "unversioned-runtime"), "17.0.15")
	bad := copyJavaHelper(t, filepath.Join(os.Getenv("ProgramFiles"), "Java", "jdk-25"), "fail")
	copyJavaHelper(t, filepath.Join(os.Getenv("NEOPANEL_JAVA_DIR"), ".java-25-stage-hidden"), "25.0.1")
	for major, want := range map[string]string{"8": managed, "17": user} {
		if got, _ := ResolveJavaBinary(major); got != want {
			t.Fatalf("Java %s resolved %q, want %q", major, got, want)
		}
	}
	if detectJavaBinaryMajor(bad) != "" {
		t.Fatal("accepted nonzero version probe")
	}
	for _, info := range GetPremadeJavaStatuses() {
		if info.Version == "25" && info.Installed {
			t.Fatal("accepted stage or failed executable")
		}
	}
	if got, _ := ResolveJavaBinary("default"); got != managed {
		t.Fatalf("managed default %q", got)
	}
	// PATH takes precedence only for default, never for a mismatched major.
	t.Setenv("PATH", filepath.Dir(user))
	if got, _ := ResolveJavaBinary("default"); got != user {
		t.Fatalf("PATH default %q", got)
	}
	if got, _ := ResolveJavaBinary("8"); got != managed {
		t.Fatalf("PATH selected wrong major: %q", got)
	}
}

func TestJavaWindowsInstallRoot(t *testing.T) {
	root := isolateJavaWindowsEnvironment(t)
	t.Setenv("NEOPANEL_JAVA_DIR", "")
	want := filepath.Join(os.Getenv("LOCALAPPDATA"), "NeoPanel", "runtimes")
	if got, err := windowsJavaInstallRoot(); err != nil || got != want {
		t.Fatalf("root %q %v", got, err)
	}
	t.Setenv("NEOPANEL_JAVA_DIR", filepath.Join(root, "Portable Java"))
	if got, err := windowsJavaInstallRoot(); err != nil || !strings.Contains(got, "Portable Java") {
		t.Fatalf("override %q %v", got, err)
	}
	t.Setenv("NEOPANEL_JAVA_DIR", `relative\java`)
	if _, err := windowsJavaInstallRoot(); err == nil {
		t.Fatal("accepted relative managed root")
	}
	if !CanAutoInstallJava() {
		t.Fatal("Windows amd64/arm64 capability disabled")
	}
}

func TestJavaWindowsInstallRealExecutable(t *testing.T) {
	isolateJavaWindowsEnvironment(t)
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(exe)
	if err != nil {
		t.Fatal(err)
	}
	archive := javaTestZIP(t,
		javaZIPMember{name: "jdk/bin/java.exe", body: string(data)},
		javaZIPMember{name: "jdk/bin/java.major", body: "21.0.7"})
	deps, _ := javaTestInstaller(t, archive, nil)
	deps.detect = detectWindowsJavaMajor
	root := os.Getenv("NEOPANEL_JAVA_DIR")
	info, err := installWindowsJava("21", "amd64", root, deps)
	if err != nil {
		t.Fatal(err)
	}
	if got, _ := ResolveJavaBinary("21"); got != info.Path {
		t.Fatalf("installed runtime not discoverable: %q", got)
	}
	if detectJavaBinaryMajor(info.Path) != "21" {
		t.Fatal("published executable failed")
	}
}
