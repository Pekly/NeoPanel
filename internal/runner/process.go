package runner

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Pekly/NeoPanel/internal/database"
)

const maxLogBuffer = 1000

type ProcessInstance struct {
	mu           sync.RWMutex
	server       *database.Server
	workingDir   string
	cmd          *exec.Cmd
	stdin        io.WriteCloser
	status       string
	startedAt    time.Time
	metrics      *ProcessMetrics
	logBuffer    []string
	subscribers  map[chan string]struct{}
	stopExpected bool
	onExit       func(srvID string, unexpected bool)
}

func NewProcessInstance(server *database.Server, workingDir string, onExit func(string, bool)) *ProcessInstance {
	return &ProcessInstance{
		server:      server,
		workingDir:  workingDir,
		status:      "offline",
		logBuffer:   make([]string, 0, maxLogBuffer),
		subscribers: make(map[chan string]struct{}),
		onExit:      onExit,
	}
}

func (p *ProcessInstance) GetStatus() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.status
}

func (p *ProcessInstance) GetStats() database.ServerStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	stats := database.ServerStats{
		ServerID: p.server.ID,
		Status:   p.status,
	}

	if p.status == "online" || p.status == "starting" {
		stats.UptimeSec = int64(time.Since(p.startedAt).Seconds())
		if p.metrics != nil {
			cpu, mem, err := p.metrics.ReadStats()
			if err == nil {
				stats.CPUUsage = cpu
				stats.MemoryMB = mem
			}
		}
	}

	// Calculate disk usage of workingDir
	stats.DiskMB = calculateDirSizeMB(p.workingDir)

	return stats
}

func (p *ProcessInstance) Start() error {
	p.mu.Lock()
	if p.status == "online" || p.status == "starting" {
		p.mu.Unlock()
		return fmt.Errorf("server is already running")
	}

	p.status = "starting"
	p.stopExpected = false
	p.mu.Unlock()

	// Build Java command
	javaBinary, javaLabel := ResolveJavaBinary(p.server.JavaPath)
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
		p.broadcast("\x1b[36m👉 Solution: Open Server Settings -> Java Version and choose an installed runtime (e.g. Java 21, Java 17, Java 25), or run 'sudo neopanel-installer --java' on your VPS.\x1b[0m")
		p.broadcast("\x1b[31;1m========================================================================\x1b[0m")
		return fmt.Errorf("Java binary '%s' (%s) was not found. Please verify your Java settings in Server Settings", javaBinary, javaLabel)
	}

	jarPath := filepath.Join(p.workingDir, p.server.JarFile)
	if fi, err := os.Stat(jarPath); os.IsNotExist(err) || (err == nil && fi.IsDir()) {
		p.mu.Lock()
		p.status = "offline"
		p.mu.Unlock()

		p.broadcast("\x1b[31;1m========================================================================\x1b[0m")
		p.broadcast("\x1b[31;1m[NeoPanel ERROR] SERVER JAR FILE NOT FOUND!\x1b[0m")
		p.broadcast("\x1b[31mCould not locate the executable JAR file in server directory:\x1b[0m")
		p.broadcast(fmt.Sprintf("  \x1b[33;1mExpected file: %s\x1b[0m", p.server.JarFile))
		p.broadcast(fmt.Sprintf("  \x1b[33;1mFull path:     %s\x1b[0m", jarPath))
		p.broadcast("\x1b[36m👉 Solution: Click 'Reinstall Server' in Settings to cleanly download an official JAR from mcjarfiles.com, or upload your server JAR to the root folder via the File Manager.\x1b[0m")
		p.broadcast("\x1b[31;1m========================================================================\x1b[0m")
		return fmt.Errorf("server JAR file '%s' not found. Please upload it or use 'Reinstall Server'", p.server.JarFile)
	}

	// Validate memory settings
	if p.server.MemoryMinMB > 0 && p.server.MemoryMaxMB > 0 && p.server.MemoryMinMB > p.server.MemoryMaxMB {
		p.mu.Lock()
		p.status = "offline"
		p.mu.Unlock()

		errMsg := fmt.Sprintf("Minimum RAM (%d MB) cannot exceed Maximum RAM (%d MB)", p.server.MemoryMinMB, p.server.MemoryMaxMB)
		p.broadcast(fmt.Sprintf("\x1b[31;1m[NeoPanel ERROR] INVALID RAM SETTINGS: %s\x1b[0m", errMsg))
		return fmt.Errorf("%s", errMsg)
	}

	args := []string{}
	if p.server.MemoryMinMB > 0 {
		args = append(args, fmt.Sprintf("-Xms%dM", p.server.MemoryMinMB))
	}
	if p.server.MemoryMaxMB > 0 {
		args = append(args, fmt.Sprintf("-Xmx%dM", p.server.MemoryMaxMB))
	}

	// Force full ANSI terminal color output even when piped in daemons/subprocesses
	args = append(args,
		"-Dterminal.ansi=true",
		"-Dterminal.jline=false",
		"-Dnet.minecrell.terminalconsole.ansi=true",
		"-Dlog4j.skipJansi=false",
	)

	// Add custom JVM flags if provided
	if strings.TrimSpace(p.server.JavaArgs) != "" {
		customArgs := strings.Fields(p.server.JavaArgs)
		args = append(args, customArgs...)
	}

	// Ensure server game port is synchronized to configuration file
	port := p.server.Port
	if port <= 0 {
		port = 25565
	}
	syncServerPort(p.workingDir, p.server.Type, port)
	p.broadcast(fmt.Sprintf("\x1b[36m[NeoPanel] Bound game server port: %d (TCP/UDP)\x1b[0m", port))

	args = append(args, "-jar", p.server.JarFile, "nogui")

	// Pass CLI port parameter for game servers that support port override
	lowerType := strings.ToLower(p.server.Type)
	if lowerType != "velocity" && lowerType != "bungeecord" && lowerType != "waterfall" {
		args = append(args, "--port", fmt.Sprintf("%d", port))
	}

	cmd := exec.Command(javaBinary, args...)
	cmd.Dir = p.workingDir

	stdin, err := cmd.StdinPipe()
	if err != nil {
		p.mu.Lock()
		p.status = "offline"
		p.mu.Unlock()
		p.broadcast(fmt.Sprintf("\x1b[31;1m[NeoPanel ERROR] Failed to open stdin pipe: %v\x1b[0m", err))
		return fmt.Errorf("failed to open stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		stdin.Close()
		p.mu.Lock()
		p.status = "offline"
		p.mu.Unlock()
		p.broadcast(fmt.Sprintf("\x1b[31;1m[NeoPanel ERROR] Failed to open stdout pipe: %v\x1b[0m", err))
		return fmt.Errorf("failed to open stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		stdin.Close()
		stdout.Close()
		p.mu.Lock()
		p.status = "offline"
		p.mu.Unlock()
		p.broadcast(fmt.Sprintf("\x1b[31;1m[NeoPanel ERROR] Failed to open stderr pipe: %v\x1b[0m", err))
		return fmt.Errorf("failed to open stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		stdin.Close()
		stdout.Close()
		stderr.Close()
		p.mu.Lock()
		p.status = "offline"
		p.mu.Unlock()

		p.broadcast("\x1b[31;1m========================================================================\x1b[0m")
		p.broadcast("\x1b[31;1m[NeoPanel ERROR] PROCESS LAUNCH FAILED!\x1b[0m")
		p.broadcast(fmt.Sprintf("\x1b[31mFailed to start Java process: %v\x1b[0m", err))
		p.broadcast(fmt.Sprintf("  \x1b[33;1mCommand: %s %s\x1b[0m", javaBinary, strings.Join(args, " ")))
		p.broadcast("\x1b[31;1m========================================================================\x1b[0m")
		return fmt.Errorf("failed to start Java process: %w", err)
	}

	p.mu.Lock()
	p.cmd = cmd
	p.stdin = stdin
	p.startedAt = time.Now()
	p.metrics = NewProcessMetrics(cmd.Process.Pid)
	p.mu.Unlock()

	p.broadcast(fmt.Sprintf("\x1b[32m[NeoPanel] Server process started (PID: %d)...\x1b[0m", cmd.Process.Pid))

	// Stream stdout & stderr concurrently
	go p.streamOutput(stdout)
	go p.streamOutput(stderr)

	// Wait for process exit in background
	go func() {
		waitErr := cmd.Wait()

		p.mu.Lock()
		p.status = "offline"
		p.cmd = nil
		p.stdin = nil
		expected := p.stopExpected
		p.mu.Unlock()

		if waitErr != nil && !expected {
			p.broadcast(fmt.Sprintf("\x1b[31;1m[NeoPanel] Server process terminated unexpectedly (%v).\x1b[0m", waitErr))
		} else {
			p.broadcast("\x1b[33m[NeoPanel] Server process stopped.\x1b[0m")
		}

		if p.onExit != nil {
			p.onExit(p.server.ID, !expected)
		}
	}()

	return nil
}

func (p *ProcessInstance) streamOutput(r io.Reader) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()

		// Detect when Minecraft finishes startup
		p.mu.Lock()
		if p.status == "starting" && (strings.Contains(line, "Done (") || strings.Contains(line, "For help, type \"help\"")) {
			p.status = "online"
		}
		p.mu.Unlock()

		p.broadcast(line)

		// Real-time runtime diagnostic detectors
		p.detectRuntimeDiagnostics(line)
	}
}

func (p *ProcessInstance) detectRuntimeDiagnostics(line string) {
	lower := strings.ToLower(line)

	// 1. Java Version Mismatch / UnsupportedClassVersionError
	if strings.Contains(line, "UnsupportedClassVersionError") ||
		strings.Contains(lower, "has been compiled by a more recent version of the java runtime") ||
		strings.Contains(lower, "requires running the server with java") ||
		strings.Contains(lower, "unsupported major.minor version") ||
		(strings.Contains(lower, "requires java") && strings.Contains(lower, "or higher")) {
		p.broadcast("\x1b[31;1m========================================================================\x1b[0m")
		p.broadcast("\x1b[31;1m[NeoPanel DIAGNOSTIC] JAVA VERSION MISMATCH DETECTED!\x1b[0m")
		p.broadcast("\x1b[31mYour Minecraft server JAR was compiled for a newer Java runtime than the one currently selected.\x1b[0m")
		p.broadcast("\x1b[33;1m👉 Quick Fix: Go to Server Settings -> Java Version, change to Java 21 or Java 25, then restart.\x1b[0m")
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
		p.broadcast("\x1b[33;1m👉 Quick Fix: Lower your RAM allocation in Settings -> Startup, or create a swap file on the host.\x1b[0m")
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

func (p *ProcessInstance) Stop(graceful bool) error {
	p.mu.Lock()
	if p.cmd == nil || p.status == "offline" {
		p.mu.Unlock()
		return nil
	}
	p.stopExpected = true
	p.status = "stopping"
	cmd := p.cmd
	stdin := p.stdin
	p.mu.Unlock()

	if graceful && stdin != nil {
		p.broadcast("\x1b[36m[NeoPanel] Sending 'stop' command to server...\x1b[0m")
		_, _ = io.WriteString(stdin, "stop\n")

		// Wait up to 15 seconds for clean exit
		done := make(chan struct{})
		go func() {
			for i := 0; i < 30; i++ {
				p.mu.RLock()
				st := p.status
				p.mu.RUnlock()
				if st == "offline" {
					close(done)
					return
				}
				time.Sleep(500 * time.Millisecond)
			}
			close(done)
		}()

		select {
		case <-done:
			if p.GetStatus() == "offline" {
				return nil
			}
		case <-time.After(15 * time.Second):
		}
	}

	// Force kill if still alive
	p.broadcast("\x1b[31m[NeoPanel] Force killing process...\x1b[0m")
	if cmd != nil && cmd.Process != nil {
		_ = cmd.Process.Kill()
	}
	return nil
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

	ch := make(chan string, 100)
	p.subscribers[ch] = struct{}{}

	unsubscribe := func() {
		p.mu.Lock()
		defer p.mu.Unlock()
		delete(p.subscribers, ch)
		close(ch)
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

