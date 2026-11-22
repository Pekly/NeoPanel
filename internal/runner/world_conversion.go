package runner

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Pekly/NeoPanel/internal/chunker"
	"github.com/Pekly/NeoPanel/internal/database"
	"github.com/Pekly/NeoPanel/internal/rooted"
)

const conversionTimeout = 30 * time.Minute
const conversionMaxJobs = 32
const conversionMaxActive = 2

const WorldConversionWarning = "Conversion is lossy: entities, structures, custom dimensions and plugin/mod data may be lost. Output is separate; nothing is automatically replaced or started. Uploaded .chunker.json settings are rejected."

type WorldConversionRequest struct {
	Input           string `json:"input"`
	Format          string `json:"format"`
	AcknowledgeLoss bool   `json:"acknowledge_loss"`
}
type WorldConversionReplaceRequest struct {
	Token           string `json:"token"`
	Confirm         string `json:"confirm"`
	Backup          bool   `json:"backup"`
	AcknowledgeLoss bool   `json:"acknowledge_loss"`
}
type WorldConversionJob struct {
	ID           string    `json:"id"`
	ServerID     string    `json:"server_id"`
	Input        string    `json:"input"`
	Format       string    `json:"format"`
	State        string    `json:"state"`
	CreatedAt    time.Time `json:"created_at"`
	FinishedAt   time.Time `json:"finished_at,omitempty"`
	Output       string    `json:"output,omitempty"`
	ReplaceToken string    `json:"replace_token,omitempty"`
	CanReplace   bool      `json:"can_replace"`
	Backup       string    `json:"backup,omitempty"`
	SafetyCopy   string    `json:"safety_copy,omitempty"`
	Error        string    `json:"error,omitempty"`
	Log          string    `json:"log"`
	Progress     float64   `json:"progress"`
}
type conversionJob struct {
	WorldConversionJob
	cancel                 context.CancelFunc
	log                    *chunker.Log
	sourceHash, outputHash string
	lease                  *conversionLease // retained when process-tree cleanup cannot be certified
}

// WorldConverter owns a finite registry. Create one per daemon, not per request.
// Auxiliary work joins the existing importWG shutdown fence; no manager changes
// or game ProcessInstance/auto-restart hooks are needed.
type WorldConverter struct {
	mu      sync.Mutex
	manager *ServerManager
	toolDir string
	enabled func(string) bool
	jobs    map[string]*conversionJob
	active  int
}

func NewWorldConverter(m *ServerManager, toolDir string, enabled func(string) bool) *WorldConverter {
	return &WorldConverter{manager: m, toolDir: toolDir, enabled: enabled, jobs: map[string]*conversionJob{}}
}
func conversionID() string {
	var b [16]byte
	if _, e := rand.Read(b[:]); e != nil {
		panic(e)
	}
	return hex.EncodeToString(b[:])
}
func (c *WorldConverter) allowed(id string) bool { return c.enabled != nil && c.enabled(id) }
func (c *WorldConverter) List(serverID string) []WorldConversionJob {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := []WorldConversionJob{}
	for _, j := range c.jobs {
		if j.ServerID == serverID {
			v := j.WorldConversionJob
			v.Log, v.Progress, _ = j.log.Snapshot()
			out = append(out, v)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out
}
func (c *WorldConverter) Cancel(serverID, id string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	j := c.jobs[id]
	if j == nil || j.ServerID != serverID {
		return os.ErrNotExist
	}
	if j.cancel != nil {
		j.cancel()
	}
	return nil
}
func (c *WorldConverter) Start(srv *database.Server, req WorldConversionRequest) (WorldConversionJob, error) {
	if srv == nil || !c.allowed(srv.ID) {
		return WorldConversionJob{}, errors.New("world conversion module is disabled")
	}
	if !req.AcknowledgeLoss {
		return WorldConversionJob{}, errors.New("acknowledge possible data loss before conversion")
	}
	if _, ok := chunker.Lookup(req.Format); !ok {
		return WorldConversionJob{}, errors.New("unsupported conversion format")
	}
	if !rooted.Basename(req.Input) || strings.HasPrefix(req.Input, ".") || strings.HasPrefix(req.Input, "converted-") {
		return WorldConversionJob{}, errors.New("input must be a direct world directory")
	}
	if err := conversionPlatform(srv); err != nil {
		return WorldConversionJob{}, err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.active >= conversionMaxActive {
		return WorldConversionJob{}, errors.New("conversion capacity reached; retry later")
	}
	if len(c.jobs) >= conversionMaxJobs {
		var oldest *conversionJob
		for _, j := range c.jobs {
			if j.cancel == nil && j.lease == nil && (oldest == nil || j.CreatedAt.Before(oldest.CreatedAt)) {
				oldest = j
			}
		}
		if oldest == nil {
			return WorldConversionJob{}, errors.New("conversion job limit reached")
		}
		delete(c.jobs, oldest.ID)
	}
	lease, err := c.manager.beginConversion(srv)
	if err != nil {
		return WorldConversionJob{}, err
	}
	j := &conversionJob{WorldConversionJob: WorldConversionJob{ID: conversionID(), ServerID: srv.ID, Input: req.Input, Format: req.Format, State: "preparing", CreatedAt: time.Now().UTC()}, cancel: lease.cancel, log: &chunker.Log{}}
	c.jobs[j.ID] = j
	c.active++
	snapshot := j.WorldConversionJob
	server := *srv
	go c.work(j, lease, func() error { return c.convert(&server, j, lease) })
	return snapshot, nil
}
func (c *WorldConverter) work(j *conversionJob, l *conversionLease, fn func() error) {
	err := fn()
	c.mu.Lock()
	if err != nil {
		j.State = "failed"
		j.Error = err.Error()
		if l.ctx.Err() != nil {
			j.State = "cancelled"
		}
	} else if j.State == "replacing" {
		j.State = "replaced"
	} else {
		j.State = "succeeded"
	}
	j.FinishedAt = time.Now().UTC()
	j.cancel = nil
	if l.unsafe {
		j.lease = l
		j.State = "blocked"
		j.Error = "Process cleanup could not be certified; administrator recovery required"
	}
	c.active--
	c.mu.Unlock()
	l.close()
}
func (c *WorldConverter) phase(j *conversionJob, state string) {
	c.mu.Lock()
	j.State = state
	c.mu.Unlock()
}

type conversionLease struct {
	manager   *ServerManager
	ctx       context.Context
	cancel    context.CancelFunc
	watchDone chan struct{}
	unlock    func()
	lock      *os.File
	live      string
	owned     *ownedCommand
	info      os.FileInfo
	unsafe    bool
}

func (m *ServerManager) beginConversion(srv *database.Server) (*conversionLease, error) {
	if srv == nil || !safeBackupComponent(srv.ID) || strings.HasPrefix(srv.ID, ".") || (srv.NodeID != "" && srv.NodeID != "local") {
		return nil, errors.New("conversion requires a local server")
	}
	unlock, err := m.beginOperation(srv.ID)
	if err != nil {
		return nil, err
	}
	// Keep managed writes excluded from initial clone through publication or
	// replacement/rollback. Try-only acquisition cannot deadlock manager gates.
	releaseFiles, err := rooted.TryExclusiveMutation(m.GetWorkingDir(srv.ID))
	if err != nil {
		unlock()
		return nil, fmt.Errorf("%w: %v", ErrServerOperation, err)
	}
	releaseOperation := unlock
	unlock = func() { releaseFiles(); releaseOperation() }
	m.launchMu.Lock()
	if m.isShuttingDown() {
		m.launchMu.Unlock()
		unlock()
		return nil, errDaemonShuttingDown
	}
	if m.GetStatus(srv.ID) != "offline" {
		m.launchMu.Unlock()
		unlock()
		return nil, errors.New("stop the server before world conversion; conversion never stops it for you")
	}
	err = m.invalidateIdleWake(srv.ID)
	if err != nil {
		m.launchMu.Unlock()
		unlock()
		return nil, err
	}
	m.mu.Lock()
	m.cancelRetryLocked(srv.ID)
	m.restorePaused.Store(srv.ID, true)
	m.mu.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), conversionTimeout)
	m.importWG.Add(1)
	l := &conversionLease{manager: m, ctx: ctx, cancel: cancel, watchDone: make(chan struct{}), unlock: unlock, live: m.GetWorkingDir(srv.ID)}
	go func() {
		defer close(l.watchDone)
		select {
		case <-m.shutdownCh:
			cancel()
		case <-ctx.Done():
		}
	}()
	m.launchMu.Unlock()
	l.info, err = os.Lstat(l.live)
	if err == nil && (!l.info.IsDir() || l.info.Mode()&os.ModeSymlink != 0) {
		err = errors.New("server directory must not be a link")
	}
	if err == nil {
		l.lock, err = acquireRunLock(l.live)
	}
	if err == nil {
		err = ensureServerProcessesStopped(l.live)
	}
	if err == nil {
		err = checkServerWorldLocks(l.live, srv.Type)
	}
	if err != nil {
		l.close()
		return nil, err
	}
	return l, nil
}
func (l *conversionLease) close() {
	l.cancel()
	<-l.watchDone
	if !l.unsafe {
		if l.lock != nil {
			l.lock.Close()
		}
		l.unlock()
	}
	l.manager.importWG.Done()
}
func (l *conversionLease) fence(fn func() error) error {
	l.manager.launchMu.Lock()
	defer l.manager.launchMu.Unlock()
	if l.manager.isShuttingDown() {
		return errDaemonShuttingDown
	}
	if err := l.ctx.Err(); err != nil {
		return err
	}
	st, err := os.Lstat(l.live)
	if err != nil || !os.SameFile(l.info, st) || st.Mode()&os.ModeSymlink != 0 {
		return errors.New("server directory changed")
	}
	return fn()
}

func (c *WorldConverter) convert(srv *database.Server, j *conversionJob, l *conversionLease) error {
	if !c.allowed(srv.ID) {
		return errors.New("world conversion module is disabled")
	}
	// Storage must be outside all server trees, including custom configurations.
	tools, err := filepath.Abs(c.toolDir)
	if err != nil {
		return err
	}
	base, err := filepath.Abs(c.manager.baseDir)
	if err != nil {
		return err
	}
	if rel, e := filepath.Rel(base, tools); e == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return errors.New("tool storage must be outside server storage")
	}
	if err := chunker.PrivateDir(tools); err != nil {
		return err
	}
	src, err := chunker.OpenWorld(l.live, j.Input)
	if err != nil {
		return err
	}
	defer src.Close()
	level, err := src.Open("level.dat")
	if err != nil {
		return errors.New("input is not an unpacked world with regular level.dat")
	}
	level.Close()
	if err = conversionRetention(l.live); err != nil {
		return err
	}
	limit := chunker.MaxBytes
	if srv.DiskLimitMB > 0 && srv.DiskLimitMB < (limit>>20)*4 {
		limit = srv.DiskLimitMB * (1 << 20) / 4
	}
	if limit < 1<<20 {
		return errors.New("insufficient assigned disk budget")
	}
	preHash, preSize, err := chunker.Tree(l.ctx, src, nil, limit)
	if err != nil {
		return err
	}
	if err = conversionDiskBudget(l.ctx, srv, l.live, base, limit, preSize); err != nil {
		return err
	}
	if err = conversionStagingRetention(base, srv.ID); err != nil {
		return err
	}
	stage, err := os.MkdirTemp(base, ".chunker-"+srv.ID+"-")
	if err != nil {
		return err
	}
	defer func() {
		if !l.unsafe {
			os.RemoveAll(stage)
		}
	}()
	input := filepath.Join(stage, "input")
	output := filepath.Join(stage, "output")
	if err = os.Mkdir(input, 0700); err != nil {
		return err
	}
	if err = os.Mkdir(output, 0700); err != nil {
		return err
	}
	dst, err := rooted.Open(input)
	if err != nil {
		return err
	}
	sourceHash, _, err := chunker.Tree(l.ctx, src, dst, limit)
	dst.Close()
	if err != nil {
		return err
	}
	if sourceHash != preHash {
		return errors.New("input changed during cloning; retry conversion")
	}
	jar, err := chunker.EnsureTool(l.ctx, tools)
	if err != nil {
		return err
	}
	java, err := conversionJava(l.ctx, tools)
	if err != nil {
		return err
	}
	if !c.allowed(srv.ID) {
		return errors.New("world conversion module is disabled")
	}
	c.phase(j, "converting")
	cmd, err := conversionCommand(srv, java, jar, stage, j.Format, limit)
	if err != nil {
		return err
	}
	reader, writer, err := os.Pipe()
	if err != nil {
		return err
	}
	defer reader.Close()
	cmd.Stdout = writer
	cmd.WaitDelay = 5 * time.Second
	cmd.Stderr = j.log
	var owned *ownedCommand
	err = l.fence(func() error { var e error; owned, e = startOwnedCommand(cmd, l.lock, int64(srv.MemoryMaxMB)); return e })
	writer.Close()
	if err != nil {
		if owned != nil {
			l.unsafe = true
			l.owned = owned
		}
		return err
	}
	extractDone := make(chan error, 1)
	go func() {
		e := chunker.ExtractOutput(l.ctx, reader, output, limit)
		if e == nil {
			_, e = io.Copy(io.Discard, io.LimitReader(reader, 64<<20))
		}
		if e != nil {
			_ = owned.stop(true)
		}
		reader.Close()
		extractDone <- e
	}()
	done := make(chan struct{})
	stopped := make(chan struct{})
	go func() {
		defer close(stopped)
		select {
		case <-l.ctx.Done():
			_ = owned.stop(true)
		case <-done:
		}
	}()
	waitErr, cleanupErr := owned.wait()
	close(done)
	<-stopped
	if cleanupErr != nil {
		reader.Close()
	}
	extractErr := <-extractDone
	if cleanupErr != nil {
		l.unsafe = true
		l.owned = owned
		c.manager.restoreBlocked.Store(srv.ID, cleanupErr)
		return cleanupErr
	}
	if err = l.ctx.Err(); err != nil {
		return err
	}
	if waitErr != nil {
		return fmt.Errorf("converter failed: %w", waitErr)
	}
	if extractErr != nil {
		return extractErr
	}
	_, _, complete := j.log.Snapshot()
	if !complete {
		return errors.New("converter did not report successful completion")
	}
	c.phase(j, "validating")
	out, err := rooted.Open(output)
	if err != nil {
		return err
	}
	defer out.Close()
	f, _ := chunker.Lookup(j.Format)
	if err = chunker.ValidateLevel(out, f); err != nil {
		return err
	}
	outputHash, _, err := chunker.Tree(l.ctx, out, nil, limit)
	if err != nil {
		return err
	}
	name := "converted-" + j.ID
	err = l.fence(func() error {
		if !c.allowed(srv.ID) {
			return errors.New("world conversion module is disabled")
		}
		r, e := os.OpenRoot(l.live)
		if e != nil {
			return e
		}
		defer r.Close()
		if _, e = r.Lstat(name); !os.IsNotExist(e) {
			return errors.New("conversion output already exists")
		}
		return os.Rename(output, filepath.Join(l.live, name))
	})
	if err != nil {
		return err
	}
	c.mu.Lock()
	j.Output = name
	j.sourceHash = sourceHash
	j.outputHash = outputHash
	j.CanReplace = conversionCompatible(srv, j.Input, f, l.live) == nil
	if j.CanReplace {
		j.ReplaceToken = conversionID()
	}
	c.mu.Unlock()
	return nil
}
func conversionRetention(live string) error {
	f, e := os.Open(live)
	if e != nil {
		return e
	}
	defer f.Close()
	count := 0
	total := 0
	for {
		names, err := f.Readdirnames(128)
		total += len(names)
		if total > chunker.MaxEntries {
			return errors.New("server directory entry limit exceeded")
		}
		for _, n := range names {
			if strings.HasPrefix(n, "converted-") {
				count++
			}
		}
		if count >= 3 {
			return errors.New("keep at most three converted outputs; remove an old output in Files first")
		}
		if len(names) == 0 {
			return errOrEOF(err)
		}
	}
}
func errOrEOF(err error) error {
	if err == io.EOF {
		return nil
	}
	return err
}

// WorldConversionSupport reports prerequisites only; launch verifies live limits.
func WorldConversionSupport(srv *database.Server) string {
	if srv == nil {
		return "Server not found"
	}
	if err := conversionPlatform(srv); err != nil {
		return err.Error()
	}
	return ""
}

func conversionStagingRetention(base, id string) error {
	d, e := os.Open(base)
	if e != nil {
		return e
	}
	defer d.Close()
	count, seen := 0, 0
	for {
		names, e := d.Readdirnames(128)
		seen += len(names)
		if seen > chunker.MaxEntries {
			return errors.New("staging inspection limit exceeded")
		}
		for _, name := range names {
			if strings.HasPrefix(name, ".chunker-"+id+"-") {
				count++
			}
		}
		if count >= 3 {
			return errors.New("interrupted conversion staging requires administrator cleanup")
		}
		if len(names) == 0 {
			return errOrEOF(e)
		}
	}
}
