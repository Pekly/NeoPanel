package runner

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type JavaRuntimeInfo struct {
	Version   string `json:"version"`    // "21", "17", "8"
	Name      string `json:"name"`       // "OpenJDK 21"
	Path      string `json:"path"`       // "/usr/lib/jvm/java-21-openjdk-amd64/bin/java"
	IsDefault bool   `json:"is_default"` // true if default system java
}

// ResolveJavaBinary converts a version string ("21", "17", "8", etc.) or path
// into a validated executable path on the host system.
func ResolveJavaBinary(versionOrPath string) (binaryPath string, displayLabel string) {
	v := strings.TrimSpace(versionOrPath)

	// 1. Direct path check (user provided full binary path)
	if v != "" && !strings.EqualFold(v, "default") && !strings.EqualFold(v, "java") {
		if filepath.IsAbs(v) {
			if fi, err := os.Stat(v); err == nil && !fi.IsDir() {
				return v, "Custom Java (" + filepath.Base(v) + ")"
			}
		}
	}

	// 2. Identify target version tag
	targetTag := "21"
	switch {
	case strings.Contains(v, "8") || strings.Contains(v, "1.8"):
		targetTag = "8"
	case strings.Contains(v, "17"):
		targetTag = "17"
	case strings.Contains(v, "21"):
		targetTag = "21"
	case strings.Contains(v, "25"):
		targetTag = "25"
	case strings.Contains(v, "11"):
		targetTag = "11"
	case v == "" || strings.EqualFold(v, "default") || strings.EqualFold(v, "java"):
		targetTag = "default"
	default:
		targetTag = v
	}

	// If explicit default requested, return system "java"
	if targetTag == "default" {
		if lp, err := exec.LookPath("java"); err == nil {
			return lp, "System Default Java"
		}
		return "java", "System Default Java"
	}

	// 3. Search standard Linux JVM installations (/usr/lib/jvm)
	if runtime.GOOS == "linux" {
		candidates := getLinuxJavaCandidates(targetTag)
		for _, cand := range candidates {
			if fi, err := os.Stat(cand); err == nil && !fi.IsDir() {
				return cand, "OpenJDK " + targetTag
			}
		}
	}

	// 4. Windows fallback paths for local development
	if runtime.GOOS == "windows" {
		candidates := getWindowsJavaCandidates(targetTag)
		for _, cand := range candidates {
			if fi, err := os.Stat(cand); err == nil && !fi.IsDir() {
				return cand, "Java " + targetTag
			}
		}
	}

	// 5. Fallback to system PATH java
	if lp, err := exec.LookPath("java"); err == nil {
		return lp, "System Default Java (Fallback for " + targetTag + ")"
	}

	return "java", "Java (Default)"
}

func getLinuxJavaCandidates(tag string) []string {
	archs := []string{"amd64", "arm64", "aarch64", "i386"}
	var candidates []string

	if tag == "8" {
		// OpenJDK 8 often has jre/bin/java or bin/java
		for _, arch := range archs {
			candidates = append(candidates,
				"/usr/lib/jvm/java-8-openjdk-"+arch+"/jre/bin/java",
				"/usr/lib/jvm/java-8-openjdk-"+arch+"/bin/java",
				"/usr/lib/jvm/java-1.8.0-openjdk-"+arch+"/jre/bin/java",
				"/usr/lib/jvm/java-1.8.0-openjdk-"+arch+"/bin/java",
			)
		}
		candidates = append(candidates,
			"/usr/lib/jvm/java-8-openjdk/jre/bin/java",
			"/usr/lib/jvm/java-8-openjdk/bin/java",
			"/usr/lib/jvm/default-java/bin/java",
		)
		return candidates
	}

	for _, arch := range archs {
		candidates = append(candidates,
			"/usr/lib/jvm/java-"+tag+"-openjdk-"+arch+"/bin/java",
			"/usr/lib/jvm/java-"+tag+"-openjdk/bin/java",
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

	return []string{
		filepath.Join(programFiles, "Eclipse Adoptium", "jdk-"+tag, "bin", "java.exe"),
		filepath.Join(programFiles, "Java", "jdk-"+tag, "bin", "java.exe"),
		filepath.Join(programFiles, "Java", "jre-"+tag, "bin", "java.exe"),
		filepath.Join(programFiles, "Microsoft", "jdk-"+tag, "bin", "java.exe"),
	}
}

// DetectInstalledJavas returns a list of installed Java runtimes found on the machine.
func DetectInstalledJavas() []JavaRuntimeInfo {
	var results []JavaRuntimeInfo
	seen := make(map[string]bool)

	checkAndAdd := func(ver, name, path string, isDefault bool) {
		if path == "" || seen[path] {
			return
		}
		if fi, err := os.Stat(path); err == nil && !fi.IsDir() {
			seen[path] = true
			results = append(results, JavaRuntimeInfo{
				Version:   ver,
				Name:      name,
				Path:      path,
				IsDefault: isDefault,
			})
		}
	}

	// 1. Detect Linux /usr/lib/jvm directories
	if runtime.GOOS == "linux" {
		entries, err := os.ReadDir("/usr/lib/jvm")
		if err == nil {
			for _, e := range entries {
				if !e.IsDir() {
					continue
				}
				dirName := e.Name()
				javaBin := filepath.Join("/usr/lib/jvm", dirName, "bin", "java")
				if _, err := os.Stat(javaBin); err != nil {
					// Check jre/bin/java (common in Java 8)
					javaBin = filepath.Join("/usr/lib/jvm", dirName, "jre", "bin", "java")
					if _, err := os.Stat(javaBin); err != nil {
						continue
					}
				}

				switch {
				case strings.Contains(dirName, "-21-") || strings.HasPrefix(dirName, "java-21"):
					checkAndAdd("21", "OpenJDK 21 (Modern 1.20.5+)", javaBin, false)
				case strings.Contains(dirName, "-17-") || strings.HasPrefix(dirName, "java-17"):
					checkAndAdd("17", "OpenJDK 17 (Legacy 1.18 - 1.20.4)", javaBin, false)
				case strings.Contains(dirName, "-8-") || strings.HasPrefix(dirName, "java-8") || strings.Contains(dirName, "1.8.0"):
					checkAndAdd("8", "OpenJDK 8 (Legacy 1.8 - 1.12.2)", javaBin, false)
				case strings.Contains(dirName, "-25-"):
					checkAndAdd("25", "OpenJDK 25 (Cutting edge)", javaBin, false)
				default:
					checkAndAdd(dirName, dirName, javaBin, false)
				}
			}
		}
	}

	// 2. Default system Java
	if lp, err := exec.LookPath("java"); err == nil {
		checkAndAdd("default", "System Default (PATH)", lp, true)
	}

	if len(results) == 0 {
		results = append(results, JavaRuntimeInfo{
			Version:   "21",
			Name:      "System Java",
			Path:      "java",
			IsDefault: true,
		})
	}

	return results
}
