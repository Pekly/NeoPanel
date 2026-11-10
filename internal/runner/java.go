package runner

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"
)

type JavaRuntimeInfo struct {
	Version   string `json:"version"`    // "25", "21", "17", "8"
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
	{Version: "17", Name: "Java 17 (LTS • Minecraft 1.17 - 1.20.4)"},
	{Version: "8", Name: "Java 8 (Legacy • Minecraft 1.8 - 1.16.5)"},
}

var (
	javaInstallMu sync.Mutex
	javaVersionRe = regexp.MustCompile(`"([^"]+)"`)
	javaSupported = map[string]bool{"25": true, "21": true, "17": true, "8": true}
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
	case "25", "21", "17", "11", "8":
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
	for _, ver := range []string{"25", "21", "17", "11", "8"} {
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
	if !isExecutableBinary(binPath) {
		return ""
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2500*time.Millisecond)
	defer cancel()

	out, err := exec.CommandContext(ctx, binPath, "-version").CombinedOutput()
	if err != nil && len(out) == 0 {
		return ""
	}
	return parseJavaVersionOutput(string(out))
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
				if maj := detectJavaBinaryMajor(cand); maj == "" || maj == targetTag {
					return cand, "OpenJDK " + targetTag
				}
			}
		}
		if dynamicPath := scanLinuxJVMsForMajor(targetTag); dynamicPath != "" {
			return dynamicPath, "OpenJDK " + targetTag
		}
	}

	// 4. Windows fallback paths for local development
	if runtime.GOOS == "windows" {
		candidates := getWindowsJavaCandidates(targetTag)
		for _, cand := range candidates {
			if isExecutableBinary(cand) {
				if maj := detectJavaBinaryMajor(cand); maj == "" || maj == targetTag {
					return cand, "Java " + targetTag
				}
			}
		}
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

	// Return canonical expected path (non-existent) so Start() surfaces a clear "Java Runtime Not Found" error
	// instead of silently running e.g. a Java 25 server under Java 21.
	if runtime.GOOS == "linux" && javaSupported[targetTag] {
		return "/usr/lib/jvm/java-" + targetTag + "-temurin/bin/java", "Java " + targetTag + " (Not Installed)"
	}

	if lp, err := exec.LookPath("java"); err == nil && !javaSupported[targetTag] {
		return lp, "System Default Java"
	}

	return "java-" + targetTag, "Java " + targetTag + " (Not Installed)"
}

func scanLinuxJVMsForMajor(targetTag string) string {
	roots := []string{"/usr/lib/jvm", "/opt/java", "/usr/java"}
	for _, root := range roots {
		entries, err := os.ReadDir(root)
		if err != nil {
			continue
		}
		for _, e := range entries {
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
				if matchJavaMajorInString(e.Name()) == targetTag || detectJavaBinaryMajor(javaBin) == targetTag {
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
	programFiles := os.Getenv("ProgramFiles")
	if programFiles == "" {
		programFiles = `C:\Program Files`
	}

	cands := []string{
		filepath.Join(programFiles, "Eclipse Adoptium", "jdk-"+tag, "bin", "java.exe"),
		filepath.Join(programFiles, "Eclipse Adoptium", "jre-"+tag, "bin", "java.exe"),
		filepath.Join(programFiles, "Java", "jdk-"+tag, "bin", "java.exe"),
		filepath.Join(programFiles, "Java", "jre-"+tag, "bin", "java.exe"),
		filepath.Join(programFiles, "Microsoft", "jdk-"+tag, "bin", "java.exe"),
	}

	for _, dir := range []string{
		filepath.Join(programFiles, "Eclipse Adoptium"),
		filepath.Join(programFiles, "Java"),
		filepath.Join(programFiles, "Microsoft"),
	} {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() && matchJavaMajorInString(e.Name()) == tag {
				cands = append(cands, filepath.Join(dir, e.Name(), "bin", "java.exe"))
			}
		}
	}

	return cands
}

// DetectInstalledJavas returns a list of installed Java runtimes found on the machine.
func DetectInstalledJavas() []JavaRuntimeInfo {
	byVersion := make(map[string]JavaRuntimeInfo)
	var extraResults []JavaRuntimeInfo
	seenPaths := make(map[string]bool)

	record := func(ver, name, path string, isDefault bool) {
		if path == "" || seenPaths[path] {
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
		for _, root := range []string{"/usr/lib/jvm", "/opt/java"} {
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

					maj := matchJavaMajorInString(dirName)
					if maj == "" {
						maj = detectJavaBinaryMajor(javaBin)
					}

					switch maj {
					case "25", "21", "17", "8":
						record(maj, premadeDisplayName(maj), javaBin, false)
					case "":
						record(dirName, dirName, javaBin, false)
					default:
						record(maj, "OpenJDK "+maj+" ("+dirName+")", javaBin, false)
					}
				}
			}
		}
	} else if runtime.GOOS == "windows" {
		for _, def := range canonicalPremadeJavas {
			for _, cand := range getWindowsJavaCandidates(def.Version) {
				if isExecutableBinary(cand) {
					record(def.Version, def.Name, cand, false)
					break
				}
			}
		}
	}

	// 2. Default system Java (and also backfill major version if e.g. system java was installed outside /usr/lib/jvm)
	if lp, err := exec.LookPath("java"); err == nil {
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

// GetPremadeJavaStatuses returns the canonical latest premade Java runtimes (25, 21, 17, 8)
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

// InstallJavaRuntime downloads and installs an official Eclipse Temurin Java runtime (25, 21, 17, or 8)
// into /usr/lib/jvm/java-<ver>-temurin on Linux hosts.
func InstallJavaRuntime(version string) (JavaRuntimeInfo, error) {
	ver := strings.TrimSpace(version)
	if !javaSupported[ver] {
		return JavaRuntimeInfo{}, fmt.Errorf("unsupported Java version '%s' (supported: 25, 21, 17, 8)", version)
	}
	if runtime.GOOS != "linux" {
		return JavaRuntimeInfo{}, fmt.Errorf("automatic runtime installation is only supported on Linux hosts")
	}

	javaInstallMu.Lock()
	defer javaInstallMu.Unlock()

	for _, st := range GetPremadeJavaStatuses() {
		if st.Version == ver && st.Installed {
			return st, nil
		}
	}

	arch := "x64"
	switch runtime.GOARCH {
	case "amd64":
		arch = "x64"
	case "arm64":
		arch = "aarch64"
	default:
		return JavaRuntimeInfo{}, fmt.Errorf("unsupported CPU architecture for auto-install: %s", runtime.GOARCH)
	}

	urls := []string{
		fmt.Sprintf("https://api.adoptium.net/v3/binary/latest/%s/ga/linux/%s/jre/hotspot/normal/eclipse", ver, arch),
		fmt.Sprintf("https://api.adoptium.net/v3/binary/latest/%s/ga/linux/%s/jdk/hotspot/normal/eclipse", ver, arch),
		fmt.Sprintf("https://api.adoptium.net/v3/binary/latest/%s/ea/linux/%s/jre/hotspot/normal/eclipse", ver, arch),
		fmt.Sprintf("https://api.adoptium.net/v3/binary/latest/%s/ea/linux/%s/jdk/hotspot/normal/eclipse", ver, arch),
	}

	_ = os.MkdirAll("/usr/lib/jvm", 0755)
	destDir := fmt.Sprintf("/usr/lib/jvm/java-%s-temurin", ver)
	stageDir := fmt.Sprintf("/usr/lib/jvm/.java-%s-temurin.stage-%d", ver, time.Now().UnixNano())
	defer os.RemoveAll(stageDir)

	client := &http.Client{Timeout: 5 * time.Minute}
	var lastErr error

	for _, dlURL := range urls {
		tmpTar, err := os.CreateTemp("", "neopanel-java-*.tar.gz")
		if err != nil {
			return JavaRuntimeInfo{}, err
		}
		tmpTarPath := tmpTar.Name()

		req, err := http.NewRequest("GET", dlURL, nil)
		if err != nil {
			tmpTar.Close()
			_ = os.Remove(tmpTarPath)
			lastErr = err
			continue
		}
		req.Header.Set("User-Agent", "NeoPanel/1.0 (+https://github.com/Pekly/NeoPanel)")

		resp, err := client.Do(req)
		if err != nil {
			tmpTar.Close()
			_ = os.Remove(tmpTarPath)
			lastErr = err
			continue
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			tmpTar.Close()
			_ = os.Remove(tmpTarPath)
			lastErr = fmt.Errorf("HTTP %d from %s", resp.StatusCode, dlURL)
			continue
		}

		_, copyErr := io.Copy(tmpTar, resp.Body)
		resp.Body.Close()
		tmpTar.Close()
		if copyErr != nil {
			_ = os.Remove(tmpTarPath)
			lastErr = copyErr
			continue
		}

		_ = os.RemoveAll(stageDir)
		if err := extractTarGzStripped(tmpTarPath, stageDir); err != nil {
			_ = os.Remove(tmpTarPath)
			lastErr = err
			continue
		}
		_ = os.Remove(tmpTarPath)

		javaBin := filepath.Join(stageDir, "bin", "java")
		if !isExecutableBinary(javaBin) {
			javaBin = filepath.Join(stageDir, "jre", "bin", "java")
		}
		if !isExecutableBinary(javaBin) {
			lastErr = fmt.Errorf("extracted archive did not contain bin/java")
			continue
		}
		if maj := detectJavaBinaryMajor(javaBin); maj == "" {
			lastErr = fmt.Errorf("downloaded java binary failed execution check")
			continue
		}

		// Replace any empty or stale broken destination before moving staging dir into place
		_ = os.RemoveAll(destDir)
		if err := os.Rename(stageDir, destDir); err != nil {
			return JavaRuntimeInfo{}, fmt.Errorf("failed to finalize Java %s directory: %w", ver, err)
		}

		finalBin := filepath.Join(destDir, "bin", "java")
		if !isExecutableBinary(finalBin) {
			finalBin = filepath.Join(destDir, "jre", "bin", "java")
		}
		return JavaRuntimeInfo{
			Version:   ver,
			Name:      premadeDisplayName(ver),
			Path:      finalBin,
			Installed: true,
		}, nil
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("no valid Adoptium build available")
	}
	return JavaRuntimeInfo{}, fmt.Errorf("failed to install Java %s: %w", ver, lastErr)
}

func extractTarGzStripped(archivePath, destDir string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("invalid gzip stream: %w", err)
	}
	defer gzr.Close()

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}
	cleanDest := filepath.Clean(destDir)

	type pendingHardLink struct {
		target string
		linkTo string
	}
	var hardLinks []pendingHardLink

	tr := tar.NewReader(gzr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		rel := stripFirstPathSegment(hdr.Name)
		if rel == "" || rel == "." {
			continue
		}
		cleanRel := filepath.Clean(filepath.FromSlash(rel))
		if cleanRel == ".." || strings.HasPrefix(cleanRel, ".."+string(os.PathSeparator)) || filepath.IsAbs(cleanRel) {
			return fmt.Errorf("unsafe tar path: %s", hdr.Name)
		}
		targetPath := filepath.Join(cleanDest, cleanRel)

		switch hdr.Typeflag {
		case tar.TypeDir:
			mode := os.FileMode(hdr.Mode) & 0777
			if mode == 0 {
				mode = 0755
			}
			if err := os.MkdirAll(targetPath, mode|0700); err != nil {
				return err
			}
		case tar.TypeReg, tar.TypeRegA:
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return err
			}
			mode := os.FileMode(hdr.Mode) & 0777
			if mode == 0 {
				mode = 0644
			}
			out, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
			if err != nil {
				return err
			}
			if _, err := io.Copy(out, tr); err != nil {
				out.Close()
				return err
			}
			out.Close()
			_ = os.Chmod(targetPath, mode)
		case tar.TypeSymlink:
			if filepath.IsAbs(hdr.Linkname) {
				return fmt.Errorf("unsafe absolute symlink in archive: %s -> %s", hdr.Name, hdr.Linkname)
			}
			resolvedSym := filepath.Clean(filepath.Join(filepath.Dir(targetPath), filepath.FromSlash(hdr.Linkname)))
			if resolvedSym != cleanDest && !strings.HasPrefix(resolvedSym, cleanDest+string(os.PathSeparator)) {
				return fmt.Errorf("unsafe escaping symlink in archive: %s -> %s", hdr.Name, hdr.Linkname)
			}
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return err
			}
			_ = os.Remove(targetPath)
			if err := os.Symlink(hdr.Linkname, targetPath); err != nil {
				return err
			}
		case tar.TypeLink:
			linkRel := stripFirstPathSegment(hdr.Linkname)
			if linkRel != "" && linkRel != "." {
				cleanLinkRel := filepath.Clean(filepath.FromSlash(linkRel))
				if !strings.HasPrefix(cleanLinkRel, "..") && !filepath.IsAbs(cleanLinkRel) {
					hardLinks = append(hardLinks, pendingHardLink{
						target: targetPath,
						linkTo: filepath.Join(cleanDest, cleanLinkRel),
					})
				}
			}
		}
	}

	for _, hl := range hardLinks {
		_ = os.MkdirAll(filepath.Dir(hl.target), 0755)
		_ = os.Remove(hl.target)
		if err := os.Link(hl.linkTo, hl.target); err != nil {
			return err
		}
	}

	return nil
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
