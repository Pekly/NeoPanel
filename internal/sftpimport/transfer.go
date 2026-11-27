package sftpimport

import (
	"context"
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	"unicode"

	"github.com/pkg/sftp"
)

func (r *remote) checkDirs(p string) error {
	for {
		st, err := r.client.Lstat(p)
		if err != nil || st.Mode().Type() != os.ModeDir {
			return problem("remote directory changed or is unsafe")
		}
		if p == "/" || p == "." {
			return nil
		}
		p = path.Dir(p)
	}
}
func (r *remote) list(p string) ([]os.FileInfo, error) {
	r.frames.listing.Store(true)
	defer r.frames.listing.Store(false)
	entries, err := r.client.ReadDir(p)
	if err != nil {
		return nil, problem("remote listing failed or exceeded limits")
	}
	return entries, nil
}
func usable(st os.FileInfo) bool {
	return st != nil && (st.Mode().Type() == os.ModeDir || st.Mode().IsRegular()) && st.Size() >= 0
}
func same(a, b os.FileInfo) bool {
	return usable(a) && usable(b) && a.Mode() == b.Mode() && a.Size() == b.Size() && a.ModTime().Equal(b.ModTime())
}
func accessTime(st os.FileInfo) time.Time {
	if attrs, ok := st.Sys().(*sftp.FileStat); ok {
		return time.Unix(int64(attrs.Atime), 0)
	}
	return st.ModTime()
}
func within(p, subtree string) bool { return p == subtree || strings.HasPrefix(p, subtree+"/") }
func selected(o Options, p string, dir bool) bool {
	if o.SkipHidden {
		for _, c := range strings.Split(p, "/") {
			if strings.HasPrefix(c, ".") {
				return false
			}
		}
	}
	for _, x := range o.ExcludePaths {
		if within(p, x) {
			return false
		}
	}
	if len(o.IncludePaths) == 0 {
		return true
	}
	for _, x := range o.IncludePaths {
		if within(p, x) || dir && within(x, p) {
			return true
		}
	}
	return false
}

// foldKey uses Unicode simple folding, not just lowercase (e.g. Kelvin sign,
// long S). Used on case-insensitive filesystems (Windows/macOS).
func foldKey(p string) string {
	return strings.Map(func(r rune) rune {
		r = unicode.ToUpper(r) // Include Windows upcase collisions (e.g. dotless I).
		min := r
		for s := unicode.SimpleFold(r); s != r; s = unicode.SimpleFold(s) {
			if s < min {
				min = s
			}
		}
		return min
	}, p)
}

// collisionKey folds case only on case-insensitive host OSes (Windows/macOS);
// Linux filesystems are case-sensitive so distinct-case filenames are valid.
func collisionKey(p string) string {
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		return foldKey(p)
	}
	return p
}

// Probe lists only the immediate remote root. Entries are relative to RemotePath.
// Filters are applied as in Transfer. Very large directories fail the bounded
// metadata/entry budget rather than consuming unbounded memory.
func Probe(ctx context.Context, options Options, allowedHosts []string) (preview Preview, err error) {
	o, err := Validate(options)
	if err != nil {
		return preview, err
	}
	if o.Protocol == "ftp" || o.Protocol == "ftps" {
		return probeFTP(ctx, o, allowedHosts)
	}
	ctx, cancel := context.WithTimeout(ctx, time.Duration(o.TimeoutMinutes)*time.Minute)
	defer cancel()
	defer func() {
		if ctx.Err() != nil {
			err = ctx.Err()
		}
	}()
	r, err := connect(ctx, o, allowedHosts)
	if err != nil {
		return preview, err
	}
	defer r.close()
	preview.RemotePath = r.root
	preview.Entries = []Entry{}
	entries, err := r.list(r.root)
	if err != nil {
		return preview, err
	}
	seen := map[string]bool{}
	exactSeen := map[string]bool{}
	for _, st := range entries {
		if !safeComponent(st.Name()) {
			return Preview{}, problem("unsafe remote entry")
		}
		if exactSeen[st.Name()] {
			continue
		}
		exactSeen[st.Name()] = true
		key := collisionKey(st.Name())
		if seen[key] {
			return Preview{}, problem("case-fold path collision: " + st.Name())
		}
		seen[key] = true
		if !selected(o, st.Name(), st.IsDir()) {
			continue
		}
		if !usable(st) {
			return Preview{}, problem("symlink or special remote entry denied")
		}
		if len(preview.Entries) == 200 {
			preview.Truncated = true
			continue
		}
		preview.Entries = append(preview.Entries, Entry{st.Name(), st.Name(), st.IsDir(), st.Size()})
	}
	return preview, nil
}

type item struct {
	rel  string
	info os.FileInfo
}

// Transfer requires an existing empty private directory. The caller must not
// allow other processes to mutate it. Writes are os.Root-confined, exclusive and
// sequential. Errors leave an unpublished partial tree for caller cleanup.
// Unix mode bits are checked; on Windows the caller must enforce private ACLs.
// Progress callbacks run synchronously and must return promptly (not block).
func Transfer(ctx context.Context, options Options, allowedHosts []string, destination string, progress func(Progress)) (err error) {
	o, err := Validate(options)
	if err != nil {
		return err
	}
	if o.Protocol == "ftp" || o.Protocol == "ftps" {
		return transferFTP(ctx, o, allowedHosts, destination, progress)
	}
	ctx, cancel := context.WithTimeout(ctx, time.Duration(o.TimeoutMinutes)*time.Minute)
	defer cancel()
	defer func() {
		if ctx.Err() != nil {
			err = ctx.Err()
		}
	}()
	local, err := destinationRoot(destination)
	if err != nil {
		return err
	}
	defer local.Close()
	r, err := connect(ctx, o, allowedHosts)
	if err != nil {
		return err
	}
	defer r.close()
	p := Progress{State: "scanning"}
	emit := func() {
		if progress != nil {
			progress(p)
		}
	}
	emit()
	var plan []item
	seen := map[string]bool{}
	exactSeen := map[string]bool{}
	count := 0
	pathBytes := 0
	var scan func(string, int) error
	scan = func(rel string, depth int) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if depth > 64 {
			return problem("remote depth limit exceeded")
		}
		full := path.Join(r.root, rel)
		if len(full) > 4096 {
			return problem("remote path too long")
		}
		if err := r.checkDirs(full); err != nil {
			return err
		}
		entries, e := r.list(full)
		if e != nil {
			return e
		}
		for _, listed := range entries {
			count++
			if count > maxEntries {
				return problem("remote entry limit exceeded")
			}
			name := listed.Name()
			child := path.Join(rel, name)
			if !safeComponent(name) || !safePath(child, false) || len(path.Join(r.root, child)) > 4096 {
				return problem("unsafe remote path")
			}
			pathBytes += len(child)
			if pathBytes > maxMetadata {
				return problem("remote path metadata limit exceeded")
			}
			if exactSeen[child] {
				continue
			}
			exactSeen[child] = true
			key := collisionKey(child)
			if seen[key] {
				return problem("case-fold path collision: " + child)
			}
			seen[key] = true
			if !selected(o, child, listed.IsDir()) {
				continue
			}
			if !usable(listed) {
				return problem("symlink or special remote entry denied")
			}
			st, e := r.client.Lstat(path.Join(r.root, child))
			if e != nil || !same(st, listed) {
				return problem("remote entry changed")
			}
			if depth+1 > 64 {
				return problem("remote depth limit exceeded")
			}
			plan = append(plan, item{child, st})
			if !st.IsDir() {
				if st.Size() > o.MaxBytes-p.BytesTotal {
					return problem("remote byte limit exceeded")
				}
				p.FilesTotal++
				p.BytesTotal += st.Size()
			}
			p.CurrentPath = child
			emit()
			if st.IsDir() {
				if e = scan(child, depth+1); e != nil {
					return e
				}
			}
		}
		return nil
	}
	rootInfo, e := r.client.Lstat(r.root)
	if e != nil || !usable(rootInfo) {
		return problem("remote root unavailable")
	}
	if err = scan("", 0); err != nil {
		return err
	}
	p.State = "transferring"
	p.CurrentPath = ""
	emit()
	start := time.Now()
	chunkSize := 32768
	if o.BandwidthKiB > 0 && o.BandwidthKiB < 32 {
		chunkSize = o.BandwidthKiB * 1024
	}
	buffer := make([]byte, chunkSize)
	for _, it := range plan {
		if err = ctx.Err(); err != nil {
			return err
		}
		p.CurrentPath = it.rel
		emit()
		full := path.Join(r.root, it.rel)
		if err = r.checkDirs(path.Dir(full)); err != nil {
			return err
		}
		st, e := r.client.Lstat(full)
		if e != nil || !same(st, it.info) {
			return problem("remote entry changed")
		}
		if it.info.IsDir() {
			if e = local.Mkdir(filepath.FromSlash(it.rel), 0755); e != nil {
				return problem("cannot create staging directory")
			}
			if e = local.Chmod(filepath.FromSlash(it.rel), 0755); e != nil {
				return problem("cannot set staging permissions")
			}
			continue
		}
		if err = download(ctx, r, local, o, it, buffer, &p, start, emit); err != nil {
			return err
		}
		p.FilesDone++
		emit()
	}
	// Recheck directories after reads, then apply times bottom-up (children must
	// not subsequently change the preserved parent directory timestamps).
	for i := len(plan) - 1; i >= 0; i-- {
		it := plan[i]
		if it.info.IsDir() {
			st, e := r.client.Lstat(path.Join(r.root, it.rel))
			if e != nil || !same(st, it.info) {
				return problem("remote directory changed")
			}
			if o.PreserveTimes {
				if e = local.Chtimes(filepath.FromSlash(it.rel), accessTime(it.info), it.info.ModTime()); e != nil {
					return problem("cannot preserve directory time")
				}
			}
		}
	}
	st, e := r.client.Lstat(r.root)
	if e != nil || !same(st, rootInfo) {
		return problem("remote root changed")
	}
	if o.PreserveTimes {
		if e = local.Chtimes(".", accessTime(rootInfo), rootInfo.ModTime()); e != nil {
			return problem("cannot preserve root time")
		}
	}
	return nil
}
func destinationRoot(destination string) (*os.Root, error) {
	st, err := os.Lstat(destination)
	if err != nil || st.Mode().Type() != os.ModeDir || runtime.GOOS != "windows" && st.Mode().Perm()&0077 != 0 {
		return nil, problem("staging directory must exist and be private")
	}
	root, err := os.OpenRoot(destination)
	if err != nil {
		return nil, problem("cannot open staging directory")
	}
	ok := false
	defer func() {
		if !ok {
			_ = root.Close()
		}
	}()
	dir, err := root.Open(".")
	if err != nil {
		return nil, problem("cannot open staging directory")
	}
	defer dir.Close()
	opened, err := dir.Stat()
	if err != nil || !os.SameFile(st, opened) {
		return nil, problem("staging directory changed")
	}
	entries, err := dir.ReadDir(1)
	if err != io.EOF || len(entries) != 0 {
		return nil, problem("staging directory must be empty")
	}
	ok = true
	return root, nil
}
func download(ctx context.Context, r *remote, local *os.Root, o Options, it item, buffer []byte, p *Progress, start time.Time, emit func()) error {
	full := path.Join(r.root, it.rel)
	source, err := r.client.Open(full)
	if err != nil {
		return problem("cannot open remote file")
	}
	defer source.Close()
	st, err := source.Stat()
	if err != nil || !same(st, it.info) {
		return problem("remote file changed")
	}
	target, err := local.OpenFile(filepath.FromSlash(it.rel), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return problem("cannot create staging file")
	}
	defer target.Close()
	if err = target.Chmod(0644); err != nil {
		return problem("cannot set staging permissions")
	}
	var copied int64
	var lastSpeedBytes int64
	lastSpeedTick := start
	for {
		if err = ctx.Err(); err != nil {
			return err
		}
		want := int64(len(buffer))
		if remaining := it.info.Size() - copied + 1; remaining < want {
			want = remaining
		}
		r.frames.readRemaining.Store(want)
		n, readErr := source.Read(buffer[:int(want)]) // No io.Copy/WriteTo or concurrent readahead.
		if n > 0 {
			if int64(n) > o.MaxBytes-p.BytesDone || int64(n) > it.info.Size()-copied {
				return problem("remote file grew or byte limit exceeded")
			}
			if _, err = target.Write(buffer[:n]); err != nil {
				return problem("cannot write staging file")
			}
			copied += int64(n)
			p.BytesDone += int64(n)
			now := time.Now()
			if elapsed := now.Sub(lastSpeedTick); elapsed >= 250*time.Millisecond {
				p.SpeedBPS = int64(float64(p.BytesDone-lastSpeedBytes) / elapsed.Seconds())
				lastSpeedBytes = p.BytesDone
				lastSpeedTick = now
			} else if p.SpeedBPS == 0 {
				if tot := now.Sub(start).Seconds(); tot > 0.05 {
					p.SpeedBPS = int64(float64(p.BytesDone) / tot)
				}
			}
			emit()
			if o.BandwidthKiB > 0 {
				// Base pacing on cumulative bytes, avoiding per-chunk rounding drift.
				delay := time.Duration(float64(p.BytesDone)/float64(int64(o.BandwidthKiB)*1024)*float64(time.Second)) - time.Since(start)
				if delay > 0 {
					timer := time.NewTimer(delay)
					select {
					case <-ctx.Done():
						timer.Stop()
						return ctx.Err()
					case <-timer.C:
					}
				}
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return problem("remote file read failed")
		}
		if n == 0 {
			return problem("remote file read stalled")
		}
	}
	if copied != it.info.Size() {
		return problem("remote file size changed")
	}
	st, err = source.Stat()
	if err != nil || !same(st, it.info) {
		return problem("remote file changed")
	}
	st, err = r.client.Lstat(full)
	if err != nil || !same(st, it.info) {
		return problem("remote file changed")
	}
	if err = r.checkDirs(path.Dir(full)); err != nil {
		return err
	}
	if err = target.Close(); err != nil {
		return problem("cannot close staging file")
	}
	if o.PreserveTimes {
		if err = local.Chtimes(filepath.FromSlash(it.rel), accessTime(it.info), it.info.ModTime()); err != nil {
			return problem("cannot preserve file time")
		}
	}
	return nil
}
