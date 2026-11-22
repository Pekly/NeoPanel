package runner

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Pekly/NeoPanel/internal/database"
	"github.com/Pekly/NeoPanel/internal/downloader"
	"github.com/Pekly/NeoPanel/internal/rooted"
	"github.com/Pekly/NeoPanel/internal/serverjar"
)

const artifactTTL = 5 * time.Minute

var artifactSlots = make(chan struct{}, 2)
var ErrArtifactBusy = errors.New("server JAR inspection is busy; retry after the current file operation finishes")

type artifactCacheEntry struct {
	info    serverjar.Info
	file    os.FileInfo
	checked time.Time
}

type artifactCache struct {
	mu      sync.Mutex
	entries map[string]artifactCacheEntry
}

func artifactKey(srv *database.Server) string { return "server_artifact_v1:" + srv.ID }

func (m *ServerManager) savedArtifact(srv *database.Server) serverjar.Info {
	var info serverjar.Info
	if m.db != nil {
		raw, err := m.db.GetPanelSetting(artifactKey(srv))
		if err == nil && len(raw) < 16384 {
			_ = json.Unmarshal([]byte(raw), &info)
		}
	}
	if info.Filename != srv.JarFile {
		return serverjar.Info{}
	}
	return info
}

func artifactStat(dir, name string) (os.FileInfo, error) {
	if !rooted.Basename(name) {
		return nil, errors.New("invalid server JAR filename")
	}
	f, err := rooted.OpenFile(dir, name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	fi, err := f.Stat()
	if err == nil && !fi.Mode().IsRegular() {
		return nil, errors.New("server JAR is not a regular file")
	}
	return fi, err
}

func (m *ServerManager) cacheArtifact(srv *database.Server, info serverjar.Info) {
	fi, err := artifactStat(m.GetWorkingDir(srv.ID), srv.JarFile)
	if err != nil || info.Size != fi.Size() || !info.Modified.Equal(fi.ModTime()) {
		return
	}
	m.artifacts.mu.Lock()
	defer m.artifacts.mu.Unlock()
	if m.artifacts.entries == nil {
		m.artifacts.entries = make(map[string]artifactCacheEntry)
	}
	if len(m.artifacts.entries) >= 256 { // bounded cache; eviction only affects display, never integrity checks
		for id := range m.artifacts.entries {
			delete(m.artifacts.entries, id)
			break
		}
	}
	m.artifacts.entries[srv.ID] = artifactCacheEntry{info: info, file: fi, checked: time.Now()}
}

// ArtifactSummary performs no hashing or network work. Stat identity is a display
// cache hint, not integrity evidence: launch always performs a fresh inspection.
func (m *ServerManager) ArtifactSummary(srv *database.Server) *serverjar.Info {
	if srv == nil {
		return nil
	}
	m.artifacts.mu.Lock()
	entry, ok := m.artifacts.entries[srv.ID]
	m.artifacts.mu.Unlock()
	if !ok || time.Since(entry.checked) > artifactTTL || entry.info.Filename != srv.JarFile {
		return nil
	}
	fi, err := artifactStat(m.GetWorkingDir(srv.ID), srv.JarFile)
	if err != nil || !os.SameFile(entry.file, fi) || entry.file.Size() != fi.Size() || !entry.file.ModTime().Equal(fi.ModTime()) {
		m.InvalidateArtifact(srv.ID)
		return nil
	}
	info := entry.info
	return &info
}

func (m *ServerManager) InvalidateArtifact(id string) {
	m.artifacts.mu.Lock()
	delete(m.artifacts.entries, id)
	m.artifacts.mu.Unlock()
}

// ForgetArtifact removes a record when its server creation is rolled back or deleted.
func (m *ServerManager) ForgetArtifact(id string) {
	m.InvalidateArtifact(id)
	if m.db != nil {
		_ = m.db.DeletePanelSetting(artifactKey(&database.Server{ID: id}))
	}
}

func bindArtifact(local, verified serverjar.Info) serverjar.Info {
	if local.SHA256 == "" || !strings.EqualFold(local.SHA256, verified.SHA256) {
		return local
	}
	verified.SHA256, verified.Filename = local.SHA256, local.Filename
	verified.Size, verified.Modified = local.Size, local.Modified
	verified.MinimumJava = local.MinimumJava
	return verified
}

// RecordArtifactInfo is called only after a successful provider installation.
// Reinspection binds the metadata to the actual installed bytes, not a pathname.
// The caller retains its installation/maintenance lease through this call.
func (m *ServerManager) RecordArtifactInfo(srv *database.Server, info serverjar.Info) error {
	if info.Status == "native" && downloader.IsNativeSoftware(srv.Type) {
		fi, err := artifactStat(m.GetWorkingDir(srv.ID), srv.JarFile)
		if err != nil {
			return errors.New("installed native server could not be verified")
		}
		info.Filename, info.Size, info.Modified = srv.JarFile, fi.Size(), fi.ModTime()
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		defer cancel()
		local, err := serverjar.Inspect(ctx, m.GetWorkingDir(srv.ID), srv.JarFile)
		if err != nil || local.SHA256 == "" || !strings.EqualFold(local.SHA256, info.SHA256) {
			return errors.New("installed server JAR changed before its build metadata could be saved; inspect it again before using Auto Java")
		}
		info = bindArtifact(local, info)
	}
	if err := m.persistArtifact(srv, info); err != nil {
		return err
	}
	m.cacheArtifact(srv, info)
	return nil
}

func (m *ServerManager) persistArtifact(srv *database.Server, info serverjar.Info) error {
	if m.db == nil {
		return nil
	}
	raw, err := json.Marshal(info)
	if err != nil {
		return err
	}
	if err := m.db.SetPanelSetting(artifactKey(srv), string(raw)); err != nil {
		return errors.New("server build metadata could not be saved; retry inspection before using Auto Java")
	}
	return nil
}

// InspectArtifact permits a bounded digest-only provider lookup. It never runs a
// JAR and never changes configured software/version/runtime. No polling caller.
func (m *ServerManager) InspectArtifact(ctx context.Context, srv *database.Server, refresh ...bool) (serverjar.Info, error) {
	if info := m.ArtifactSummary(srv); info != nil && !(len(refresh) > 0 && refresh[0]) {
		return *info, nil
	}
	select {
	case artifactSlots <- struct{}{}:
		defer func() { <-artifactSlots }()
	default:
		return serverjar.Info{}, ErrArtifactBusy
	}
	release, err := rooted.TryExclusiveMutation(m.GetWorkingDir(srv.ID))
	if err != nil {
		return serverjar.Info{}, ErrArtifactBusy
	}
	defer release()
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	if downloader.IsNativeSoftware(srv.Type) {
		// Native identity is only claimed for a previously recorded installation.
		info := m.savedArtifact(srv)
		if info.Status == "native" {
			fi, err := artifactStat(m.GetWorkingDir(srv.ID), srv.JarFile)
			if err == nil && fi.Size() == info.Size && fi.ModTime().Equal(info.Modified) {
				m.cacheArtifact(srv, info)
				return info, nil
			}
		}
		return serverjar.Info{Status: "unknown", Filename: srv.JarFile, Message: "Native server identity has not been verified."}, nil
	}
	local, err := serverjar.Inspect(ctx, m.GetWorkingDir(srv.ID), srv.JarFile)
	if err != nil {
		return local, nil
	} // inspector supplies bounded, path-free diagnostics
	saved := m.savedArtifact(srv)
	if saved.Source == "mcjars" && local.SHA256 != "" && strings.EqualFold(local.SHA256, saved.SHA256) {
		local = bindArtifact(local, saved)
	} else if local.SHA256 != "" {
		build, lookupErr := downloader.LookupMCJarsBuild(ctx, local.SHA256)
		if lookupErr == nil && build != nil {
			candidate := build.Info
			candidate.SHA256 = local.SHA256 // exact digest lookup, never the response's latest build
			local = bindArtifact(local, candidate)
		} else if lookupErr != nil && !errors.Is(lookupErr, downloader.ErrMCJarsNotFound) {
			local.Message = "Build lookup is unavailable. Embedded JAR information is shown; retry later if Auto Java cannot resolve a requirement."
		}
	}
	// Recheck after remote work, so an external replacement cannot publish stale identity.
	current, err := serverjar.Inspect(ctx, m.GetWorkingDir(srv.ID), srv.JarFile)
	if err != nil || current.SHA256 == "" || current.SHA256 != local.SHA256 {
		m.InvalidateArtifact(srv.ID)
		return serverjar.Info{Status: "error", Message: "Server JAR changed during inspection; retry."}, nil
	}
	local = bindArtifact(current, local)
	if err := m.persistArtifact(srv, local); err != nil {
		return serverjar.Info{}, err
	}
	m.cacheArtifact(srv, local)
	return local, nil
}

func autoJavaMajor(info serverjar.Info) (string, error) {
	major := info.RequiredJava
	if major == 7 {
		major = 8
	} // obsolete Java 7 builds use the supported legacy Java 8 family
	switch major {
	case 8, 11, 16, 17, 21, 25:
		return fmt.Sprint(major), nil
	case 0:
		return "", errors.New("Auto Java could not verify an exact runtime requirement for this JAR. Inspect the JAR in server settings, or select a Java version manually")
	default:
		return "", fmt.Errorf("this build requires Java %d, which is not a supported managed runtime; install a compatible runtime and select its path manually", major)
	}
}

// resolveAutoJava is deliberately local-only, including restore/wake/retry starts.
func (m *ServerManager) resolveAutoJava(srv *database.Server) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	info, err := serverjar.Inspect(ctx, m.GetWorkingDir(srv.ID), srv.JarFile)
	if err != nil {
		return "", errors.New("Auto Java could not inspect the server JAR; check its filename and upload a valid JAR")
	}
	info = bindArtifact(info, m.savedArtifact(srv))
	m.cacheArtifact(srv, info)
	return autoJavaMajor(info)
}
