package runner

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

type StatsSnapshot struct {
	CPUPercent float64
	MemoryMB   int64
	NetInKiB   float64
	NetOutKiB  float64
}

type ProcessMetrics struct {
	mu           sync.Mutex
	PID          int
	CPUPercent   float64
	MemoryMB     int64
	NetInKiB     float64
	NetOutKiB    float64
	lastCPUTime  int64
	lastReadTime time.Time
	lastNetRx    uint64
	lastNetTx    uint64
	lastNetTime  time.Time
}

func NewProcessMetrics(pid int) *ProcessMetrics {
	pm := &ProcessMetrics{
		PID:          pid,
		lastReadTime: time.Now(),
	}
	if runtime.GOOS == "linux" && pid > 0 {
		// Seed initial CPU & Net baselines immediately
		_, _ = pm.readLinuxStatsLocked()
	}
	return pm
}

// ReadStats returns the current CPU %, Resident Memory (in MB), and Network In/Out (KiB/s).
func (pm *ProcessMetrics) ReadStats() (StatsSnapshot, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.PID <= 0 {
		return StatsSnapshot{}, nil
	}

	if runtime.GOOS == "linux" {
		return pm.readLinuxStatsLocked()
	}

	// Non-linux fallback (e.g. dev mode on Windows/macOS)
	return StatsSnapshot{
		CPUPercent: 0.5,
		MemoryMB:   256,
		NetInKiB:   1.2,
		NetOutKiB:  2.4,
	}, nil
}

func (pm *ProcessMetrics) readLinuxStatsLocked() (StatsSnapshot, error) {
	snap := StatsSnapshot{
		CPUPercent: pm.CPUPercent,
		MemoryMB:   pm.MemoryMB,
		NetInKiB:   pm.NetInKiB,
		NetOutKiB:  pm.NetOutKiB,
	}

	// 1. Read Memory from /proc/<pid>/status
	statusPath := fmt.Sprintf("/proc/%d/status", pm.PID)
	if file, err := os.Open(statusPath); err == nil {
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
		file.Close()
		snap.MemoryMB = memoryKB / 1024
		pm.MemoryMB = snap.MemoryMB
	} else {
		return snap, err
	}

	// 2. Read CPU ticks from /proc/<pid>/stat safely past closing ')' of comm field
	statPath := fmt.Sprintf("/proc/%d/stat", pm.PID)
	if statData, err := os.ReadFile(statPath); err == nil {
		statStr := string(statData)
		if endParen := strings.LastIndex(statStr, ")"); endParen != -1 && endParen+2 < len(statStr) {
			fields := strings.Fields(statStr[endParen+2:])
			// After (comm), field index 11 = utime (#14 overall), 12 = stime (#15 overall)
			if len(fields) >= 13 {
				utime, _ := strconv.ParseInt(fields[11], 10, 64)
				stime, _ := strconv.ParseInt(fields[12], 10, 64)
				totalCPUTime := utime + stime

				now := time.Now()
				elapsedSeconds := now.Sub(pm.lastReadTime).Seconds()

				if pm.lastCPUTime > 0 && elapsedSeconds >= 0.2 {
					deltaTicks := totalCPUTime - pm.lastCPUTime
					if deltaTicks < 0 {
						deltaTicks = 0
					}
					numCPU := float64(runtime.NumCPU())
					if numCPU < 1 {
						numCPU = 1
					}
					// Divide by numCPU so percentage represents total host CPU utilization (0 - 100%)
					cpuPct := ((float64(deltaTicks) / 100.0) / elapsedSeconds) * 100.0 / numCPU
					if cpuPct > 100.0 {
						cpuPct = 100.0
					}
					snap.CPUPercent = cpuPct
					pm.CPUPercent = cpuPct
					pm.lastCPUTime = totalCPUTime
					pm.lastReadTime = now
				} else if pm.lastCPUTime == 0 {
					pm.lastCPUTime = totalCPUTime
					pm.lastReadTime = now
				}
			}
		}
	}

	// 3. Read Network Inbound / Outbound from /proc/<pid>/net/dev (or /proc/net/dev fallback)
	rxTotal, txTotal, netOk := readProcNetDev(fmt.Sprintf("/proc/%d/net/dev", pm.PID))
	if !netOk {
		rxTotal, txTotal, netOk = readProcNetDev("/proc/net/dev")
	}
	if netOk {
		now := time.Now()
		if !pm.lastNetTime.IsZero() {
			elapsed := now.Sub(pm.lastNetTime).Seconds()
			if elapsed >= 0.2 && rxTotal >= pm.lastNetRx && txTotal >= pm.lastNetTx {
				snap.NetInKiB = float64(rxTotal-pm.lastNetRx) / 1024.0 / elapsed
				snap.NetOutKiB = float64(txTotal-pm.lastNetTx) / 1024.0 / elapsed
				pm.NetInKiB = snap.NetInKiB
				pm.NetOutKiB = snap.NetOutKiB
				pm.lastNetRx = rxTotal
				pm.lastNetTx = txTotal
				pm.lastNetTime = now
			}
		} else {
			pm.lastNetRx = rxTotal
			pm.lastNetTx = txTotal
			pm.lastNetTime = now
		}
	}

	return snap, nil
}

func readProcNetDev(path string) (uint64, uint64, bool) {
	f, err := os.Open(path)
	if err != nil {
		return 0, 0, false
	}
	defer f.Close()

	var totalRx, totalTx uint64
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		colonIdx := strings.Index(line, ":")
		if colonIdx <= 0 {
			continue
		}
		iface := strings.TrimSpace(line[:colonIdx])
		if iface == "lo" {
			continue
		}
		fields := strings.Fields(line[colonIdx+1:])
		if len(fields) >= 9 {
			rx, err1 := strconv.ParseUint(fields[0], 10, 64)
			tx, err2 := strconv.ParseUint(fields[8], 10, 64)
			if err1 == nil && err2 == nil {
				totalRx += rx
				totalTx += tx
			}
		}
	}
	return totalRx, totalTx, true
}
