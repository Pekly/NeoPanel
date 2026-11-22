package runner

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
)

type JavaRuntimeInfo struct {
	Version   string `json:"version"`    // Managed majors: 8, 11, 16, 17, 21, 25
	Name      string `json:"name"`       // "Java 25 (Latest LTS • Minecraft 26+)"
	Path      string `json:"path"`       // "/usr/lib/jvm/java-25-temurin/bin/java"
	IsDefault bool   `json:"is_default"` // true if default system java
	Installed bool   `json:"installed"`
}

type premadeJavaDef struct {
	Version string
	Name    string
}

var canonicalPremadeJavas = []premadeJavaDef{
	{Version: "25", Name: "Java 25 (Latest LTS • Minecraft 26+)"},
	{Version: "21", Name: "Java 21 (LTS • Minecraft 1.20.5 - 1.21.x)"},
	{Version: "17", Name: "Java 17 (LTS • Minecraft 1.18 - 1.20.4)"},
	{Version: "16", Name: "Java 16 (Minecraft 1.17)"},
	{Version: "11", Name: "Java 11 (LTS)"},
	{Version: "8", Name: "Java 8 (Legacy • Minecraft 1.8 - 1.16.5)"},
}

var (
	javaInstallMu sync.Mutex
	javaVersionRe = regexp.MustCompile(`"([^"]+)"`)
	javaSupported = func() map[string]bool {
		m := make(map[string]bool)
		for _, def := range canonicalPremadeJavas {
			m[def.Version] = true
		}
		return m
	}()
)

func premadeDisplayName(ver string) string {
	for _, p := range canonicalPremadeJavas {
		if p.Version == ver {
			return p.Name
		}
	}
	return "OpenJDK " + ver
}

func parseJavaTag(versionOrPath string) string {
	v := strings.TrimSpace(strings.ToLower(versionOrPath))
	if v == "" || v == "default" || v == "java" {
		return "default"
	}

	switch v {
	case "25", "21", "17", "16", "11", "8":
		return v
	case "1.8":
		return "8"
	}

	if maj := matchJavaMajorInString(v); maj != "" {
		return maj
	}
	return strings.TrimSpace(versionOrPath)
}

func matchJavaMajorInString(s string) string {
	lower := strings.ToLower(strings.TrimSpace(s))
	if lower == "" {
		return ""
	}
	// Check 1.8 before tokenized checks
	if containsBoundedToken(lower, "1.8") {
		return "8"
	}
	for _, ver := range SupportedJavaMajors() {
		if containsBoundedToken(lower, ver) {
			return ver
		}
	}
	return ""
}

func containsBoundedToken(s, token string) bool {
	idx := 0
	for {
		pos := strings.Index(s[idx:], token)
		if pos < 0 {
			return false
		}
		start := idx + pos
		end := start + len(token)

		prevIsDigit := start > 0 && s[start-1] >= '0' && s[start-1] <= '9'
		nextIsDigit := end < len(s) && s[end] >= '0' && s[end] <= '9'
		if !prevIsDigit && !nextIsDigit {
			return true
		}
		idx = start + 1
	}
}

func detectJavaBinaryMajor(binPath string) string {
	if runtime.GOOS == "windows" {
		return detectWindowsJavaMajor(binPath)
	}
	if !isExecutableBinary(binPath) {
		return ""
	}
	return probeJavaMajor(binPath)
}

func parseJavaVersionOutput(out string) string {
	m := javaVersionRe.FindStringSubmatch(out)
	if len(m) >= 2 {
		verStr := strings.TrimSpace(m[1])
		if strings.HasPrefix(verStr, "1.") {
			parts := strings.Split(verStr, ".")
			if len(parts) >= 2 {
				return trimNonDigits(parts[1])
			}
		}
		parts := strings.FieldsFunc(verStr, func(r rune) bool {
			return r == '.' || r == '-' || r == '_' || r == '+'
		})
		if len(parts) > 0 {
			return trimNonDigits(parts[0])
		}
	}
	return ""
}

func trimNonDigits(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] >= '0' && s[i] <= '9' {
			b.WriteByte(s[i])
		} else {
			break
		}
	}
	return b.String()
}

// ResolveJavaBinary converts a version string ("25", "21", "17", "8", etc.) or path
// into a validated executable path on the host system.
func ResolveJavaBinary(versionOrPath string) (binaryPath string, displayLabel string) {
	v := strings.TrimSpace(versionOrPath)
	if runtime.GOOS == "windows" {
		v = strings.Trim(v, `"`)
	}

	// 1. Direct path check (user provided full binary path)
	if v != "" && !strings.EqualFold(v, "default") && !strings.EqualFold(v, "java") {
		if filepath.IsAbs(v) {
			if fi, err := os.Stat(v); err == nil && !fi.IsDir() {
				return v, "Custom Java (" + filepath.Base(v) + ")"
			}
			return v, "Custom Java (" + filepath.Base(v) + " - Not Found)"
		}
	}

	// 2. Identify target version tag
	targetTag := parseJavaTag(v)

	// If explicit default requested, return system "java"
	if targetTag == "default" {
		if runtime.GOOS == "windows" {
			if bin := windowsDefaultJava(); bin != "" {
				return bin, "System Default (OpenJDK " + detectJavaBinaryMajor(bin) + ")"
			}
			return "", "System Default Java (Not Installed)"
		}
		if lp, err := exec.LookPath("java"); err == nil {
			if maj := detectJavaBinaryMajor(lp); maj != "" {
				return lp, "System Default (OpenJDK " + maj + ")"
			}
			return lp, "System Default Java"
		}
		return "java", "System Default Java"
	}

	// 3. Search standard Linux JVM installations (/usr/lib/jvm, /opt/java)
	if runtime.GOOS == "linux" {
		candidates := getLinuxJavaCandidates(targetTag)
		for _, cand := range candidates {
			if isExecutableBinary(cand) {
				if maj := detectJavaBinaryMajor(cand); maj == targetTag {
					return cand, "OpenJDK " + targetTag
				}
			}
		}
		if dynamicPath := scanLinuxJVMsForMajor(targetTag); dynamicPath != "" {
			return dynamicPath, "OpenJDK " + targetTag
		}
	}

	// 4. Windows: validate JAVA_HOME, managed, per-user, vendor and PATH runtimes.
	if runtime.GOOS == "windows" {
		candidates := getWindowsJavaCandidates(targetTag)
		for _, cand := range candidates {
			if isExecutableBinary(cand) {
				if maj := detectJavaBinaryMajor(cand); maj == targetTag {
					return cand, "Java " + targetTag
				}
			}
		}
		// An empty path fails Start's executable check without allowing a
		// similarly named but unverified executable on PATH to be launched.
		return "", "Java " + targetTag + " (Not Installed)"
	}

	// 5. Check system PATH "java" only if its major version matches the requested target
	if lp, err := exec.LookPath("java"); err == nil {
		maj := detectJavaBinaryMajor(lp)
		if maj == targetTag {
			return lp, "OpenJDK " + targetTag + " (System PATH)"
		}
		if !javaSupported[targetTag] && maj == "" {
			return lp, "System Default Java"
		}
	}

	// Never return a canonical path that may contain an unverified/wrong-major
	// executable. Missing explicit families fail closed on both platforms.
	if runtime.GOOS == "linux" && javaSupported[targetTag] {
		return "", "Java " + targetTag + " (Not Installed)"
	}

	if lp, err := exec.LookPath("java"); err == nil && !javaSupported[targetTag] {
		return lp, "System Default Java"
	}

	return "java-" + targetTag, "Java " + targetTag + " (Not Installed)"
}

// Only host-managed runtime installations may add mounts. Tenant executables
// run under the existing server bind; neither their ancestors nor symlinked
// conf/security targets are inferred as host mounts.
func javaRuntimeMountArgs(binary, serverDir string) []string {
	if rel, e := filepath.Rel(serverDir, binary); e == nil && filepath.IsLocal(rel) {
		return nil
	}
	real, e := filepath.EvalSymlinks(binary)
	if e != nil {
		return nil
	}
	if rel, e := filepath.Rel(serverDir, real); e == nil && filepath.IsLocal(rel) {
		return nil
	}
	for _, base := range []string{"/usr/lib/jvm", "/usr/java", "/opt/java"} {
		rel, e := filepath.Rel(base, real)
		if e != nil || !filepath.IsLocal(rel) {
			continue
		}
		parts := strings.Split(filepath.ToSlash(rel), "/")
		if len(parts) < 3 || parts[len(parts)-2] != "bin" || parts[len(parts)-1] != "java" {
			return nil
		}
		home := filepath.Join(base, parts[0])
		if !hostManagedJavaPath(real) || !hostManagedJavaPath(home) {
			return nil
		}
		return []string{"--ro-bind", home, home}
	}
	return nil
}

func scanLinuxJVMsForMajor(targetTag string) string {
	return scanLinuxJavaRoots(targetTag, []string{"/usr/lib/jvm", "/opt/java", "/usr/java"}, detectJavaBinaryMajor)
}

func scanLinuxJavaRoots(targetTag string, roots []string, detect func(string) string) string {
	for _, root := range roots {
		entries, err := os.ReadDir(root)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if strings.HasPrefix(e.Name(), ".") {
				continue
			}
			fullDir := filepath.Join(root, e.Name())
			fi, err := os.Stat(fullDir)
			if err != nil || !fi.IsDir() {
				continue
			}
			for _, rel := range []string{filepath.Join("bin", "java"), filepath.Join("jre", "bin", "java")} {
				javaBin := filepath.Join(fullDir, rel)
				if !isExecutableBinary(javaBin) {
					continue
				}
				if detect(javaBin) == targetTag {
					return javaBin
				}
			}
		}
	}
	return ""
}

func getLinuxJavaCandidates(tag string) []string {
	archs := []string{"amd64", "x64", "arm64", "aarch64", "i386"}
	var candidates []string

	candidates = append(candidates,
		"/usr/lib/jvm/java-"+tag+"-temurin/bin/java",
		"/usr/lib/jvm/temurin-"+tag+"/bin/java",
		"/usr/lib/jvm/jdk-"+tag+"/bin/java",
		"/usr/lib/jvm/jre-"+tag+"/bin/java",
		"/usr/lib/jvm/java-"+tag+"/bin/java",
		"/usr/lib/jvm/java-"+tag+"-openjdk/bin/java",
	)

	if tag == "8" {
		for _, arch := range archs {
			candidates = append(candidates,
				"/usr/lib/jvm/java-8-openjdk-"+arch+"/jre/bin/java",
				"/usr/lib/jvm/java-8-openjdk-"+arch+"/bin/java",
				"/usr/lib/jvm/java-1.8.0-openjdk-"+arch+"/jre/bin/java",
				"/usr/lib/jvm/java-1.8.0-openjdk-"+arch+"/bin/java",
				"/usr/lib/jvm/java-8-temurin-"+arch+"/bin/java",
			)
		}
		candidates = append(candidates,
			"/usr/lib/jvm/java-8-temurin/jre/bin/java",
			"/usr/lib/jvm/java-8-openjdk/jre/bin/java",
			"/usr/lib/jvm/java-8-openjdk/bin/java",
			"/usr/lib/jvm/default-java/bin/java",
		)
		return candidates
	}

	for _, arch := range archs {
		candidates = append(candidates,
			"/usr/lib/jvm/java-"+tag+"-temurin-"+arch+"/bin/java",
			"/usr/lib/jvm/temurin-"+tag+"-jdk-"+arch+"/bin/java",
			"/usr/lib/jvm/temurin-"+tag+"-jre-"+arch+"/bin/java",
			"/usr/lib/jvm/java-"+tag+"-openjdk-"+arch+"/bin/java",
			"/usr/lib/jvm/jdk-"+tag+"-openjdk-"+arch+"/bin/java",
		)
	}

	return candidates
}

func getWindowsJavaCandidates(tag string) []string {
	// Directory names are only discovery hints, never version evidence.
	return windowsJavaCandidates()
}

// DetectInstalledJavas returns a list of installed Java runtimes found on the machine.
func DetectInstalledJavas() []JavaRuntimeInfo {
	byVersion := make(map[string]JavaRuntimeInfo)
	var extraResults []JavaRuntimeInfo
	seenPaths := make(map[string]bool)

	record := func(ver, name, path string, isDefault bool) {
		if path == "" || (seenPaths[path] && !(runtime.GOOS == "windows" && isDefault)) {
			return
		}
		if !isExecutableBinary(path) {
			return
		}
		seenPaths[path] = true
		info := JavaRuntimeInfo{
			Version:   ver,
			Name:      name,
			Path:      path,
			IsDefault: isDefault,
			Installed: true,
		}
		if ver == "default" {
			extraResults = append(extraResults, info)
			return
		}
		if _, exists := byVersion[ver]; !exists {
			byVersion[ver] = info
		}
	}

	// 1. Detect Linux /usr/lib/jvm & /opt/java directories (real directories first, then symlinks)
	if runtime.GOOS == "linux" {
		for _, root := range []string{"/usr/lib/jvm", "/opt/java", "/usr/java"} {
			entries, err := os.ReadDir(root)
			if err != nil {
				continue
			}
			for _, passSymlinks := range []bool{false, true} {
				for _, e := range entries {
					dirName := e.Name()
					if strings.HasPrefix(dirName, ".") || dirName == "default-java" {
						continue
					}
					isSym := e.Type()&os.ModeSymlink != 0
					if isSym != passSymlinks {
						continue
					}
					fullDir := filepath.Join(root, dirName)
					fi, err := os.Stat(fullDir)
					if err != nil || !fi.IsDir() {
						continue
					}

					javaBin := filepath.Join(fullDir, "bin", "java")
					if !isExecutableBinary(javaBin) {
						javaBin = filepath.Join(fullDir, "jre", "bin", "java")
						if !isExecutableBinary(javaBin) {
							continue
						}
					}

					maj := detectJavaBinaryMajor(javaBin)
					switch maj {
					case "25", "21", "17", "16", "11", "8":
						record(maj, premadeDisplayName(maj), javaBin, false)
					case "":
						continue
					default:
						record(maj, "OpenJDK "+maj+" ("+dirName+")", javaBin, false)
					}
				}
			}
		}
	} else if runtime.GOOS == "windows" {
		for _, cand := range windowsJavaCandidates() {
			if maj := detectJavaBinaryMajor(cand); maj != "" {
				record(maj, premadeDisplayName(maj), cand, false)
			}
		}
	}

	// 2. Default system Java (and also backfill major version if e.g. system java was installed outside /usr/lib/jvm)
	lp, lookupErr := exec.LookPath("java")
	if runtime.GOOS == "windows" {
		lp = windowsDefaultJava()
		if lp == "" {
			lookupErr = fmt.Errorf("no working Java runtime")
		} else {
			lookupErr = nil
		}
	}
	if lookupErr == nil {
		resolvedLP, errReal := filepath.EvalSymlinks(lp)
		if errReal == nil && isExecutableBinary(resolvedLP) {
			lp = resolvedLP
		}
		maj := detectJavaBinaryMajor(lp)
		if maj != "" && javaSupported[maj] {
			if _, exists := byVersion[maj]; !exists {
				byVersion[maj] = JavaRuntimeInfo{
					Version:   maj,
					Name:      premadeDisplayName(maj),
					Path:      lp,
					IsDefault: false,
					Installed: true,
				}
			}
		}
		defLabel := "System Default (PATH)"
		if maj != "" {
			defLabel = "System Default (OpenJDK " + maj + ")"
		}
		record("default", defLabel, lp, true)
	}

	var ordered []JavaRuntimeInfo
	for _, def := range canonicalPremadeJavas {
		if info, ok := byVersion[def.Version]; ok {
			ordered = append(ordered, info)
			delete(byVersion, def.Version)
		}
	}
	for _, info := range byVersion {
		ordered = append(ordered, info)
	}
	ordered = append(ordered, extraResults...)

	if len(ordered) == 0 {
		ordered = append(ordered, JavaRuntimeInfo{
			Version:   "25",
			Name:      premadeDisplayName("25"),
			Path:      "java",
			IsDefault: true,
			Installed: false,
		})
	}

	return ordered
}

// GetPremadeJavaStatuses returns the supported managed Java runtimes
// along with whether each runtime is currently installed on the host.
func GetPremadeJavaStatuses() []JavaRuntimeInfo {
	installed := DetectInstalledJavas()
	installedMap := make(map[string]JavaRuntimeInfo)
	for _, info := range installed {
		if info.Installed && info.Version != "" && info.Version != "default" {
			installedMap[info.Version] = info
		}
	}

	out := make([]JavaRuntimeInfo, 0, len(canonicalPremadeJavas))
	for _, def := range canonicalPremadeJavas {
		if hit, ok := installedMap[def.Version]; ok {
			hit.Name = def.Name
			hit.Installed = true
			out = append(out, hit)
		} else {
			out = append(out, JavaRuntimeInfo{
				Version:   def.Version,
				Name:      def.Name,
				Path:      "",
				Installed: false,
			})
		}
	}
	return out
}

// InstallJavaRuntime downloads and installs a supported official Eclipse Temurin runtime
// into /usr/lib/jvm on Linux or the per-user managed runtime directory on Windows.
func InstallJavaRuntime(version string) (JavaRuntimeInfo, error) {
	ver := strings.TrimSpace(version)
	if !javaSupported[ver] {
		return JavaRuntimeInfo{}, fmt.Errorf("unsupported Java version %q (supported: %s)", version, strings.Join(SupportedJavaMajors(), ", "))
	}
	if !CanAutoInstallJava() {
		return JavaRuntimeInfo{}, fmt.Errorf("automatic runtime installation is unsupported on %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	javaInstallMu.Lock()
	defer javaInstallMu.Unlock()

	for _, st := range GetPremadeJavaStatuses() {
		if st.Version == ver && st.Installed {
			return st, nil
		}
	}
	root := "/usr/lib/jvm"
	if runtime.GOOS == "windows" {
		var err error
		root, err = windowsJavaInstallRoot()
		if err != nil {
			return JavaRuntimeInfo{}, err
		}
	}
	deps := defaultWindowsJavaInstaller()
	deps.detect = detectJavaBinaryMajor
	return installTemurinJava(ver, runtime.GOOS, runtime.GOARCH, root, deps)
}

// SupportedJavaMajors returns a copy of the deterministic managed-runtime catalog.
func SupportedJavaMajors() []string {
	versions := make([]string, 0, len(canonicalPremadeJavas))
	for _, def := range canonicalPremadeJavas {
		versions = append(versions, def.Version)
	}
	return versions
}

// InstallRequiredJavaRuntimes provisions and verifies every required family on this host.
// Existing exact-major installations are reused. A partial failure is never success.
func InstallRequiredJavaRuntimes() error {
	return installRequiredJavaRuntimes(ResolveJavaBinary, detectJavaBinaryMajor, InstallJavaRuntime)
}

func installRequiredJavaRuntimes(resolve func(string) (string, string), detect func(string) string, install func(string) (JavaRuntimeInfo, error)) error {
	for _, ver := range SupportedJavaMajors() {
		bin, _ := resolve(ver)
		if bin != "" && detect(bin) == ver {
			continue
		}
		info, err := install(ver)
		if err != nil {
			return fmt.Errorf("required Java %s is unavailable: %w", ver, err)
		}
		if !info.Installed || info.Version != ver || detect(info.Path) != ver {
			return fmt.Errorf("required Java %s failed exact-major verification", ver)
		}
	}
	return nil
}

func extractTarGzStripped(archivePath, destDir string) error {
	return extractJavaTarGz(context.Background(), archivePath, destDir)
}

// Extract a single stripped top-level directory. Internal file links used by
// older JREs are materialized as bounded copies, never filesystem links. Escapes,
// directory links, link cycles and special files are rejected.
func extractJavaTarGz(ctx context.Context, archivePath, destDir string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}
	root, err := os.OpenRoot(destDir)
	if err != nil {
		return err
	}
	defer root.Close()
	// Also cap decompressed metadata/padding, not only regular file contents.
	limited := &io.LimitedReader{R: gz, N: javaExpandedLimit + javaZIPMetadataLimit + 1}
	tr := tar.NewReader(limited)
	top := ""
	seen := make(map[string]bool)
	var expanded int64
	type fileLink struct{ name, target string }
	var links []fileLink
	regular := make(map[string]bool)
	for count := 0; ; count++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if count >= javaEntryLimit {
			return fmt.Errorf("Java archive entry limit exceeded")
		}
		name := strings.TrimSuffix(h.Name, "/")
		if strings.ContainsAny(name, "\\:\x00") {
			return fmt.Errorf("unsafe Java archive path %q", h.Name)
		}
		parts := strings.Split(name, "/")
		for _, part := range parts {
			if part == "" || part == "." || part == ".." {
				return fmt.Errorf("unsafe Java archive path %q", h.Name)
			}
		}
		if top == "" {
			top = parts[0]
		}
		if top != parts[0] {
			return fmt.Errorf("multiple Java archive roots")
		}
		if h.Typeflag != tar.TypeDir && h.Typeflag != tar.TypeReg && h.Typeflag != tar.TypeRegA && h.Typeflag != tar.TypeSymlink && h.Typeflag != tar.TypeLink {
			return fmt.Errorf("Java archive special files are not allowed: %q", h.Name)
		}
		if len(parts) == 1 {
			if h.Typeflag != tar.TypeDir {
				return fmt.Errorf("Java archive lacks top-level directory")
			}
			continue
		}
		rel := filepath.Join(parts[1:]...)
		if !filepath.IsLocal(rel) || seen[rel] {
			return fmt.Errorf("unsafe or duplicate Java archive path %q", h.Name)
		}
		seen[rel] = true
		if h.Typeflag == tar.TypeSymlink || h.Typeflag == tar.TypeLink {
			if h.Linkname == "" || strings.ContainsAny(h.Linkname, "\\:\x00") || strings.HasPrefix(h.Linkname, "/") {
				return fmt.Errorf("unsafe Java archive link %q", h.Name)
			}
			var target string
			if h.Typeflag == tar.TypeLink {
				if !strings.HasPrefix(h.Linkname, top+"/") {
					return fmt.Errorf("Java hardlink outside archive root")
				}
				target = filepath.Clean(filepath.FromSlash(strings.TrimPrefix(h.Linkname, top+"/")))
			} else {
				target = filepath.Join(filepath.Dir(rel), filepath.FromSlash(h.Linkname))
			}
			if !filepath.IsLocal(target) {
				return fmt.Errorf("escaping Java archive link %q", h.Name)
			}
			links = append(links, fileLink{rel, target})
			continue
		}
		if h.Typeflag == tar.TypeDir {
			if err := root.MkdirAll(rel, 0755); err != nil {
				return err
			}
			continue
		}
		if h.Size < 0 || h.Size > javaDownloadLimit || expanded > javaExpandedLimit-h.Size {
			return fmt.Errorf("Java archive expanded size limit exceeded")
		}
		expanded += h.Size
		if err := root.MkdirAll(filepath.Dir(rel), 0755); err != nil {
			return err
		}
		mode := os.FileMode(0644) | (os.FileMode(h.Mode) & 0111)
		out, err := root.OpenFile(rel, os.O_CREATE|os.O_EXCL|os.O_WRONLY, mode)
		if err != nil {
			return err
		}
		_, copyErr := io.CopyN(out, tr, h.Size)
		closeErr := out.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeErr != nil {
			return closeErr
		}
		regular[rel] = true
	}
	// Drain to validate gzip CRC, within the same decompression budget.
	if _, err := io.Copy(io.Discard, limited); err != nil {
		return err
	}
	if limited.N <= 0 {
		return fmt.Errorf("Java archive decompression limit exceeded")
	}
	// Resolve only to files from this archive. Copying instead of creating links
	// also prevents symlink-chain tricks and keeps publication safe after rename.
	for len(links) > 0 {
		pending := make([]fileLink, 0, len(links))
		for _, link := range links {
			if err := ctx.Err(); err != nil {
				return err
			}
			if !regular[link.target] {
				pending = append(pending, link)
				continue
			}
			in, err := root.Open(link.target)
			if err != nil {
				return err
			}
			st, err := in.Stat()
			if err != nil {
				in.Close()
				return err
			}
			if !st.Mode().IsRegular() || st.Size() > javaDownloadLimit || expanded > javaExpandedLimit-st.Size() {
				in.Close()
				return fmt.Errorf("invalid or excessive Java archive file link")
			}
			expanded += st.Size()
			if err := root.MkdirAll(filepath.Dir(link.name), 0755); err != nil {
				in.Close()
				return err
			}
			out, err := root.OpenFile(link.name, os.O_CREATE|os.O_EXCL|os.O_WRONLY, st.Mode().Perm())
			if err != nil {
				in.Close()
				return err
			}
			_, copyErr := io.CopyN(out, in, st.Size())
			in.Close()
			closeErr := out.Close()
			if copyErr != nil {
				return copyErr
			}
			if closeErr != nil {
				return closeErr
			}
			regular[link.name] = true
		}
		if len(pending) == len(links) {
			return fmt.Errorf("Java archive has unresolved/cyclic or directory links")
		}
		links = pending
	}
	// The Linux installer commonly runs as root for an unprivileged service.
	// Do not let a restrictive installer umask leave published runtimes private.
	return fs.WalkDir(root.FS(), ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if entry.IsDir() {
			return root.Chmod(name, 0755)
		}
		st, err := entry.Info()
		if err != nil {
			return err
		}
		return root.Chmod(name, 0644|(st.Mode().Perm()&0111))
	})
}

func stripFirstPathSegment(p string) string {
	trimmed := strings.TrimPrefix(filepath.ToSlash(p), "./")
	trimmed = strings.TrimPrefix(trimmed, "/")
	idx := strings.IndexByte(trimmed, '/')
	if idx < 0 {
		return ""
	}
	return trimmed[idx+1:]
}
