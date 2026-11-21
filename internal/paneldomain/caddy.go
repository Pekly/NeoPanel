// Package paneldomain provides narrowly scoped, additive provisioning for the
// Linux installer's standard Caddy deployment. It never changes panel settings.
package paneldomain

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"
)

type Capability struct {
	Available bool     `json:"available"`
	Reason    string   `json:"reason"`
	Hosts     []string `json:"hosts"`
	Upstream  string   `json:"upstream"`
}

type ProvisionResult struct {
	Changed  bool   `json:"changed"`
	Message  string `json:"message"`
	Hostname string `json:"hostname"`
}

// Serialize even separate manager instances. The exclusive durable pending file
// additionally prevents another process from publishing a simultaneous change.
var operations = make(chan struct{}, 1)

func acquireOperation(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	select {
	case operations <- struct{}{}:
		// Cancellation and availability can become ready simultaneously.
		if err := ctx.Err(); err != nil {
			<-operations
			return err
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func releaseOperation() { <-operations }

type CaddyManager struct {
	upstream string
	setupErr error
	path     string
	platform func() error
	safe     func(string, pathKind) error
	run      func(context.Context, string, ...string) ([]byte, error)
	running  func(context.Context) ([]byte, error)
	service  func(context.Context) error
	read     func(string) (fileRecord, error)
	syncDir  func(string) error
}

func NewCaddyManager(listenerHost string, listenerPort int, panel bool, behindProxy bool) *CaddyManager {
	upstream, err := listenerUpstream(listenerHost, listenerPort, panel, behindProxy)
	m := &CaddyManager{upstream: upstream, setupErr: err, path: configPath, platform: supportedPlatform, safe: safePath, run: runCommand, running: runningConfig, syncDir: syncDirectory}
	m.service = m.inspectService
	m.read = m.snapshot
	return m
}

func (m *CaddyManager) pendingPath() string { return m.path + ".neopanel-pending" }

type inspection struct {
	original fileRecord
	hosts    []string
	adapted  []byte
}

func (m *CaddyManager) inspect(ctx context.Context) (inspection, error) {
	var s inspection
	if err := ctx.Err(); err != nil {
		return s, err
	}
	if m.setupErr != nil {
		return s, m.setupErr
	}
	if err := m.platform(); err != nil {
		return s, err
	}
	if err := m.pendingAbsent(); err != nil {
		return s, err
	}
	if err := m.service(ctx); err != nil {
		return s, err
	}
	var err error
	if s.original, err = m.read(m.path); err != nil {
		return s, err
	}
	if s.hosts, err = parseConfig(s.original.data, m.upstream); err != nil {
		return s, err
	}
	if s.adapted, err = m.adapt(ctx, m.path); err != nil {
		return s, err
	}
	active, err := m.running(ctx)
	if err != nil {
		return s, err
	}
	if !sameJSON(s.adapted, active) {
		return s, errors.New("Caddy running configuration differs from the Caddyfile; automatic provisioning refused")
	}
	if err = m.unchanged(s.original); err != nil {
		return s, err
	}
	return s, nil
}

// Capability is read-only: it inspects the service, filesystem, adapter output
// and local admin API, but does not validate/provision, reload or create files.
func (m *CaddyManager) Capability(ctx context.Context) Capability {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	cap := Capability{Hosts: []string{}, Upstream: m.upstream}
	if err := acquireOperation(ctx); err != nil {
		cap.Reason = err.Error()
		return cap
	}
	defer releaseOperation()
	s, err := m.inspect(ctx)
	if err != nil {
		cap.Reason = err.Error()
		return cap
	}
	cap.Available, cap.Hosts = true, s.hosts
	cap.Reason = "Standard installer Caddy deployment; new hostnames can be added without removing existing sites"
	if len(s.hosts) > maxAliases {
		cap.Reason = "Managed alias limit reached; existing hostnames remain idempotent"
	}
	return cap
}

func (m *CaddyManager) adapt(ctx context.Context, path string) ([]byte, error) {
	out, err := m.run(ctx, caddyBinary, "adapt", "--config", path, "--adapter", "caddyfile")
	if err != nil {
		return nil, fmt.Errorf("Caddy adaptation failed: %w", err)
	}
	if !sameJSON(out, out) {
		return nil, errors.New("Caddy adapter returned invalid or oversized JSON")
	}
	return out, nil
}

// Provision only appends a canonical site. The caller is responsible for admin
// authentication, confirmation, DNS/public-host policy and subsequent public TLS
// verification. Successful reload is not proof of certificate issuance.
func (m *CaddyManager) Provision(ctx context.Context, hostname, email string) (ProvisionResult, error) {
	result := ProvisionResult{Hostname: hostname}
	if !validHostname(hostname) {
		return result, errors.New("invalid hostname: use a lowercase ASCII DNS hostname without scheme, port, wildcard or trailing dot")
	}
	if !validEmail(email) {
		return result, errors.New("invalid ACME email: use a plain ASCII mailbox without display name or special quoting")
	}
	ctx, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()
	if err := acquireOperation(ctx); err != nil {
		return result, err
	}
	defer releaseOperation()
	s, err := m.inspect(ctx)
	if err != nil {
		return result, err
	}
	for _, host := range s.hosts {
		if host == hostname {
			result.Message = "Hostname is already configured; no rewrite or reload performed. Certificate status must be verified separately."
			return result, nil
		}
	}
	if len(s.hosts) > maxAliases {
		return result, errors.New("managed alias limit reached (8); use manual Caddy administration")
	}
	candidateBytes := append(append([]byte(nil), s.original.data...), []byte(aliasBlock(hostname, email, m.upstream))...)
	staged, err := m.stage(".neopanel-candidate-*", candidateBytes, &s.original)
	if err != nil {
		return result, err
	}
	defer os.Remove(staged)
	candidateJSON, err := m.adapt(ctx, staged)
	if err != nil {
		return result, err
	}
	if _, err = m.run(ctx, caddyBinary, "validate", "--config", staged, "--adapter", "caddyfile"); err != nil {
		return result, fmt.Errorf("candidate validation failed; original unchanged: %w", err)
	}
	candidate, err := m.read(staged)
	if err != nil {
		return result, err
	}
	if !bytes.Equal(candidate.data, candidateBytes) {
		return result, errors.New("candidate changed during validation")
	}
	// Recheck both disk and live state after validation, immediately before transaction publication.
	if err = m.service(ctx); err != nil {
		return result, err
	}
	active, err := m.running(ctx)
	if err != nil {
		return result, err
	}
	if !sameJSON(active, s.adapted) {
		return result, errors.New("running Caddy configuration changed concurrently")
	}
	if err = m.unchanged(s.original); err != nil {
		return result, err
	}
	if err = ctx.Err(); err != nil {
		return result, err
	}
	backup, err := m.stage(".neopanel-backup-*", s.original.data, nil)
	if err != nil {
		return result, err
	}
	// Never overwrite a prior backup or recovery record. A crash at any later
	// point leaves enough durable information for deliberate manual recovery.
	record := struct {
		Version         int    `json:"version"`
		Backup          string `json:"backup"`
		OriginalSHA256  string `json:"original_sha256"`
		CandidateSHA256 string `json:"candidate_sha256"`
		UID             int    `json:"uid"`
		GID             int    `json:"gid"`
		Mode            uint32 `json:"mode"`
	}{1, backup, fmt.Sprintf("%x", sha256.Sum256(s.original.data)), fmt.Sprintf("%x", sha256.Sum256(candidate.data)), s.original.uid, s.original.gid, uint32(s.original.info.Mode().Perm())}
	recovery, err := json.Marshal(record)
	if err != nil {
		return result, err
	}
	if err = m.writePending(recovery); err != nil {
		return result, fmt.Errorf("cannot reserve recovery state; backup retained at %s: %w", backup, err)
	}
	uncertain := func(cause error) (ProvisionResult, error) {
		return result, fmt.Errorf("Caddy provisioning outcome uncertain; do not retry automatically. Inspect recovery state %s and retained original backup %s; the generated candidate retains the old site. Administrator must reconcile disk and running config: %w", m.pendingPath(), backup, cause)
	}
	if err = m.publish(staged, s.original); err != nil {
		return uncertain(err)
	}
	if err = m.unchanged(candidate); err != nil {
		return uncertain(err)
	}
	_, reloadErr := m.run(ctx, systemctlBinary, "reload", "caddy.service")
	// Even a nonzero reload exit can mean the admin response was lost after
	// applying the config. A killed/timed-out systemctl may also leave a job
	// running. Establish the actual outcome before deciding to restore disk.
	// Reconcile with a fresh bounded context, independent of cancellation.
	reconcile, finish := context.WithTimeout(context.Background(), 10*time.Second)
	active, activeErr := m.running(reconcile)
	finish()
	if activeErr != nil {
		return uncertain(activeErr)
	}
	if sameJSON(active, candidateJSON) {
		if err = m.unchanged(candidate); err != nil {
			return uncertain(err)
		}
		if err = m.clearPending(recovery); err != nil {
			return uncertain(err)
		}
		result.Changed = true
		result.Message = "Hostname added and Caddy reloaded; existing sites preserved. Let's Encrypt certificate issuance is pending; verify public HTTPS before using the new address. Original backup retained."
		return result, nil
	}
	// Observing the original config does not establish that a systemd reload
	// job has finished: it may already hold the candidate and activate it later.
	// Keep disk and recovery state intact until completion can be established.
	if errors.Is(reloadErr, context.DeadlineExceeded) || errors.Is(reloadErr, context.Canceled) || ctx.Err() != nil {
		return uncertain(fmt.Errorf("reload completion is not established; candidate and recovery state retained (reload result: %v)", reloadErr))
	}
	if sameJSON(active, s.adapted) {
		if err = m.rollback(s.original, candidate, recovery); err != nil {
			return uncertain(err)
		}
		return result, fmt.Errorf("new Caddy configuration is not active; original Caddyfile restored and backup retained at %s (reload result: %v)", backup, reloadErr)
	}
	return uncertain(errors.New("running config matches neither original nor candidate"))
}

func (m *CaddyManager) rollback(original, candidate fileRecord, recovery []byte) error {
	staged, err := m.stage(".neopanel-rollback-*", original.data, &original)
	if err != nil {
		return err
	}
	defer os.Remove(staged)
	if err = m.publish(staged, candidate); err != nil {
		return err
	}
	return m.clearPending(recovery)
}
