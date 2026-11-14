package runner

import (
	"bufio"
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Pekly/NeoPanel/internal/database"
	"github.com/Pekly/NeoPanel/internal/downloader"
)

type OnlinePlayer struct {
	Name     string    `json:"name"`
	UUID     string    `json:"uuid,omitempty"`
	IP       string    `json:"ip,omitempty"`
	JoinedAt time.Time `json:"joined_at,omitempty"`
}

var (
	ansiEscapeRe      = regexp.MustCompile("\x1b\\[[0-9;?]*[A-Za-z]")
	rePlayerUUID      = regexp.MustCompile(`(?i)UUID of player ([A-Za-z0-9_.]{1,20}) is ([0-9a-fA-F-]{32,36})`)
	rePlayerLogin     = regexp.MustCompile(`([A-Za-z0-9_.]{1,20})\[/([0-9a-fA-F.:]+):\d+\]\s+logged in with entity id`)
	rePlayerJoined    = regexp.MustCompile(`(?:^|\]:\s+)([A-Za-z0-9_.]{1,20})\s+joined the game\b`)
	rePlayerLeft      = regexp.MustCompile(`(?:^|\]:\s+)([A-Za-z0-9_.]{1,20})\s+left the game\b`)
	rePlayerLost      = regexp.MustCompile(`(?:^|\]:\s+)([A-Za-z0-9_.]{1,20})\s+lost connection:`)
	reVelocityJoin    = regexp.MustCompile(`\[connected player\]\s+([A-Za-z0-9_.]{1,20})\s+\(/([0-9a-fA-F.:]+):\d+\)\s+has connected`)
	reVelocityLeave   = regexp.MustCompile(`\[connected player\]\s+([A-Za-z0-9_.]{1,20})(?:\s+\([^)]*\))?(?::\s*disconnected|\s+has disconnected)`)
	reBungeeJoin      = regexp.MustCompile(`(?:^|\s)\[([A-Za-z0-9_.]{1,20})(?:\|([0-9a-fA-F.:]+)(?::\d+)?)?\]\s+->\s+UpstreamBridge has connected`)
	reBungeeLeave     = regexp.MustCompile(`(?:^|\s)\[([A-Za-z0-9_.]{1,20})(?:\|[^\]]+)?\]\s+(?:<->|->)?\s*(?:UpstreamBridge|DownstreamBridge)\s+(?:has )?disconnected`)
	reVanillaListLine = regexp.MustCompile(`(?:^|\]:\s+)There are (\d+) of a max of \d+ players online:(.*)$`)
)

const (
	maxLogBuffer          = 1000
	DefaultStartupTimeout = 3 * time.Minute
	MaxStartupAttempts    = 2
	minStableUptime       = 20 * time.Second
)

type ProcessInstance struct {
	lifecycleMu      sync.Mutex
	runDone          chan struct{}
	owned            *ownedCommand
	runLock          *os.File
	cleanupErr       error
	mu               sync.RWMutex
	server           *database.Server
	workingDir       string
	cmd              *exec.Cmd
	stdin            io.WriteCloser
	status           string
	startedAt        time.Time
	metrics          *ProcessMetrics
	logBuffer        []string
	subscribers      map[chan string]struct{}
	onlinePlayers    map[string]*OnlinePlayer
	uuidCache        map[string]string
	zeroPlayersSince time.Time
	stopExpected     bool
	reachedOnline    bool
	timedOut         bool
	runID            uint64
	startupTimeout   time.Duration
	startupStopCh    chan struct{}
	onReady          func(srvID string)
	onExit           func(srvID string, unexpected bool, reachedOnline bool, timedOut bool, uptime time.Duration)
	cachedDiskMB     int64
	lastDiskCheck    time.Time
	diskScanning     bool
	isolationEngine  string
	tasksMax         int
}

func NewProcessInstance(server *database.Server, workingDir string, startupTimeout time.Duration, onReady func(string), onExit func(string, bool, bool, bool, time.Duration)) *ProcessInstance {
	if startupTimeout <= 0 {
		startupTimeout = DefaultStartupTimeout
	}
	return &ProcessInstance{
		server:         server,
		workingDir:     workingDir,
		status:         "offline",
		logBuffer:      make([]string, 0, maxLogBuffer),
		subscribers:    make(map[chan string]struct{}),
		onlinePlayers:  make(map[string]*OnlinePlayer),
		uuidCache:      make(map[string]string),
		startupTimeout: startupTimeout,
		onReady:        onReady,
		onExit:         onExit,
	}
}

func (p *ProcessInstance) SetStartupTimeout(d time.Duration) {
	if d <= 0 {
		return
	}
	p.mu.Lock()
	p.startupTimeout = d
	p.mu.Unlock()
}

func (p *ProcessInstance) SetIsolationConfig(engine string, tasksMax int) {
	p.mu.Lock()
	p.isolationEngine = engine
	p.tasksMax = tasksMax
	p.mu.Unlock()
}

func (p *ProcessInstance) stopStartupTimerLocked() {
	if p.startupStopCh != nil {
		close(p.startupStopCh)
		p.startupStopCh = nil
	}
}

func (p *ProcessInstance) GetStatus() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.status
}

func (p *ProcessInstance) GetStats() database.ServerStats {
	p.mu.RLock()
	srvID := p.server.ID
	memMax := p.server.MemoryMaxMB
	status := p.status
	startedAt := p.startedAt
	metrics := p.metrics
	diskMB := p.cachedDiskMB
	needsDiskScan := !p.diskScanning && (p.lastDiskCheck.IsZero() || time.Since(p.lastDiskCheck) > 20*time.Second)
	workDir := p.workingDir
	p.mu.RUnlock()

	stats := database.ServerStats{
		ServerID:    srvID,
		Status:      status,
		MemoryMaxMB: memMax,
		DiskMB:      diskMB,
	}

	if (status == "online" || status == "starting") && !startedAt.IsZero() {
		stats.UptimeSec = int64(time.Since(startedAt).Seconds())
		if metrics != nil {
			if snap, err := metrics.ReadStats(); err == nil {
				stats.CPUUsage = snap.CPUPercent
				stats.MemoryMB = snap.MemoryMB
				stats.NetInKiB = snap.NetInKiB
				stats.NetOutKiB = snap.NetOutKiB
			}
		}
	}

	if needsDiskScan {
		p.mu.Lock()
		if !p.diskScanning {
			p.diskScanning = true
			firstScan := diskMB == 0 && p.lastDiskCheck.IsZero()
			p.mu.Unlock()
			if firstScan {
				// First synchronous read before caching asynchronously thereafter
				size := calculateDirSizeMB(workDir)
				p.mu.Lock()
				p.cachedDiskMB = size
				p.lastDiskCheck = time.Now()
				p.diskScanning = false
				p.mu.Unlock()
				stats.DiskMB = size
			} else {
				go func() {
					size := calculateDirSizeMB(workDir)
					p.mu.Lock()
					p.cachedDiskMB = size
					p.lastDiskCheck = time.Now()
					p.diskScanning = false
					p.mu.Unlock()
				}()
			}
		} else {
			p.mu.Unlock()
		}
	}

	return stats
}

func (p *ProcessInstance) Start() error {
	p.lifecycleMu.Lock()
	defer p.lifecycleMu.Unlock()
	p.mu.Lock()
	if p.status != "offline" || p.cmd != nil {
		p.mu.Unlock()
		return fmt.Errorf("server is running or still stopping; wait for shutdown to complete")
	}

	runLock, err := acquireRunLock(p.workingDir)
	if err != nil {
		p.mu.Unlock()
		p.broadcast("[NeoPanel] Start refused: " + err.Error())
		return err
	}
	launched := false
	defer func() {
		if !launched {
			runLock.Close()
			p.mu.Lock()
			p.status = "offline"
			p.mu.Unlock()
		}
	}()
	p.stopStartupTimerLocked()
	p.status = "starting"
	p.stopExpected = false
	p.reachedOnline = false
	p.timedOut = false
	p.onlinePlayers = make(map[string]*OnlinePlayer)
	p.uuidCache = make(map[string]string)
	p.runID++
	currentRunID := p.runID
	srv := *p.server
	timeout := p.startupTimeout
	if timeout <= 0 {
		timeout = DefaultStartupTimeout
	}
	p.mu.Unlock()
	if err := checkServerWorldLocks(p.workingDir, srv.Type); err != nil {
		p.broadcast("[NeoPanel] Start refused: " + err.Error())
		return err
	}

	native := downloader.IsNativeSoftware(srv.Type)
	javaBinary, javaLabel := "", ""
	args := []string{}
	if native {
		javaBinary, err = steelExecutable(p.workingDir, srv.JarFile)
		if err != nil {
			p.broadcast("[NeoPanel] " + err.Error())
			return err
		}
		p.broadcast("[NeoPanel] Starting Steel native server (no Java runtime required)...")
	} else {
		javaBinary, javaLabel = ResolveJavaBinary(srv.JavaPath)
		p.broadcast(fmt.Sprintf("\x1b[36m[NeoPanel] Starting instance using %s (%s)...\x1b[0m", javaLabel, javaBinary))

		// Validate Java binary existence
		if !isExecutableBinary(javaBinary) {
			p.mu.Lock()
			p.status = "offline"
			p.mu.Unlock()

			p.broadcast("\x1b[31;1m========================================================================\x1b[0m")
			p.broadcast("\x1b[31;1m[NeoPanel ERROR] JAVA RUNTIME NOT FOUND!\x1b[0m")
			p.broadcast(fmt.Sprintf("\x1b[31mThe specified Java binary does not exist or is not executable on this host:\x1b[0m"))
			p.broadcast(fmt.Sprintf("  \x1b[33;1mPath:  %s\x1b[0m", javaBinary))
			p.broadcast(fmt.Sprintf("  \x1b[33;1mLabel: %s\x1b[0m", javaLabel))
			p.broadcast("\x1b[36m👉 Solution: Open Server -> Settings and choose or install a Java runtime (Java 25, 21, 17, or 8), or run 'sudo neopanel-installer --java' on your VPS.\x1b[0m")
			p.broadcast("\x1b[31;1m========================================================================\x1b[0m")
			return fmt.Errorf("Java binary '%s' (%s) was not found. Please install it or verify your Java settings in Settings", javaBinary, javaLabel)
		}

		jarPath := filepath.Join(p.workingDir, srv.JarFile)
		if fi, err := os.Stat(jarPath); os.IsNotExist(err) || (err == nil && fi.IsDir()) {
			p.mu.Lock()
			p.status = "offline"
			p.mu.Unlock()

			p.broadcast("\x1b[31;1m========================================================================\x1b[0m")
			p.broadcast("\x1b[31;1m[NeoPanel ERROR] SERVER JAR FILE NOT FOUND!\x1b[0m")
			p.broadcast("\x1b[31mCould not locate the executable JAR file in server directory:\x1b[0m")
			p.broadcast(fmt.Sprintf("  \x1b[33;1mExpected file: %s\x1b[0m", srv.JarFile))
			p.broadcast(fmt.Sprintf("  \x1b[33;1mFull path:     %s\x1b[0m", jarPath))
			p.broadcast("\x1b[36m👉 Solution: Click 'Reinstall Server' in Settings to cleanly download an official JAR from mcjarfiles.com, or upload your server JAR to the root folder via the File Manager.\x1b[0m")
			p.broadcast("\x1b[31;1m========================================================================\x1b[0m")
			return fmt.Errorf("server JAR file '%s' not found. Please upload it or use 'Reinstall Server'", srv.JarFile)
		}

		// Validate memory settings
		if srv.MemoryMinMB > 0 && srv.MemoryMaxMB > 0 && srv.MemoryMinMB > srv.MemoryMaxMB {
			p.mu.Lock()
			p.status = "offline"
			p.mu.Unlock()

			errMsg := fmt.Sprintf("Minimum RAM (%d MB) cannot exceed Maximum RAM (%d MB)", srv.MemoryMinMB, srv.MemoryMaxMB)
			p.broadcast(fmt.Sprintf("\x1b[31;1m[NeoPanel ERROR] INVALID RAM SETTINGS: %s\x1b[0m", errMsg))
			return fmt.Errorf("%s", errMsg)
		}

		args = []string{}
		if srv.MemoryMinMB > 0 {
			args = append(args, fmt.Sprintf("-Xms%dM", srv.MemoryMinMB))
		}
		if srv.MemoryMaxMB > 0 {
			args = append(args, fmt.Sprintf("-Xmx%dM", srv.MemoryMaxMB))
		}

		// Force full ANSI terminal color output even when piped in daemons/subprocesses
		args = append(args,
			"-Dterminal.ansi=true",
			"-Dterminal.jline=false",
			"-Dnet.minecrell.terminalconsole.ansi=true",
			"-Dlog4j.skipJansi=false",
		)

		// Add custom JVM flags if provided
		if strings.TrimSpace(srv.JavaArgs) != "" {
			customArgs := strings.Fields(srv.JavaArgs)
			args = append(args, customArgs...)
		}
	}

	// Ensure server game port is synchronized to configuration file
	port := srv.Port
	if port <= 0 {
		port = 25565
	}
	if native {
		if err := SyncSteelPort(p.workingDir, port); err != nil {
			return err
		}
	} else {
		syncServerPort(p.workingDir, srv.Type, port)
	}
	p.broadcast(fmt.Sprintf("\x1b[36m[NeoPanel] Configured game server port: %d (waiting for server to bind)\x1b[0m", port))

	if !native {
		args = append(args, "-jar", srv.JarFile, "nogui")

		// Pass CLI port parameter for game servers that support port override
		lowerType := strings.ToLower(srv.Type)
		if lowerType != "velocity" && lowerType != "bungeecord" && lowerType != "waterfall" {
			args = append(args, "--port", fmt.Sprintf("%d", port))
		}
	}

	cmd, isolationMode, err := p.buildCommand(javaBinary, args)
	if err != nil {
		p.broadcast("[NeoPanel] Isolation unavailable: " + err.Error())
		return err
	}
	configureProcessOwnership(cmd, runLock)
	p.broadcast(fmt.Sprintf("\x1b[35m[NeoPanel Security] Isolation: %s\x1b[0m", isolationMode))

	stdin, err := cmd.StdinPipe()
	if err != nil {
		p.mu.Lock()
		p.status = "offline"
		p.mu.Unlock()
		p.broadcast(fmt.Sprintf("\x1b[31;1m[NeoPanel ERROR] Failed to open stdin pipe: %v\x1b[0m", err))
		return fmt.Errorf("failed to open stdin pipe: %w", err)
	}

	stdout, stdoutWriter, err := os.Pipe()
	if err != nil {
		stdin.Close()
		p.mu.Lock()
		p.status = "offline"
		p.mu.Unlock()
		p.broadcast(fmt.Sprintf("\x1b[31;1m[NeoPanel ERROR] Failed to open stdout pipe: %v\x1b[0m", err))
		return fmt.Errorf("failed to open stdout pipe: %w", err)
	}

	stderr, stderrWriter, err := os.Pipe()
	if err != nil {
		stdin.Close()
		stdout.Close()
		stdoutWriter.Close()
		p.mu.Lock()
		p.status = "offline"
		p.mu.Unlock()
		p.broadcast(fmt.Sprintf("\x1b[31;1m[NeoPanel ERROR] Failed to open stderr pipe: %v\x1b[0m", err))
		return fmt.Errorf("failed to open stderr pipe: %w", err)
	}

	cmd.Stdout, cmd.Stderr = stdoutWriter, stderrWriter
	if err := cmd.Start(); err != nil {
		stdin.Close()
		stdout.Close()
		stderr.Close()
		stdoutWriter.Close()
		stderrWriter.Close()
		p.mu.Lock()
		p.status = "offline"
		p.mu.Unlock()

		p.broadcast("\x1b[31;1m========================================================================\x1b[0m")
		p.broadcast("\x1b[31;1m[NeoPanel ERROR] PROCESS LAUNCH FAILED!\x1b[0m")
		p.broadcast(fmt.Sprintf("\x1b[31mFailed to start server process: %v\x1b[0m", err))
		p.broadcast(fmt.Sprintf("  \x1b[33;1mCommand: %s %s\x1b[0m", javaBinary, strings.Join(args, " ")))
		p.broadcast("\x1b[31;1m========================================================================\x1b[0m")
		return fmt.Errorf("failed to start server process: %w", err)
	}
	stdoutWriter.Close()
	stderrWriter.Close()
	launched = true
	done := make(chan struct{})

	stopTimerCh := make(chan struct{})
	startedAt := time.Now()
	p.mu.Lock()
	p.cmd = cmd
	p.runDone = done
	p.owned = &ownedCommand{cmd: cmd}
	owned := p.owned
	p.runLock = runLock
	p.cleanupErr = nil
	p.stdin = stdin
	p.startedAt = startedAt
	p.metrics = NewProcessMetrics(cmd.Process.Pid)
	p.startupStopCh = stopTimerCh
	p.mu.Unlock()

	p.broadcast(fmt.Sprintf("\x1b[32m[NeoPanel] Server process started (PID: %d)...\x1b[0m", cmd.Process.Pid))

	// Watchdog timeout if the server stays stuck in "starting" without coming online
	if timeout > 0 {
		go func(runID uint64, ch <-chan struct{}, limit time.Duration) {
			timer := time.NewTimer(limit)
			defer timer.Stop()
			select {
			case <-ch:
				return
			case <-timer.C:
				p.mu.Lock()
				if p.runID != runID || p.status != "starting" || p.reachedOnline || p.stopExpected {
					p.mu.Unlock()
					return
				}
				p.timedOut = true
				proc := p.owned
				p.mu.Unlock()

				p.broadcast("\x1b[31;1m========================================================================\x1b[0m")
				p.broadcast(fmt.Sprintf("\x1b[31;1m[NeoPanel ERROR] SERVER STARTUP TIMED OUT (%s)!\x1b[0m", limit.Round(time.Second)))
				p.broadcast("\x1b[31mThe server did not finish starting before the timeout elapsed. Terminating stalled process...\x1b[0m")
				p.broadcast("\x1b[31;1m========================================================================\x1b[0m")

				if proc != nil {
					if err := proc.stop(false); err != nil {
						p.broadcast("[NeoPanel] Could not terminate startup process tree: " + err.Error())
					}
				}
			}
		}(currentRunID, stopTimerCh, timeout)
	}

	// Own the pipes: Wait must reap the launcher before draining child-held streams.
	var streamWg sync.WaitGroup
	streamWg.Add(2)
	go func() {
		defer streamWg.Done()
		p.streamOutput(stdout, currentRunID)
	}()
	go func() {
		defer streamWg.Done()
		p.streamOutput(stderr, currentRunID)
	}()

	// Wait for process exit in background
	go func(runID uint64, started time.Time) {
		waitErr, cleanupErr := owned.wait()
		if cleanupErr == nil {
			cleanupErr = ensureServerProcessesStopped(p.workingDir)
		}
		drained := make(chan struct{})
		go func() { streamWg.Wait(); close(drained) }()
		select {
		case <-drained:
		case <-time.After(2 * time.Second):
			stdout.Close()
			stderr.Close()
			<-drained
		}
		stdout.Close()
		stderr.Close()
		stdin.Close()

		p.mu.Lock()
		if p.runID != runID {
			p.mu.Unlock()
			return
		}
		p.stopStartupTimerLocked()
		p.cleanupErr = cleanupErr
		if cleanupErr == nil {
			runLock.Close()
			p.runLock = nil
			p.status = "offline"
			p.cmd = nil
		} else {
			// Keep ownership and block restarts when cleanup cannot be confirmed.
			p.status = "stopping"
		}
		p.stdin = nil
		close(done)
		p.onlinePlayers = make(map[string]*OnlinePlayer)
		expected := p.stopExpected
		reached := p.reachedOnline
		timedOut := p.timedOut
		onExit := p.onExit
		srvID := p.server.ID
		p.mu.Unlock()
		if cleanupErr != nil {
			p.broadcast("[NeoPanel] Process cleanup failed; restart blocked: " + cleanupErr.Error())
			return
		}

		uptime := time.Since(started)

		if timedOut {
			p.broadcast("\x1b[31;1m[NeoPanel] Startup attempt aborted due to startup timeout.\x1b[0m")
		} else if waitErr != nil && !expected {
			p.broadcast(fmt.Sprintf("\x1b[31;1m[NeoPanel] Server process terminated unexpectedly (%v).\x1b[0m", waitErr))
		} else if !expected && !reached {
			p.broadcast("\x1b[31;1m[NeoPanel] Server exited during startup before coming online.\x1b[0m")
		} else {
			p.broadcast("\x1b[33m[NeoPanel] Server process stopped.\x1b[0m")
		}

		if onExit != nil {
			onExit(srvID, !expected, reached, timedOut, uptime)
		}
	}(currentRunID, startedAt)

	return nil
}

func (p *ProcessInstance) streamOutput(r io.Reader, runID uint64) {
	p.mu.RLock()
	native := downloader.IsNativeSoftware(p.server.Type)
	p.mu.RUnlock()
	scanner := bufio.NewScanner(r)
	// Support very large single-line outputs (up to 1MB) without dropping stream
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if native {
			var keep bool
			line, keep = normalizeSteelConsoleLine(line)
			if !keep {
				continue
			}
		}

		// Detect when Minecraft / Proxy finishes startup
		var readyCallback func(string)
		var readySrvID string
		p.mu.Lock()
		if p.runID == runID && p.status == "starting" && serverReadyLine(p.server.Type, line) {
			p.status = "online"
			p.reachedOnline = true
			p.stopStartupTimerLocked()
			if len(p.onlinePlayers) == 0 {
				p.zeroPlayersSince = time.Now()
			}
			readyCallback = p.onReady
			readySrvID = p.server.ID
		}
		p.mu.Unlock()

		p.trackPlayerLine(line, runID)
		p.broadcast(line)
		if readyCallback != nil {
			readyCallback(readySrvID)
		}

		// Real-time runtime diagnostic detectors
		p.detectRuntimeDiagnostics(line)
	}
}

func (p *ProcessInstance) trackPlayerLine(rawLine string, runID uint64) {
	clean := strings.TrimSpace(ansiEscapeRe.ReplaceAllString(rawLine, ""))
	p.mu.RLock()
	native := downloader.IsNativeSoftware(p.server.Type)
	p.mu.RUnlock()
	if native {
		// Steel redraws its interactive prompt with carriage returns before logs.
		if i := strings.LastIndex(clean, "\r"); i >= 0 {
			clean = strings.TrimSpace(clean[i+1:])
		}
		clean = steelLogPrefix.ReplaceAllString(clean, "")
	}
	if clean == "" {
		return
	}

	if m := rePlayerUUID.FindStringSubmatch(clean); len(m) == 3 {
		name := strings.TrimSpace(m[1])
		uuid := strings.TrimSpace(m[2])
		low := strings.ToLower(name)
		p.mu.Lock()
		if p.runID == runID {
			if p.uuidCache == nil {
				p.uuidCache = make(map[string]string)
			}
			p.uuidCache[low] = uuid
			if existing, ok := p.onlinePlayers[low]; ok {
				existing.UUID = uuid
			}
		}
		p.mu.Unlock()
		return
	}

	if m := rePlayerLogin.FindStringSubmatch(clean); len(m) == 3 {
		name := strings.TrimSpace(m[1])
		ip := strings.TrimSpace(m[2])
		low := strings.ToLower(name)
		p.mu.Lock()
		if p.runID == runID {
			p.zeroPlayersSince = time.Time{}
			if p.onlinePlayers == nil {
				p.onlinePlayers = make(map[string]*OnlinePlayer)
			}
			p.onlinePlayers[low] = &OnlinePlayer{
				Name:     name,
				UUID:     p.uuidCache[low],
				IP:       ip,
				JoinedAt: time.Now().UTC(),
			}
		}
		p.mu.Unlock()
		return
	}

	if m := reVelocityJoin.FindStringSubmatch(clean); len(m) == 3 {
		name := strings.TrimSpace(m[1])
		ip := strings.TrimSpace(m[2])
		low := strings.ToLower(name)
		p.mu.Lock()
		if p.runID == runID {
			p.zeroPlayersSince = time.Time{}
			if p.onlinePlayers == nil {
				p.onlinePlayers = make(map[string]*OnlinePlayer)
			}
			p.onlinePlayers[low] = &OnlinePlayer{
				Name:     name,
				UUID:     p.uuidCache[low],
				IP:       ip,
				JoinedAt: time.Now().UTC(),
			}
		}
		p.mu.Unlock()
		return
	}

	if m := reBungeeJoin.FindStringSubmatch(clean); len(m) >= 2 {
		name := strings.TrimSpace(m[1])
		ip := ""
		if len(m) > 2 {
			ip = strings.TrimSpace(m[2])
		}
		low := strings.ToLower(name)
		p.mu.Lock()
		if p.runID == runID {
			p.zeroPlayersSince = time.Time{}
			if p.onlinePlayers == nil {
				p.onlinePlayers = make(map[string]*OnlinePlayer)
			}
			p.onlinePlayers[low] = &OnlinePlayer{
				Name:     name,
				UUID:     p.uuidCache[low],
				IP:       ip,
				JoinedAt: time.Now().UTC(),
			}
		}
		p.mu.Unlock()
		return
	}

	if m := rePlayerJoined.FindStringSubmatch(clean); len(m) == 2 {
		name := strings.TrimSpace(m[1])
		low := strings.ToLower(name)
		p.mu.Lock()
		if p.runID == runID {
			p.zeroPlayersSince = time.Time{}
			if p.onlinePlayers == nil {
				p.onlinePlayers = make(map[string]*OnlinePlayer)
			}
			if _, exists := p.onlinePlayers[low]; !exists {
				p.onlinePlayers[low] = &OnlinePlayer{
					Name:     name,
					UUID:     p.uuidCache[low],
					JoinedAt: time.Now().UTC(),
				}
			}
		}
		p.mu.Unlock()
		return
	}

	if m := rePlayerLeft.FindStringSubmatch(clean); len(m) == 2 {
		low := strings.ToLower(strings.TrimSpace(m[1]))
		p.mu.Lock()
		if p.runID == runID && p.onlinePlayers != nil {
			delete(p.onlinePlayers, low)
			if len(p.onlinePlayers) == 0 && p.status == "online" {
				p.zeroPlayersSince = time.Now()
			}
		}
		p.mu.Unlock()
		return
	}

	if m := rePlayerLost.FindStringSubmatch(clean); len(m) == 2 {
		low := strings.ToLower(strings.TrimSpace(m[1]))
		p.mu.Lock()
		if p.runID == runID && p.onlinePlayers != nil {
			delete(p.onlinePlayers, low)
			if len(p.onlinePlayers) == 0 && p.status == "online" {
				p.zeroPlayersSince = time.Now()
			}
		}
		p.mu.Unlock()
		return
	}

	if m := reVelocityLeave.FindStringSubmatch(clean); len(m) == 2 {
		low := strings.ToLower(strings.TrimSpace(m[1]))
		p.mu.Lock()
		if p.runID == runID && p.onlinePlayers != nil {
			delete(p.onlinePlayers, low)
			if len(p.onlinePlayers) == 0 && p.status == "online" {
				p.zeroPlayersSince = time.Now()
			}
		}
		p.mu.Unlock()
		return
	}

	if m := reBungeeLeave.FindStringSubmatch(clean); len(m) == 2 {
		low := strings.ToLower(strings.TrimSpace(m[1]))
		p.mu.Lock()
		if p.runID == runID && p.onlinePlayers != nil {
			delete(p.onlinePlayers, low)
			if len(p.onlinePlayers) == 0 && p.status == "online" {
				p.zeroPlayersSince = time.Now()
			}
		}
		p.mu.Unlock()
		return
	}

	if m := reVanillaListLine.FindStringSubmatch(clean); len(m) == 3 {
		cnt, _ := strconv.Atoi(m[1])
		rawNames := strings.TrimSpace(m[2])
		p.mu.Lock()
		if p.runID == runID {
			if cnt == 0 || rawNames == "" {
				p.onlinePlayers = make(map[string]*OnlinePlayer)
				if p.status == "online" {
					p.zeroPlayersSince = time.Now()
				}
			} else {
				p.zeroPlayersSince = time.Time{}
				nextMap := make(map[string]*OnlinePlayer)
				for _, part := range strings.Split(rawNames, ",") {
					nm := strings.TrimSpace(part)
					if idx := strings.IndexAny(nm, " (["); idx > 0 {
						nm = strings.TrimSpace(nm[:idx])
					}
					if nm == "" {
						continue
					}
					low := strings.ToLower(nm)
					if prev, ok := p.onlinePlayers[low]; ok {
						nextMap[low] = prev
					} else {
						nextMap[low] = &OnlinePlayer{
							Name:     nm,
							UUID:     p.uuidCache[low],
							JoinedAt: time.Now().UTC(),
						}
					}
				}
				p.onlinePlayers = nextMap
			}
		}
		p.mu.Unlock()
		return
	}
}

func (p *ProcessInstance) GetOnlinePlayers() []OnlinePlayer {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if len(p.onlinePlayers) == 0 {
		return []OnlinePlayer{}
	}
	res := make([]OnlinePlayer, 0, len(p.onlinePlayers))
	for _, pl := range p.onlinePlayers {
		res = append(res, *pl)
	}
	sort.Slice(res, func(i, j int) bool {
		if !res[i].JoinedAt.Equal(res[j].JoinedAt) {
			return res[i].JoinedAt.Before(res[j].JoinedAt)
		}
		return strings.ToLower(res[i].Name) < strings.ToLower(res[j].Name)
	})
	return res
}

func (p *ProcessInstance) detectRuntimeDiagnostics(line string) {
	lower := strings.ToLower(line)
	p.mu.RLock()
	srv := *p.server
	p.mu.RUnlock()
	if downloader.IsNativeSoftware(srv.Type) && strings.Contains(lower, "failed to bind to server port") &&
		(strings.Contains(lower, "address in use") || strings.Contains(lower, "address already in use")) {
		p.broadcast("\x1b[31;1m[NeoPanel DIAGNOSTIC] STEEL TCP PORT BIND FAILED\x1b[0m")
		p.broadcast(fmt.Sprintf("[NeoPanel] Steel could not bind 0.0.0.0:%d. Configuring a port in NeoPanel does not reserve or bind it.", srv.Port))
		p.broadcast(fmt.Sprintf("[NeoPanel] On this server's Linux node, inspect all TCP socket states: sudo ss -tanp '( sport = :%d )'", srv.Port))
		p.broadcast("[NeoPanel] Check IPv4 and IPv6 listeners and non-listening sockets, including TIME_WAIT. Identify the owner before stopping anything, or assign another unused port in Settings -> Network.")
		return
	}

	// 1. Java Version Mismatch / UnsupportedClassVersionError
	if strings.Contains(line, "UnsupportedClassVersionError") ||
		strings.Contains(lower, "has been compiled by a more recent version of the java runtime") ||
		strings.Contains(lower, "requires running the server with java") ||
		strings.Contains(lower, "unsupported major.minor version") ||
		(strings.Contains(lower, "requires java") && strings.Contains(lower, "or higher")) {
		p.broadcast("\x1b[31;1m========================================================================\x1b[0m")
		p.broadcast("\x1b[31;1m[NeoPanel DIAGNOSTIC] JAVA VERSION MISMATCH DETECTED!\x1b[0m")
		p.broadcast("\x1b[31mYour Minecraft server JAR was compiled for a newer Java runtime than the one currently selected.\x1b[0m")
		p.broadcast("\x1b[33;1m👉 Quick Fix: Go to Server -> Settings, change Java Runtime to Java 25 or Java 21, then restart.\x1b[0m")
		p.broadcast("\x1b[31;1m========================================================================\x1b[0m")
		return
	}

	// 2. Minecraft EULA not accepted
	if strings.Contains(lower, "agree to the eula") || strings.Contains(lower, "eula.txt") {
		p.broadcast("\x1b[31;1m========================================================================\x1b[0m")
		p.broadcast("\x1b[31;1m[NeoPanel DIAGNOSTIC] MOJANG EULA NOT ACCEPTED!\x1b[0m")
		p.broadcast("\x1b[31mThe Minecraft server requires accepting Mojang's End User License Agreement.\x1b[0m")
		p.broadcast("\x1b[33;1m👉 Quick Fix: Open 'eula.txt' in the File Manager and change 'eula=false' to 'eula=true'.\x1b[0m")
		p.broadcast("\x1b[31;1m========================================================================\x1b[0m")
		return
	}

	// 3. Port Bind / Address In Use
	if strings.Contains(line, "FAILED TO BIND TO PORT") ||
		strings.Contains(line, "BindException: Address already in use") ||
		strings.Contains(lower, "port already in use") {
		p.broadcast("\x1b[31;1m========================================================================\x1b[0m")
		p.broadcast("\x1b[31;1m[NeoPanel DIAGNOSTIC] PORT BIND CONFLICT (PORT IN USE)!\x1b[0m")
		p.broadcast(fmt.Sprintf("\x1b[31mGame port %d is already bound by another running server or service on this host.\x1b[0m", p.server.Port))
		p.broadcast("\x1b[33;1m👉 Quick Fix: Stop conflicting servers, or allocate a different port under Settings -> Network.\x1b[0m")
		p.broadcast("\x1b[31;1m========================================================================\x1b[0m")
		return
	}

	// 4. Out of Memory / Object Heap Allocation
	if strings.Contains(lower, "could not reserve enough space for") ||
		strings.Contains(lower, "invalid maximum heap size") ||
		strings.Contains(lower, "outofmemoryerror") ||
		strings.Contains(lower, "insufficient memory for the java runtime") {
		p.broadcast("\x1b[31;1m========================================================================\x1b[0m")
		p.broadcast("\x1b[31;1m[NeoPanel DIAGNOSTIC] OUT OF MEMORY (RAM ALLOCATION FAILED)!\x1b[0m")
		p.broadcast(fmt.Sprintf("\x1b[31mThe host machine cannot allocate the configured %d MB of RAM to Java.\x1b[0m", p.server.MemoryMaxMB))
		p.broadcast("\x1b[33;1m👉 Quick Fix: Lower your RAM allocation in Server -> Settings, or create a swap file on the host.\x1b[0m")
		p.broadcast("\x1b[31;1m========================================================================\x1b[0m")
		return
	}

	// 5. Corrupt or Invalid JAR
	if strings.Contains(lower, "invalid or corrupt jarfile") ||
		strings.Contains(lower, "unable to access jarfile") ||
		strings.Contains(lower, "zipexception: error in opening zip file") {
		p.broadcast("\x1b[31;1m========================================================================\x1b[0m")
		p.broadcast("\x1b[31;1m[NeoPanel DIAGNOSTIC] CORRUPT OR INVALID JAR FILE!\x1b[0m")
		p.broadcast(fmt.Sprintf("\x1b[31mThe server jar file '%s' cannot be executed because it is damaged or incomplete.\x1b[0m", p.server.JarFile))
		p.broadcast("\x1b[33;1m👉 Quick Fix: Click 'Reinstall Server' in Settings to cleanly re-download the JAR.\x1b[0m")
		p.broadcast("\x1b[31;1m========================================================================\x1b[0m")
		return
	}
}

func isExecutableBinary(path string) bool {
	if path == "" {
		return false
	}
	if filepath.IsAbs(path) {
		fi, err := os.Stat(path)
		return err == nil && !fi.IsDir()
	}
	_, err := exec.LookPath(path)
	return err == nil
}

func (p *ProcessInstance) WriteCommand(cmd string) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.stdin == nil || (p.status != "online" && p.status != "starting") {
		return fmt.Errorf("server is not running")
	}

	cleanCmd := strings.TrimSpace(cmd) + "\n"
	_, err := io.WriteString(p.stdin, cleanCmd)
	return err
}

func (p *ProcessInstance) WriteInput(cmd string) error {
	return p.WriteCommand(cmd)
}

func (p *ProcessInstance) Stop(graceful bool) error {
	p.lifecycleMu.Lock()
	defer p.lifecycleMu.Unlock()
	p.mu.Lock()
	if p.cmd == nil {
		err := p.cleanupErr
		p.mu.Unlock()
		return err
	}
	p.stopExpected = true
	p.stopStartupTimerLocked()
	p.status = "stopping"
	owned, stdin, done := p.owned, p.stdin, p.runDone
	p.mu.Unlock()
	if done == nil {
		return fmt.Errorf("server process completion is unavailable; restart refused")
	}
	select {
	case <-done:
		p.mu.RLock()
		cleanupErr := p.cleanupErr
		p.mu.RUnlock()
		if cleanupErr == nil {
			return nil
		}
		if err := owned.reconcile(p.workingDir); err != nil {
			return err
		}
		p.mu.Lock()
		if p.runLock != nil {
			p.runLock.Close()
			p.runLock = nil
		}
		p.cleanupErr, p.cmd, p.status = nil, nil, "offline"
		p.mu.Unlock()
		return nil
	default:
	}

	if graceful && stdin != nil {
		p.broadcast("[NeoPanel] Sending 'stop' command; waiting for the server to exit...")
		// A full stdin pipe must not prevent the shutdown timeout from firing.
		go func() { _, _ = io.WriteString(stdin, "stop\n") }()
		select {
		case <-done:
			p.mu.RLock()
			defer p.mu.RUnlock()
			return p.cleanupErr
		case <-time.After(15 * time.Second):
		}
	}
	p.broadcast("[NeoPanel] Terminating the owned server process tree...")
	if err := owned.stop(false); err != nil {
		return fmt.Errorf("could not terminate server process tree: %w", err)
	}
	select {
	case <-done:
		p.mu.RLock()
		defer p.mu.RUnlock()
		return p.cleanupErr
	case <-time.After(10 * time.Second):
		return fmt.Errorf("server has not exited; restart is blocked until shutdown completes")
	}
}

func (p *ProcessInstance) broadcast(msg string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Prevent consecutive duplicate lines (e.g. concurrent stdout & stderr duplication)
	if len(p.logBuffer) > 0 && p.logBuffer[len(p.logBuffer)-1] == msg {
		return
	}

	// Append to buffer
	if len(p.logBuffer) >= maxLogBuffer {
		p.logBuffer = p.logBuffer[1:]
	}
	p.logBuffer = append(p.logBuffer, msg)

	// Send non-blocking to active subscribers
	for ch := range p.subscribers {
		select {
		case ch <- msg:
		default:
		}
	}
}

func (p *ProcessInstance) GetLogs() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	logs := make([]string, len(p.logBuffer))
	copy(logs, p.logBuffer)
	return logs
}

func (p *ProcessInstance) Subscribe() (chan string, func()) {
	p.mu.Lock()
	defer p.mu.Unlock()

	ch := make(chan string, 512)
	p.subscribers[ch] = struct{}{}

	unsubscribe := func() {
		p.mu.Lock()
		defer p.mu.Unlock()
		if _, ok := p.subscribers[ch]; ok {
			delete(p.subscribers, ch)
			close(ch)
		}
	}

	return ch, unsubscribe
}

func calculateDirSizeMB(path string) int64 {
	var totalSize int64
	_ = filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err == nil && info != nil && !info.IsDir() {
			totalSize += info.Size()
		}
		return nil
	})
	return totalSize / (1024 * 1024)
}

func syncServerPort(workingDir, serverType string, port int) {
	if port <= 0 {
		return
	}
	st := strings.ToLower(serverType)
	switch st {
	case "velocity":
		syncVelocityPort(workingDir, port)
	case "bungeecord", "waterfall":
		syncBungeePort(workingDir, port)
	default:
		syncPropertiesPort(workingDir, port)
	}
}

func syncPropertiesPort(workingDir string, port int) {
	propsPath := filepath.Join(workingDir, "server.properties")
	content, err := os.ReadFile(propsPath)
	if err != nil {
		initialProps := fmt.Sprintf("# Minecraft server properties generated by NeoPanel\nserver-port=%d\nquery.port=%d\nmotd=A NeoPanel Minecraft Server\nenable-rcon=false\n", port, port)
		_ = os.WriteFile(propsPath, []byte(initialProps), 0644)
		return
	}

	lines := strings.Split(string(content), "\n")
	hasPort := false
	hasQuery := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "server-port=") {
			lines[i] = fmt.Sprintf("server-port=%d", port)
			hasPort = true
		} else if strings.HasPrefix(trimmed, "query.port=") {
			lines[i] = fmt.Sprintf("query.port=%d", port)
			hasQuery = true
		}
	}
	if !hasPort {
		lines = append(lines, fmt.Sprintf("server-port=%d", port))
	}
	if !hasQuery {
		lines = append(lines, fmt.Sprintf("query.port=%d", port))
	}
	_ = os.WriteFile(propsPath, []byte(strings.Join(lines, "\n")), 0644)
}

func syncVelocityPort(workingDir string, port int) {
	cfgPath := filepath.Join(workingDir, "velocity.toml")
	content, err := os.ReadFile(cfgPath)
	if err != nil {
		return
	}
	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "bind =") {
			lines[i] = fmt.Sprintf("bind = \"0.0.0.0:%d\"", port)
			break
		}
	}
	_ = os.WriteFile(cfgPath, []byte(strings.Join(lines, "\n")), 0644)
}

func syncBungeePort(workingDir string, port int) {
	cfgPath := filepath.Join(workingDir, "config.yml")
	content, err := os.ReadFile(cfgPath)
	if err != nil {
		return
	}
	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "host:") && strings.Contains(trimmed, "25577") {
			lines[i] = fmt.Sprintf("  host: 0.0.0.0:%d", port)
			break
		}
	}
	_ = os.WriteFile(cfgPath, []byte(strings.Join(lines, "\n")), 0644)
}

// buildCommand constructs the execution command for the server process.
// If bubblewrap (`bwrap`) is installed, it runs within an unprivileged Linux
// sandbox jailing the filesystem so malicious plugins cannot read host files
// or traverse into other servers. If `systemd-run` is available, it applies slice
// resource limits (e.g. TasksMax=512) to thwart fork bombs.
func (p *ProcessInstance) buildCommand(javaBinary string, args []string) (*exec.Cmd, string, error) {
	p.mu.RLock()
	isoEngine := p.isolationEngine
	tasksMax := p.tasksMax
	srv := *p.server
	p.mu.RUnlock()
	if isoEngine == "" {
		isoEngine = "auto"
	}
	if tasksMax <= 0 {
		tasksMax = 512
	}

	if isoEngine == "none" || isoEngine == "process" {
		cmd := exec.Command(javaBinary, args...)
		cmd.Dir = p.workingDir
		return cmd, "Unrestricted process (no filesystem isolation)", nil
	}

	bwrapPath, bwrapErr := exec.LookPath("bwrap")
	systemdPath, sysErr := exec.LookPath("systemd-run")
	if isoEngine == "bwrap" && bwrapErr != nil {
		return nil, "", fmt.Errorf("Bubblewrap was requested but bwrap is not installed; install bubblewrap on this node")
	}
	if isoEngine == "systemd" && sysErr != nil {
		return nil, "", fmt.Errorf("systemd isolation was requested but systemd-run is unavailable")
	}
	if isoEngine != "auto" && isoEngine != "bwrap" && isoEngine != "systemd" {
		return nil, "", fmt.Errorf("unknown isolation engine %q", isoEngine)
	}
	nonce := make([]byte, 16)
	if _, err := rand.Read(nonce); err != nil {
		return nil, "", err
	}
	unitArg := fmt.Sprintf("--unit=neopanel-%x.scope", nonce)

	// If user specifically requested systemd-only or bwrap isn't installed
	if (isoEngine == "systemd" || bwrapErr != nil || bwrapPath == "") && sysErr == nil && systemdPath != "" {
		sysArgs := []string{
			"--scope",
			unitArg,
			"--property=KillMode=control-group",
			"-q",
			fmt.Sprintf("--property=TasksMax=%d", tasksMax),
		}
		if downloader.IsNativeSoftware(srv.Type) && srv.MemoryMaxMB > 0 {
			sysArgs = append(sysArgs, fmt.Sprintf("--property=MemoryMax=%dM", srv.MemoryMaxMB))
		}
		if srv.CPULimitPercent > 0 {
			sysArgs = append(sysArgs, fmt.Sprintf("--property=CPUQuota=%d%%", srv.CPULimitPercent))
		}
		sysArgs = append(sysArgs, javaBinary)
		sysArgs = append(sysArgs, args...)

		cmd := exec.Command(systemdPath, sysArgs...)
		cmd.Dir = p.workingDir
		return cmd, fmt.Sprintf("Systemd resource limits only (%s; NO filesystem isolation)", systemdLimitsLabel(srv, tasksMax)), nil
	}

	if bwrapErr == nil && bwrapPath != "" && isoEngine != "systemd" {
		bwrapArgs := []string{
			"--unshare-all",
			"--cap-drop", "ALL",
			"--clearenv",
			"--setenv", "PATH", "/usr/bin:/bin",
			"--setenv", "LANG", "C.UTF-8",
			"--setenv", "HOME", filepath.Clean(p.workingDir),
			"--share-net",
			"--die-with-parent",
			"--ro-bind", "/usr", "/usr",
			"--ro-bind", "/lib", "/lib",
			"--ro-bind", "/bin", "/bin",
			"--dev", "/dev",
			"--proc", "/proc",
			"--tmpfs", "/tmp",
		}

		// Allowlist public runtime configuration; never mount all host /etc or /opt.
		for _, optDir := range []string{
			"/lib64", "/sbin", "/usr/lib/jvm",
			"/etc/hosts", "/etc/hostname", "/etc/nsswitch.conf", "/etc/passwd", "/etc/group",
			"/etc/ssl/certs", "/etc/ssl/openssl.cnf", "/etc/pki/tls/certs", "/etc/pki/ca-trust", "/etc/pki/java", "/etc/crypto-policies", "/etc/localtime", "/etc/timezone",
			"/var/lib/ca-certificates", "/usr/share/ca-certificates",
			"/run/systemd/resolve", "/run/resolvconf", "/run/NetworkManager",
		} {
			if _, err := os.Stat(optDir); err == nil {
				bwrapArgs = append(bwrapArgs, "--ro-bind-try", optDir, optDir)
			}
		}

		// Bind the resolved DNS file directly at /etc/resolv.conf. Binding its
		// parent would expose all /etc on hosts using a regular resolver file.
		bwrapArgs = append(bwrapArgs, sandboxResolverArgs("/etc/resolv.conf")...)

		// Security mask: hide sensitive host credentials from the jail
		for _, maskPath := range []string{"/etc/shadow", "/etc/gshadow", "/etc/sudoers"} {
			if _, err := os.Stat(maskPath); err == nil {
				bwrapArgs = append(bwrapArgs, "--ro-bind-try", "/dev/null", maskPath)
			}
		}
		if _, err := os.Stat("/etc/sudoers.d"); err == nil {
			bwrapArgs = append(bwrapArgs, "--tmpfs", "/etc/sudoers.d")
		}

		// Glob all /etc/java* config directories (critical for OpenJDK java.security symlinks)
		if javaEtcMatches, err := filepath.Glob("/etc/java*"); err == nil {
			for _, m := range javaEtcMatches {
				bwrapArgs = append(bwrapArgs, "--ro-bind-try", m, m)
			}
		}

		if !downloader.IsNativeSoftware(srv.Type) {
			// Resolve java binary symlinks and ensure conf / security targets are explicitly bound
			if realJava, err := filepath.EvalSymlinks(javaBinary); err == nil {
				javaBinDir := filepath.Dir(realJava)
				bwrapArgs = append(bwrapArgs, "--ro-bind-try", javaBinDir, javaBinDir)
				javaHome := filepath.Dir(javaBinDir)
				bwrapArgs = append(bwrapArgs, "--ro-bind-try", javaHome, javaHome)

				// Modern OpenJDK: conf directory symlinks to /etc/java-<version>-openjdk
				confDir := filepath.Join(javaHome, "conf")
				if realConf, err := filepath.EvalSymlinks(confDir); err == nil {
					bwrapArgs = append(bwrapArgs, "--ro-bind-try", realConf, realConf)
				}

				// Legacy OpenJDK 8: jre/lib/security symlinks to /etc/java-8-openjdk/security
				for _, secPath := range []string{
					filepath.Join(javaHome, "lib", "security"),
					filepath.Join(javaHome, "jre", "lib", "security"),
				} {
					if realSec, err := filepath.EvalSymlinks(secPath); err == nil {
						bwrapArgs = append(bwrapArgs, "--ro-bind-try", realSec, realSec)
					}
				}
			}
		}

		cleanWorkDir := filepath.Clean(p.workingDir)
		bwrapArgs = append(bwrapArgs,
			"--bind", cleanWorkDir, cleanWorkDir,
			"--chdir", cleanWorkDir,
		)

		javaDir := filepath.Dir(javaBinary)
		if !downloader.IsNativeSoftware(srv.Type) && !strings.HasPrefix(javaDir, "/usr") && !strings.HasPrefix(javaDir, "/bin") && !strings.HasPrefix(javaDir, "/lib") {
			if _, err := os.Stat(javaDir); err == nil {
				bwrapArgs = append(bwrapArgs, "--ro-bind-try", javaDir, javaDir)
			}
		}

		bwrapArgs = append(bwrapArgs, javaBinary)
		bwrapArgs = append(bwrapArgs, args...)

		if sysErr == nil && systemdPath != "" {
			sysArgs := []string{
				"--scope",
				unitArg,
				"--property=KillMode=control-group",
				"-q",
				fmt.Sprintf("--property=TasksMax=%d", tasksMax),
			}
			if downloader.IsNativeSoftware(srv.Type) && srv.MemoryMaxMB > 0 {
				sysArgs = append(sysArgs, fmt.Sprintf("--property=MemoryMax=%dM", srv.MemoryMaxMB))
			}
			if srv.CPULimitPercent > 0 {
				sysArgs = append(sysArgs, fmt.Sprintf("--property=CPUQuota=%d%%", srv.CPULimitPercent))
			}
			sysArgs = append(sysArgs, bwrapPath)
			sysArgs = append(sysArgs, bwrapArgs...)

			cmd := exec.Command(systemdPath, sysArgs...)
			cmd.Dir = p.workingDir
			return cmd, fmt.Sprintf("Bubblewrap filesystem sandbox + Systemd resource limits (%s)", systemdLimitsLabel(srv, tasksMax)), nil
		}

		cmd := exec.Command(bwrapPath, bwrapArgs...)
		cmd.Dir = p.workingDir
		return cmd, "Bubblewrap filesystem sandbox (no systemd resource limits)", nil
	}

	cmd := exec.Command(javaBinary, args...)
	cmd.Dir = p.workingDir
	return cmd, "Unrestricted process: bwrap/systemd-run unavailable (NO filesystem isolation)", nil
}

// GetIdleDuration returns how long the server has been online with 0 players connected.
func (p *ProcessInstance) GetIdleDuration() (time.Duration, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.status != "online" || len(p.onlinePlayers) > 0 || p.zeroPlayersSince.IsZero() {
		return 0, false
	}
	return time.Since(p.zeroPlayersSince), true
}

// GetPlayerCount returns the count of currently connected players.
func (p *ProcessInstance) GetPlayerCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.onlinePlayers)
}
