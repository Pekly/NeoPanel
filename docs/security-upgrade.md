# Security upgrade guide

The same security rules apply to free and premium servers. Update **the panel and every worker**; updating the repository alone does not protect running services.

**Audit follow-up:** see [panel-wide security fixes](security-audit.md) for credential-revocation races, local 2FA QR generation, SFTP/archive limits, managed file-operation coordination, and provider request bounds. **Users who enrolled through the older external QR setup should re-enroll 2FA after upgrading**, generating a new secret and recovery codes. No automatic factor reset is performed.

## Before upgrading

1. Back up the database, configuration and server data. Check that you can restore them.
2. Use a current patched Go toolchain, at least Go 1.26. The installer checks the minimum version and verifies compiler downloads against their published SHA-256 checksum.
3. Set the public HTTPS URL in Admin settings or `server.public_url`, and configure trusted proxies as described below.
4. On Linux, check that Bubblewrap, systemd-run, namespaces and cgroup delegation work before restarting games.
5. Deploy the binaries and restart the panel and workers during a maintenance window. Sandbox changes take effect on each game's next start, not on processes already running.

## HTTPS and reverse proxies

For a panel reachable only through local Caddy:

```yaml
server:
  host: "127.0.0.1"
  public_url: "https://panel.example.com"
security:
  trusted_proxies:
    - "127.0.0.1/32"
    - "::1/128"
```

Merge these entries into your existing configuration; keep its secrets and other settings. Existing files are not rewritten automatically.

For direct connections, leave `trusted_proxies` empty. For remote proxies, trust only their actual addresses, firewall the backend against direct access, and have the proxy sanitize forwarded headers. **Never trust `0.0.0.0/0` or `::/0`.**

Client identity comes from the first untrusted hop in `X-Forwarded-For`, and only when the socket peer is trusted. `X-Real-IP` and `CF-Connecting-IP` are not automatically trusted. Incorrect proxy settings can let clients spoof their address or put everyone in one rate-limit bucket.

Account emails use the configured public URL, not the request Host header. Password resets need working SMTP and a public URL. The public reset response does not reveal whether an account exists or email delivery succeeded.

## Linux and Windows hosting

### Linux

`auto` and `bwrap` require both Bubblewrap (`bwrap`) and `systemd-run`. Missing tools stop startup instead of silently switching to weaker isolation. Legacy Bubblewrap mode names still work.

Keep `security.allow_insecure_game_fallback` at its default, `false`, for public hosting. This host-only setting allows the old fallback behavior when explicitly enabled for trusted workloads. It is not a tenant or subscription setting.

Explicit `process`, `none` and systemd-only modes remain available, but do not provide a complete filesystem boundary between tenants. Unrestricted or bwrap-only fallback modes have no cgroup memory ceiling.

Systemd-backed Java processes have a whole-process memory ceiling of **heap + max(512 MiB, 50% of heap)**. Budget for that overhead. Steel retains its exact limit. See [memory and node policy details](../internal/runner/MEMORY_POLICY.md).

Node RAM restrictions also cover custom JVM arguments. Managed heap flags take precedence; launcher operands and indirect argument files are rejected. Convert argument-file configurations before restarting.

### Windows

Automatic mode uses Job Objects for supported process and resource management. **Job Objects are not a filesystem jail.** Untrusted Windows workloads need a separate, enforced tenant boundary, such as appropriately configured VMs.

### On either platform

- Use patched hosts, least-privilege service accounts and restrictive filesystem ownership. Game processes must not be able to alter panel configuration, databases, binaries or other servers' files.
- Enforce disk and inode quotas at the OS level. Restore and memory limits do not stop games or SFTP uploads from filling a disk.
- Restrict worker access and game-process network/egress access. Application limits are not DDoS protection.
- Keep tested off-host backups, review audit logs, and rotate exposed passwords, API keys and worker tokens.
- Keep tenant applications off the panel origin. Panel extensions run with panel-origin authority; install only trusted extensions. Administrative update, extension and node configuration is privileged access, not a sandbox.

**Linux staging is still required.** Verify namespaces, cgroups/OOM behavior, FIFO rejection, firewall rules and installer permissions on a Linux host before production rollout. Windows tests and Linux cross-compilation cannot prove runtime isolation. These changes also cannot establish whether an older deployment was compromised.

## Browser sessions and API clients

Cookie-authenticated writes need a matching `Origin`. When `Sec-Fetch-Site` is present, it must be `same-origin`; a sibling subdomain is not enough. A proxy that rewrites the backend Host needs explicit trusted-proxy configuration.

For scripts and API clients, send `Authorization: Bearer <token>` instead of relying on cookies. Query-string tokens remain supported for reads/downloads, but cannot authorize writes. Avoid tokens in URLs and access logs where possible.

Session cookies are HttpOnly and SameSite=Lax. Secure is enabled for TLS, trusted forwarded HTTPS, or a configured HTTPS public URL. Profile renewal uses the canonical `neopanel_token` cookie and clears the old duplicate `jwt` cookie when present.

Responses block framing, restrict base URLs, suppress referrers and MIME sniffing, and disable unused browser capabilities. APIs default to private/no-store. HSTS is sent only over secure transport. The CSP does **not** block inline scripts: the SPA and trusted extensions still use inline handlers.

## Accounts and credentials

Only admins can create servers. Users retain their owned/shared servers and team permissions. Old custom-role assignments are removed on upgrade without deleting accounts, credentials, API keys or servers. Invitation-created accounts remain independent users.

Password changes and resets revoke older JWT sessions and pending password-login 2FA tokens through an authentication version. Console WebSockets recheck credentials and permissions before commands and every second, including proxied connections.

New SSH/API keys are inserted only if the original authenticated user ID and JWT version, or authenticating API key, are still valid. A request delayed while reading its body cannot enroll a key after that authority is revoked.

**API keys are independent of password resets.** Revoke compromised API and SSH keys separately. Previously created credentials are not cascade-deleted.

Administrator creation, factor enrollment, profile updates, reset consumption and recovery-code redemption use atomic database checks. Reset tokens are never returned by the public request endpoint and become usable only after successful email delivery.

## Uploads, files and restores

### Avatars

Accepted formats are PNG, JPEG, GIF and WebP, with these limits:

| Limit | Value |
| --- | --- |
| Uploaded file | 5 MiB |
| Width or height | 4096 pixels |
| Total pixels | 4,194,304 |
| Saved image | Static PNG, at most 512 pixels per side |

Animation is removed. Request size and processing concurrency are bounded; multipart temporary files are cleaned up. Writes replace files atomically rather than truncate an existing hardlink. Older SVG uploads are retained as sandboxed downloads, not active panel-origin images.

### File access and startup

File APIs, installers, downloaders, configuration writes and backups use descriptor-confined operations. Replacement writes do not truncate planted hardlinks. Traversal and invalid software identifiers are rejected before cleanup or downloads. Cross-filesystem moves fail rather than fall back to an unsafe pathname copy; use an explicit checked copy workflow instead.

SFTP on non-Linux systems uses pinned roots and parent handles, with root-identity checks and handle-based attribute updates. Random-access uploads still have normal open-file semantics: a privileged local actor that can plant outside hardlinks or relocate open directories needs OS ownership/mount isolation, not just SFTP path checks.

Java startup reads `server.properties` as a rooted regular file, capped at 1 MiB. Unsafe links, special files and lock targets are rejected; safe relative nested world paths still work. Missing or unrecognized proxy port configuration stops startup rather than selecting an unintended port.

Tenant-selected Java executables cannot add host ancestor/configuration mounts. Extra Linux Java runtime mounts require recognized, root-owned installations without group/world write permissions.

### Archives and restore recovery

Archive extraction is staged and limited to 256 MiB compressed input, 1 GiB expanded data, 10,000 entries and 8 MiB ZIP central metadata. Links, special entries, duplicate members, ZIP64 and multidisk archives are rejected. Publication is atomic **per file**, not for the whole tree; cancellation can leave a bounded partial extraction.

Restores retain at most **three recovery directories per server**. At that limit, another restore is refused before staging. Inspect and archive or remove older recovery copies before retrying. NeoPanel never deletes live data or recovery copies automatically to make room. This cap is not a disk quota.

Rename, move, unzip and extract require both `files.edit` and `files.delete`. Port changes honor node restrictions. The authenticated panel-to-worker hop carries destination policy; users cannot supply their own policy headers. If tenants can access workers directly, configure each worker's local policy too.

**Known firewall limitation:** legacy UFW rules have no ownership markers. An authorized allocation deletion can remove a shared or administrator-created rule. Leave tenant port changes disabled and review firewall changes administratively until rule reconciliation is available.

## Request, console and SQL limits

| Resource | Limit or behavior |
| --- | --- |
| General control API body | 1 MiB; explicit allowances for file uploads, editor, SQL, image and mod routes |
| Worker-proxied bodies | Same route limits as local requests |
| HTTP headers | Five-second read timeout; 32 KiB limit |
| Rate-limiter state | Bounded client maps; active quota counters are not evicted to accept new clients |
| Console output | 16 KiB per line; retained history capped at 1 MiB and 1000 lines |
| SQL results | 500 rows; 8 MiB encoded-row budget |
| SQL schema inspection | Deadline and 500-object cap; views are not executed to count rows |

Declared oversized request bodies return HTTP 413. Streamed bodies are bounded too. Oversized console lines are drained so later readiness/player events can still be read. Console text is escaped; generated command markup has a separate rendering path. Filenames and labels are not interpolated into inline JavaScript arguments.

SQL batches reject filesystem and engine-control operations, including ATTACH, DETACH, VACUUM, PRAGMA, extension loading and file helpers. The SQLite connection also has attachment, SQL and value limits. Ordinary queries remain supported.

Dashboard refresh, power actions and store searches also guard against repeated clicks and stale responses. Power requests are not automatically retried. These UI guards do not replace server-side authorization or rate limiting.

## Installer and update safety

Installer builds and logs use a protected workspace under `/var/lib/neopanel/installer`, not predictable shared `/tmp` paths. Update repository/branch inputs are validated and passed as arguments, not interpolated into shell source.

Keep dependencies and the toolchain patched. Run `govulncheck ./...` regularly; a clean scan only describes the advisories known at that time.

## Verification record

The shared-hosting hardening commits (`0600ae1` on main, `218510a` on Windows) were checked with Go 1.27.4 on Windows/amd64:

- Full Go tests and `go vet` passed on both branches.
- Frontend tests passed: 461 on main, 463 on Windows, including browser-backed layout tests.
- Windows and Linux panel/worker binaries built. Linux regression test packages cross-compiled and Linux-targeted vet passed.
- Windows and Linux `govulncheck` scans reported no vulnerabilities.
- Tests used temporary local fixtures. No production deployment or live-node exploit testing was performed; Linux runtime isolation remains untested on this development host.

Earlier passes also checked installer behavior (27 isolation and 8 security checks; the native symlink case was skipped under Git Bash), Windows junction/root/ancestor replacement, and browser layout/interaction behavior. Their frontend suites had 180 and 199 passing tests respectively; those are historical counts, not the current suite size.

These checks are regression evidence, not a security certification. Keep host controls in place and continue testing as the code, dependencies and deployment change.
