package api

import (
	"bufio"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

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

func HandleSystemInfo(w http.ResponseWriter, r *http.Request) {
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

	writeJSON(w, http.StatusOK, info)
}

func HandleRunGC(w http.ResponseWriter, r *http.Request) {
	runtime.GC()
	writeJSON(w, http.StatusOK, map[string]string{"message": "Garbage collection executed successfully"})
}

func HandleGetJavaRuntimes(w http.ResponseWriter, r *http.Request) {
	javas := runner.DetectInstalledJavas()
	writeJSON(w, http.StatusOK, javas)
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

