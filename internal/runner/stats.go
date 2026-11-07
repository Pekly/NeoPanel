package runner

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type ProcessMetrics struct {
	PID          int
	CPUPercent   float64
	MemoryMB     int64
	lastCPUTime  int64
	lastReadTime time.Time
}

func NewProcessMetrics(pid int) *ProcessMetrics {
	return &ProcessMetrics{
		PID:          pid,
		lastReadTime: time.Now(),
	}
}

// ReadStats returns the current CPU % and Resident Memory (in MB).
func (pm *ProcessMetrics) ReadStats() (float64, int64, error) {
	if pm.PID <= 0 {
		return 0, 0, nil
	}

	if runtime.GOOS == "linux" {
		return pm.readLinuxStats()
	}

	// Non-linux fallback (e.g. dev mode on Windows/macOS)
	return 0.5, 256, nil
}

func (pm *ProcessMetrics) readLinuxStats() (float64, int64, error) {
	// 1. Read Memory from /proc/<pid>/status
	statusPath := fmt.Sprintf("/proc/%d/status", pm.PID)
	file, err := os.Open(statusPath)
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()

	var memoryKB int64
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "VmRSS:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				memoryKB, _ = strconv.ParseInt(fields[1], 10, 64)
			}
			break
		}
	}

	// 2. Read CPU ticks from /proc/<pid>/stat
	statPath := fmt.Sprintf("/proc/%d/stat", pm.PID)
	statData, err := os.ReadFile(statPath)
	if err != nil {
		return 0, memoryKB / 1024, nil
	}

	parts := strings.Fields(string(statData))
	if len(parts) < 15 {
		return 0, memoryKB / 1024, nil
	}

	utime, _ := strconv.ParseInt(parts[13], 10, 64)
	stime, _ := strconv.ParseInt(parts[14], 10, 64)
	totalCPUTime := utime + stime

	now := time.Now()
	elapsedSeconds := now.Sub(pm.lastReadTime).Seconds()

	var cpuPercent float64
	if pm.lastCPUTime > 0 && elapsedSeconds > 0 {
		// Linux USER_HZ is typically 100
		deltaTicks := totalCPUTime - pm.lastCPUTime
		cpuPercent = (float64(deltaTicks) / 100.0) / elapsedSeconds * 100.0
		if cpuPercent < 0 {
			cpuPercent = 0
		}
	}

	pm.lastCPUTime = totalCPUTime
	pm.lastReadTime = now

	return cpuPercent, memoryKB / 1024, nil
}

