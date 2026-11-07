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

	jarPath := filepath.Join(p.workingDir, p.server.JarFile)
	if _, err := os.Stat(jarPath); os.IsNotExist(err) {
		p.mu.Lock()
		p.status = "offline"
		p.mu.Unlock()
		return fmt.Errorf("server jar file not found at %s", jarPath)
	}

	args := []string{}
	if p.server.MemoryMinMB > 0 {
		args = append(args, fmt.Sprintf("-Xms%dM", p.server.MemoryMinMB))
	}
	if p.server.MemoryMaxMB > 0 {
		args = append(args, fmt.Sprintf("-Xmx%dM", p.server.MemoryMaxMB))
	}

	// Add custom JVM flags if provided
	if strings.TrimSpace(p.server.JavaArgs) != "" {
		customArgs := strings.Fields(p.server.JavaArgs)
		args = append(args, customArgs...)
	}

	args = append(args, "-jar", p.server.JarFile, "nogui")

	cmd := exec.Command(javaBinary, args...)
	cmd.Dir = p.workingDir

	stdin, err := cmd.StdinPipe()
	if err != nil {
		p.mu.Lock()
		p.status = "offline"
		p.mu.Unlock()
		return fmt.Errorf("failed to open stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		stdin.Close()
		p.mu.Lock()
		p.status = "offline"
		p.mu.Unlock()
		return fmt.Errorf("failed to open stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		stdin.Close()
		stdout.Close()
		p.mu.Lock()
		p.status = "offline"
		p.mu.Unlock()
		return fmt.Errorf("failed to open stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		stdin.Close()
		stdout.Close()
		stderr.Close()
		p.mu.Lock()
		p.status = "offline"
		p.mu.Unlock()
		return fmt.Errorf("failed to start process: %w", err)
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
		_ = cmd.Wait()

		p.mu.Lock()
		p.status = "offline"
		p.cmd = nil
		p.stdin = nil
		expected := p.stopExpected
		p.mu.Unlock()

		p.broadcast("\x1b[33m[NeoPanel] Server process stopped.\x1b[0m")

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
	}
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

