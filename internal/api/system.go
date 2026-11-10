package api

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/Pekly/NeoPanel/internal/database"
	"github.com/Pekly/NeoPanel/internal/runner"
)

var startTime = time.Now()

type SystemInfo struct {
	OS          string `json:"os"`
	Arch        string `json:"arch"`
	CPUs        int    `json:"cpus"`
	TotalRAMMB  int64  `json:"total_ram_mb"`
	FreeRAMMB   int64  `json:"free_ram_mb"`
	UptimeHours int64  `json:"uptime_hours"`
	Version     string `json:"version"`
	GoVersion   string `json:"go_version"`
	Goroutines  int    `json:"goroutines"`
}

func collectSystemInfo() SystemInfo {
	info := SystemInfo{
		OS:          runtime.GOOS,
		Arch:        runtime.GOARCH,
		CPUs:        runtime.NumCPU(),
		TotalRAMMB:  4096, // default fallback
		FreeRAMMB:   2048,
		UptimeHours: int64(time.Since(startTime).Hours()),
		Version:     "1.0.0",
		GoVersion:   runtime.Version(),
		Goroutines:  runtime.NumGoroutine(),
	}

	if runtime.GOOS == "linux" {
		readLinuxMemInfo(&info)
	}
	return info
}

func HandleSystemInfo(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, collectSystemInfo())
}

func HandleRunGC(w http.ResponseWriter, r *http.Request) {
	runtime.GC()
	writeJSON(w, http.StatusOK, map[string]string{"message": "Garbage collection executed successfully"})
}

func HandleGetJavaRuntimes(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("matrix") == "1" || r.URL.Query().Get("all") == "1" {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"installed":        runner.DetectInstalledJavas(),
			"premade":          runner.GetPremadeJavaStatuses(),
			"can_auto_install": runtime.GOOS == "linux",
		})
		return
	}
	javas := runner.DetectInstalledJavas()
	writeJSON(w, http.StatusOK, javas)
}

func HandleInstallJavaRuntime(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, _ := r.Context().Value(contextUserKey).(*JWTClaims)
		if claims == nil || claims.Role != "admin" {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "Admin privileges are required to install Java runtimes"})
			return
		}

		var req struct {
			Version string `json:"version"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON body"})
			return
		}

		info, err := runner.InstallJavaRuntime(req.Version)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}

		logAudit(db, r, "admin_java_install", fmt.Sprintf("Installed %s at %s", info.Name, info.Path))

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"message":          fmt.Sprintf("Java %s installed successfully", info.Version),
			"runtime":          info,
			"installed":        runner.DetectInstalledJavas(),
			"premade":          runner.GetPremadeJavaStatuses(),
			"can_auto_install": runtime.GOOS == "linux",
		})
	}
}

func readLinuxMemInfo(info *SystemInfo) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "MemTotal:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				val, _ := strconv.ParseInt(parts[1], 10, 64)
				info.TotalRAMMB = val / 1024
			}
		} else if strings.HasPrefix(line, "MemAvailable:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				val, _ := strconv.ParseInt(parts[1], 10, 64)
				info.FreeRAMMB = val / 1024
			}
		}
	}
}
