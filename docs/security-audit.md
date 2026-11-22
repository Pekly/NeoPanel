# Panel-wide security audit follow-up

This follow-up reviews high-risk boundaries across authentication and credential revocation, account/server/admin permissions, worker forwarding, browser rendering, provider metadata, files and archives, SFTP/import, and process/conversion isolation. It fixes the concrete issues below; it is not a certification that every vulnerability has been found. Free and premium servers receive the same safeguards.

Update **the panel and every worker**. Repository publication does not update running binaries. Follow the [security upgrade guide](security-upgrade.md), preserve existing configuration and secrets, and deploy during an operator-controlled maintenance window.

## Required action for existing 2FA users

Older QR setup sent the full `otpauth` URI, including the TOTP secret, to an external QR-image service. New setup renders the QR PNG locally on the panel and embeds it as a bounded data URI; no external QR request or fallback is used. Older/malformed responses offer manual-key setup instead. Setup responses are non-cacheable, and stale account/dialog responses cannot display another setup's secret or recovery codes.

**If you enrolled using the old QR setup, re-enroll 2FA after upgrading to generate a new secret and recovery codes.** Confirm you have working account access and recovery information before changing the factor, then replace the authenticator entry and store the new codes safely. Existing 2FA is not silently disabled or reset. The exposure is established; this audit does not establish that it was exploited.

## Credential revocation races

Profile changes and 2FA activation now capture the authenticating credential before decoding the body and validate it in the same SQL mutation. A JWT must retain its original authentication version; an API key must still exist for that user. Revocation while a request is paused reading its body returns 401 without profile changes, factor enrollment, a new session cookie/token, or recovery codes. An already-enabled factor returns 409 without replacement.

Profile session generation uses the committed user snapshot rather than a later fresh read that could accidentally upgrade an obsolete session. API keys intentionally remain independent of password changes; revoke a compromised API key separately.

## SFTP resource abuse

- At most 64 pending SSH handshakes, 8 per source IP, 256 total connections, and 8 channels per connection.
- A 30-second socket deadline covers negotiation/authentication; successful authentication clears it for normal long transfers. Shutdown closes tracked sockets and joins workers.
- WRITE, writable OPEN and size-changing SETSTAT/FSETSTAT enforce a logical per-file ceiling of **8 GiB or a smaller positive assigned disk limit**. Sparse holes count, integer overflow is rejected, and old handles recheck the current assignment.
- Existing oversized files remain readable/deletable and can be explicitly shrunk. Repeated legitimate replacement uploads do not consume a lifetime allowance.

These are local-listener and per-file bounds, **not aggregate disk quotas**. Multiple files, existing bytes, game/API/external writes, and other processes can still fill a volume. Use OS disk/inode quotas. Deadlines do not interrupt a database/password callback already executing, although admission bounds concurrency. See [SFTP limits](../internal/sftp/LIMITS.md).

## Backup and compression amplification

TAR/GZIP backups and ZIP/TAR.GZ file compression now preflight logical input before reading/compressing file contents, then check sizes again on opened handles and bound each streamed copy. A sparse file no longer bypasses the input budget through its small allocated size. Changed-size inputs fail without publishing an incomplete archive.

| Bound | Value |
| --- | --- |
| Logical input per job | 100 GiB, or smaller positive server disk assignment |
| Visited entries | 100,000 |
| Path depth | 64 |
| Cooperative deadline | 30 minutes |
| Shared archive workers | 2 per daemon |
| Concurrent managed archive/maintenance work | One operation per server; exclusive managed-write lease |

Directory enumeration uses bounded batches. HTTP compression honors request cancellation; managed backup/compression work also participates in daemon-shutdown cancellation and joining. Failed jobs remove unpublished temporary outputs. Existing formats and no-replace backup publication are retained. No game is automatically stopped to make an archive.

These limits bound workload, not total retained backup storage, daemon CPU/RAM through cgroups, or application-consistent snapshots of running games. Cancellation is cooperative and cannot forcibly interrupt a blocked filesystem syscall. Game/external writes remain outside the managed-write lease.

## Directory replacement and managed writers

Restore, SFTP-import publication, reinstall, and deletion now combine the existing server-operation gate with the same canonical exclusive mutation lease used by managed file writers. An active editor/SFTP write prevents a directory swap; maintenance prevents new managed writes. Missing directories derive a key from their existing canonical parent so creation cannot evade an already-held lease. Safety copies are retained under the existing recovery policy.

Leases are process-local coordination, not an OS filesystem lock. External host edits, other daemons and unsupported game processes require host isolation. Windows Job Objects still are **not a filesystem jail**. Existing Linux Bubblewrap/systemd requirements and Linux-only guarded world conversion are unchanged.

## Provider request amplification

Version catalogs share a single deadline-bounded fetch. The usual 30-minute cache remains, and explicit refresh is throttled to once per minute per daemon. Canceled waiters leave without canceling the shared fetch. Version and Modrinth metadata are bounded to 4 MiB; malformed or concatenated JSON is rejected.

## Verification and remaining coverage limits

Regressions use temporary databases/directories, mocked provider responses, loopback sockets, and real Chrome with mocked APIs. They cover delayed-body credential revocation, local QR/no external fallback and stale response isolation, provider coalescing/oversized JSON, sparse WRITE/truncate and archive input, ordinary repeat uploads, connection admission/release, cancellation, and managed mutation exclusion.

Full Go/frontend tests, vet, Windows/Linux panel and worker builds, module verification and vulnerability scans are run for publication. Linux test binaries are cross-compiled, **not executed on Linux** on this Windows development host. Race-detector execution requires a CGO toolchain not available here. No production attacks, service/game restarts, firewall/DNS changes or live-node exploitation are part of these checks.

A clean advisory scan only covers known advisories at scan time. Continue Linux staging/runtime isolation tests and enforce least privilege, OS quotas, patched hosts, backups, network restrictions and trusted-proxy configuration. See the upgrade guide for other existing deployment limitations.
