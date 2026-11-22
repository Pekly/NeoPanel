package runner

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/subtle"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Pekly/NeoPanel/internal/chunker"
	"github.com/Pekly/NeoPanel/internal/database"
	"github.com/Pekly/NeoPanel/internal/rooted"
)

func conversionCompatible(srv *database.Server, input string, f chunker.Format, live string) error {
	// Bukkit splits dimensions; mods/custom server versions do not establish
	// compatibility. These outputs remain available for manual export only.
	if srv.Type != "vanilla" || f.Edition != "java" || srv.Version != f.Version {
		return errors.New("replacement requires vanilla Java with the exact selected Minecraft version; other software and Bedrock targets are export-only")
	}
	root, e := rooted.Open(live)
	if e != nil {
		return e
	}
	defer root.Close()
	name := "world"
	raw, e := root.ReadFile("server.properties")
	if e != nil {
		return e
	}
	if len(raw) > 1<<20 {
		return errors.New("server.properties too large")
	}
	// Accept only an unambiguous subset of Java Properties. In particular, do
	// not silently default to "world" for colon/whitespace separators, escaped
	// keys/values or continuations. Those configurations remain export-only.
	text := strings.ReplaceAll(string(raw), "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimLeft(line, " \t\f")
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "!") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		k = strings.TrimRight(k, " \t\f")
		if !ok || strings.ContainsAny(k, ": \t\f") || strings.ContainsRune(line, '\\') {
			return errors.New("automatic replacement requires simple key=value server.properties without escapes or continuations")
		}
		if k == "level-name" {
			// Java discards leading property whitespace, but preserves trailing spaces.
			name = strings.TrimLeft(v, " \t\f")
		}
	}
	if name != input || !rooted.Basename(name) {
		return errors.New("replacement must target the configured primary world")
	}
	for _, n := range []string{name + "_nether", name + "_the_end"} {
		if _, e = root.Lstat(n); !os.IsNotExist(e) {
			return errors.New("split-dimension worlds cannot be replaced automatically")
		}
	}
	return nil
}
func (c *WorldConverter) Replace(srv *database.Server, id string, req WorldConversionReplaceRequest) (WorldConversionJob, error) {
	if srv == nil || !c.allowed(srv.ID) {
		return WorldConversionJob{}, errors.New("world conversion module is disabled")
	}
	if req.Confirm != "REPLACE" || !req.Backup || !req.AcknowledgeLoss {
		return WorldConversionJob{}, errors.New("replacement requires confirm REPLACE, backup=true and acknowledge_loss=true")
	}
	if err := conversionPlatform(srv); err != nil {
		return WorldConversionJob{}, err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	j := c.jobs[id]
	if j == nil || j.ServerID != srv.ID {
		return WorldConversionJob{}, os.ErrNotExist
	}
	if j.State != "succeeded" || !j.CanReplace || j.ReplaceToken == "" || subtle.ConstantTimeCompare([]byte(req.Token), []byte(j.ReplaceToken)) != 1 {
		return WorldConversionJob{}, errors.New("replacement requires a successful compatible job and its one-time confirmation token")
	}
	if c.active >= conversionMaxActive {
		return WorldConversionJob{}, errors.New("conversion capacity reached")
	}
	f, _ := chunker.Lookup(j.Format)
	if err := conversionCompatible(srv, j.Input, f, c.manager.GetWorkingDir(srv.ID)); err != nil {
		return WorldConversionJob{}, err
	}
	l, err := c.manager.beginConversion(srv)
	if err != nil {
		return WorldConversionJob{}, err
	}
	j.State = "replacing"
	j.ReplaceToken = ""
	j.CanReplace = false
	j.FinishedAt = time.Time{}
	j.cancel = l.cancel
	c.active++
	view := j.WorldConversionJob
	server := *srv
	go c.work(j, l, func() error { return c.replace(&server, j, l) })
	return view, nil
}
func (c *WorldConverter) replace(srv *database.Server, j *conversionJob, l *conversionLease) error {
	f, _ := chunker.Lookup(j.Format)
	if err := conversionCompatible(srv, j.Input, f, l.live); err != nil {
		return err
	}
	root, err := rooted.Open(l.live)
	if err != nil {
		return err
	}
	defer root.Close()
	original, err := root.Sub(j.Input)
	if err != nil {
		return err
	}
	defer original.Close()
	originalInfo, err := root.Lstat(j.Input)
	if err != nil {
		return err
	}
	output, err := root.Sub(j.Output)
	if err != nil {
		return err
	}
	defer output.Close()
	if err = chunker.ValidateLevel(output, f); err != nil {
		return err
	}
	// Keep no more than three original safety trees, including leftovers from
	// interrupted operations. Never automatically delete recovery data.
	base := filepath.Dir(l.live)
	if err = conversionSafetyRetention(base, srv.ID); err != nil {
		return err
	}
	_, preOutSize, err := chunker.Tree(l.ctx, output, nil, chunker.MaxBytes)
	if err != nil {
		return err
	}
	_, preInSize, err := chunker.Tree(l.ctx, original, nil, chunker.MaxBytes)
	if err != nil {
		return err
	}
	if err = conversionDiskBudget(l.ctx, srv, l.live, base, preOutSize, preInSize); err != nil {
		return err
	}
	stage, err := os.MkdirTemp(base, ".chunker-original-"+srv.ID+"-")
	if err != nil {
		return err
	}
	keep := false
	defer func() {
		if !keep {
			os.RemoveAll(stage)
		}
	}()
	fresh := filepath.Join(stage, "world")
	snapshot := filepath.Join(stage, "snapshot")
	if err = os.Mkdir(fresh, 0700); err != nil {
		return err
	}
	if err = os.Mkdir(snapshot, 0700); err != nil {
		return err
	}
	newRoot, err := rooted.Open(fresh)
	if err != nil {
		return err
	}
	hash, _, err := chunker.Tree(l.ctx, output, newRoot, chunker.MaxBytes)
	newRoot.Close()
	if err != nil {
		return err
	}
	if hash != j.outputHash {
		return errors.New("converted output changed; run a new conversion")
	}
	snapRoot, err := rooted.Open(snapshot)
	if err != nil {
		return err
	}
	hash, _, err = chunker.Tree(l.ctx, original, snapRoot, chunker.MaxBytes)
	snapRoot.Close()
	if err != nil {
		return err
	}
	if hash != j.sourceHash {
		return errors.New("original world changed since conversion; run a new conversion")
	}
	c.manager.mu.RLock()
	// Keep world-only archives OUT of the whole-server RestoreBackup catalog.
	backupDir := filepath.Join(c.manager.backupDir, srv.ID, "world-conversions")
	c.manager.mu.RUnlock()
	backup := "world-before-chunker-" + j.ID + ".tar.gz"
	if err = conversionWorldBackup(l.ctx, snapshot, backupDir, backup); err != nil {
		return err
	}
	c.mu.Lock()
	j.Backup = "world-conversions/" + backup
	c.mu.Unlock()
	// The conversion lease excludes managed writes through exchange. Fingerprints
	// additionally detect intervening edits between jobs; external host writers
	// remain unsupported and cannot be made atomic by this process-local lease.
	// Re-read original after backup. No replacement of a different inode/tree.
	current, err := root.Sub(j.Input)
	if err != nil {
		return err
	}
	hash, _, err = chunker.Tree(l.ctx, current, nil, chunker.MaxBytes)
	current.Close()
	if err != nil {
		return err
	}
	if hash != j.sourceHash {
		return errors.New("original changed during backup; backup kept, replacement refused")
	}
	safety := filepath.Join(filepath.Base(stage), "world")
	// Write recovery instructions BEFORE exchange. A crash at the atomic
	// exchange leaves both trees, never an absent or half-merged live world.
	if err = os.WriteFile(filepath.Join(stage, "RECOVERY.txt"), []byte("World conversion "+j.ID+"\nServer: "+srv.ID+"\nLive world: "+j.Input+"\nBackup: "+backup+"\nAfter successful atomic exchange, world/ contains the original. Before exchange, world/ contains the converted candidate. Compare level.dat before manual recovery. Never delete without verification.\n"), 0600); err != nil {
		return err
	}
	err = l.fence(func() error {
		if !c.allowed(srv.ID) {
			return errors.New("world conversion module is disabled")
		}
		info, e := root.Lstat(j.Input)
		if e != nil || !os.SameFile(originalInfo, info) || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return errors.New("original directory changed before replacement")
		}
		if e = conversionCompatible(srv, j.Input, f, l.live); e != nil {
			return e
		}
		return conversionExchange(l.live, j.Input, fresh, "world")
	})
	if err != nil {
		return err
	}
	keep = true // From this point fresh is the original; NEVER run deferred cleanup.
	c.mu.Lock()
	j.SafetyCopy = safety
	c.mu.Unlock()
	return nil
}
func conversionSafetyRetention(base, id string) error {
	d, e := os.Open(base)
	if e != nil {
		return e
	}
	defer d.Close()
	count := 0
	seen := 0
	for {
		names, err := d.Readdirnames(128)
		seen += len(names)
		if seen > chunker.MaxEntries {
			return errors.New("recovery directory inspection limit exceeded")
		}
		for _, n := range names {
			if strings.HasPrefix(n, ".chunker-original-"+id+"-") {
				count++
			}
		}
		if count >= 3 {
			return errors.New("three world recovery trees retained; ask an administrator to review them before another replacement")
		}
		if len(names) == 0 {
			return errOrEOF(err)
		}
	}
}

type conversionContextWriter struct {
	ctx context.Context
	w   io.Writer
}

func (w conversionContextWriter) Write(p []byte) (int, error) {
	if e := w.ctx.Err(); e != nil {
		return 0, e
	}
	return w.w.Write(p)
}

// Back up the private, bounded original snapshot, not a changing live tree.
// The archive is world-only, and is only published after gzip/tar close + fsync.
func conversionWorldBackup(ctx context.Context, source, dir, name string) error {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, ".world-backup-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	defer f.Close()
	gz := gzip.NewWriter(conversionContextWriter{ctx, f})
	tw := tar.NewWriter(gz)
	err = filepath.Walk(source, func(p string, st os.FileInfo, e error) error {
		if e != nil {
			return e
		}
		if e = ctx.Err(); e != nil {
			return e
		}
		if p == source {
			return nil
		}
		if !st.IsDir() && !st.Mode().IsRegular() {
			return errors.New("backup snapshot has unsafe entries")
		}
		h, e := tar.FileInfoHeader(st, "")
		if e != nil {
			return e
		}
		h.Name, e = filepath.Rel(source, p)
		if e != nil {
			return e
		}
		h.Name = filepath.ToSlash(h.Name)
		if e = tw.WriteHeader(h); e != nil {
			return e
		}
		if st.IsDir() {
			return nil
		}
		in, e := os.Open(p)
		if e != nil {
			return e
		}
		_, e = rooted.Copy(ctx, tw, in, st.Size())
		return errors.Join(e, in.Close())
	})
	err = errors.Join(err, tw.Close(), gz.Close())
	if err != nil {
		return err
	}
	if err = f.Sync(); err != nil {
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return os.Link(f.Name(), filepath.Join(dir, name))
}
