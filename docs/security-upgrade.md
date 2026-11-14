# Security hardening and upgrade notes

This hardening pass fixes confirmed application vulnerabilities; it is not a certification that NeoPanel is free of vulnerabilities. Deploy the updated binary to **both the panel and every worker**. Local source changes do not protect an already-running installation.

## Before upgrading

1. Back up the panel database and configuration, plus server data. Test recovery before relying on the backup.
2. Build with **Go 1.26 or newer**, using a current patched toolchain. The installer checks the minimum version and verifies official compiler archives against their published SHA-256 checksum.
3. Set the panel's public URL in Admin settings or `server.public_url`. Account-email links no longer fall back to the request Host header. Password resets require working SMTP and a configured public URL; the public response deliberately does not disclose account existence or delivery failures.
4. Configure trusted reverse proxies explicitly, then restart. Existing configuration files are not silently rewritten by this change.

Example for a panel reachable only through local Caddy:

```yaml
server:
  host: "127.0.0.1"
  public_url: "https://panel.example.com"
security:
  trusted_proxies:
    - "127.0.0.1/32"
    - "::1/128"
```

Merge these entries into the existing sections; preserve secrets and other settings. For direct connections, leave `trusted_proxies` empty. For remote proxies, list only their actual addresses/networks, restrict direct backend access with a firewall, and make the edge proxy sanitize forwarded headers. Do **not** trust `0.0.0.0/0` or `::/0`. Client identity uses the first untrusted hop in `X-Forwarded-For` from a trusted peer. `X-Real-IP` and `CF-Connecting-IP` are not automatically trusted. Misconfigured proxy trust can group users into one rate-limit bucket or permit spoofing.

## Patched boundaries

- Password-reset tokens are never returned by the public request endpoint and become usable only after successful email delivery. Reset links require a fixed administrator-configured origin.
- Initial administrator creation, factor enrollment, password-reset consumption, profile updates, and recovery-code redemption use atomic database conditions to prevent the identified concurrency attacks.
- Password changes/resets revoke older JWT sessions and pending password-login 2FA tokens through an authentication version. Console WebSockets recheck credentials/permissions before commands and every second, including proxied connections.
- API keys remain independent credentials. Password changes do not delete them: revoke compromised API keys separately, and rotate worker credentials when needed.
- Untrusted console logs are escaped; locally generated command markup uses a separate rendering path. Dynamic filenames and other labels no longer enter inline JavaScript arguments.
- Server SQL batches reject filesystem/engine-control operations, including ATTACH, DETACH, VACUUM, PRAGMA, extension loading and file helpers. The executing SQLite connection also has attachment, SQL and value limits. Ordinary SQL remains supported. Query results are capped at 500 rows and an 8 MiB encoded-row budget. Schema inspection has a deadline, lists at most 500 objects, and does not execute views to count rows; view row counts are not computed.
- Non-Linux SFTP now uses pinned roots and parent handles, including root-identity checks and handle-based file attribute updates. Windows junction/root/ancestor replacement regressions pass; random-access uploads remain supported.
- File APIs, installers/downloaders, configuration writes and backups use descriptor-confined operations. Replacement writes publish a new inode instead of truncating planted hardlinks. Traversal and invalid software identifiers fail before cleanup or downloads.
- Archive extraction is staged and bounded: 256 MiB compressed input, 1 GiB expanded data, 10,000 entries, and 8 MiB ZIP central metadata. Links, special entries, duplicate members, ZIP64 and multidisk archives are rejected. Publication is atomic **per file**, not for the entire tree; cancellation during publication can leave a bounded partial extraction.
- Cross-filesystem file moves fail rather than falling back to an unsafe pathname copy. Move data through an explicit, checked copy workflow instead.
- Tenant-selected Java executables cannot request extra host ancestor/configuration mounts. Additional Linux Java runtime mounts require recognized, root-owned installations without group/world write permissions. Missing or unrecognizable proxy port configuration fails startup rather than silently using an unintended port.
- Installer logs/builds use a protected workspace under `/var/lib/neopanel/installer`, not predictable shared `/tmp` files. Update repository/branch inputs are validated and passed as arguments rather than interpolated shell source.
- Vulnerable SSH/JWT dependencies have been upgraded. Continue running `govulncheck ./...` with a current toolchain as new advisories appear.

## Isolation and interaction follow-up

- Node RAM restrictions now lock changes to custom JVM arguments as well as memory fields. Managed heap flags take precedence at startup; launcher operands and indirect option files are rejected. Convert existing argument-file configurations before restarting.
- Systemd-backed Java processes now receive a whole-process memory ceiling of **heap + max(512 MiB, 50% of heap)**. Reserve this overhead when sizing nodes. Steel retains its existing exact limit. See [memory and node policy details](../internal/runner/MEMORY_POLICY.md).
- Port add/delete/set-primary operations honor configured node restrictions. Rename, move, unzip, and extract require both `files.edit` and `files.delete`; frontend controls reflect the same requirements.
- Remote proxies carry the destination node's policy over the authenticated master-to-worker hop. Workers validate it with the forwarded identity; ordinary users cannot spoof policy headers or JWT fields. Update **both master and workers**. If tenants can contact workers directly, configure their local node policy too.
- Layout switches are local and burst-coalesced, preserve filters/selection, and use bounded sliding motion. Store searches reject stale responses and have a deadline. Dashboard refresh and power actions reject repeated clicks while in flight; power requests are never automatically retried.
- These guards are usability protections, not server-side rate limiting or additional authorization. They do not guarantee that networks or game processes cannot time out.
- **Known residual:** legacy UFW rules lack ownership markers. An authorized allocation deletion can still remove an administrator-created/shared rule. Until rule reconciliation is implemented, leave tenant port changes disabled and review firewall changes administratively.
- **Linux staging required:** verify namespaces, cgroup OOM behavior, resource budgets, and host firewall rules. Windows tests and Linux cross-compilation cannot prove runtime confinement. Unrestricted/bwrap-only modes have no cgroup memory ceiling.

## Operational limits

- Use Linux with the supported systemd/bubblewrap isolation and restrictive filesystem ownership for untrusted game workloads. Path confinement does not replace process, mount, resource or network isolation. Do not grant tenant processes authority over panel configuration, database, binaries, host mounts or sibling server directories.
- SFTP random-access writes retain normal file-handle semantics. A privileged local actor that can plant outside hardlinks or relocate open directories still requires OS ownership/mount isolation; SFTP must not be treated as isolation from such an actor.
- Panel extensions execute JavaScript with panel-origin authority by design. Only install trusted extensions. Admin ability to configure extensions, updates, remote nodes and credentials is privileged functionality, not a sandbox.
- Keep HTTPS enabled, constrain worker network reachability, rotate exposed passwords/API keys/node tokens, and use backups and audit logs. These changes do not establish whether an older deployment was compromised.
- Windows regression tests and Linux cross-compilation do not substitute for Linux runtime tests. Verify actual systemd/bubblewrap behavior and installer filesystem permissions on a staging Linux host before production rollout.

## Earlier hardening verification

- Toolchain used: Go 1.27.4 on Windows/amd64 (the declared minimum is Go 1.26).
- `go test ./... -count=1 -timeout 180s` and `go vet ./...`: passed.
- Frontend Node test suite: 180 passed.
- Mocked installer tests: 27 isolation checks and 8 security checks passed. The native installer symlink case was skipped on this Git Bash host.
- `govulncheck ./...`: no vulnerabilities found for Windows/amd64 or Linux/amd64 with the updated dependencies. This is a point-in-time advisory scan, not proof of absence of application vulnerabilities.
- Linux/amd64 panel/node binary built; API, database, rooted filesystem, runner, downloader and SFTP test binaries compiled; Linux-targeted vet passed.
- Exploit regressions used local temporary fixtures. No live-node exploit tests or deployment were performed.
- Linux runtime isolation and installer filesystem behavior still require staging-host verification.

## Follow-up verification

- Go 1.27.4: `go test ./...` and `go vet ./...` passed on Windows/amd64.
- Linux/amd64 panel and `-tags node` binaries built successfully; API/runner regression test binaries also cross-compiled. These are build checks, not Linux execution.
- Frontend suite: **199 tests passed**, including new interaction, stale-response, permission, timeout, and motion regressions.
- Fresh `govulncheck ./...` scans for Windows/amd64 and Linux/amd64 reported no vulnerabilities. This remains a point-in-time advisory check.
- Browser fixture checks: centered announcement icon; 40 rapid mode clicks produced one render and zero API requests; keyboard focus, file selection, and filters preserved. All three dashboard modes fit 375/768/1440 px; dark/light themes inspected. Reduced-motion behavior is covered by automated tests.
- No production deployment or live-node exploit testing was performed.
