# Per-server file trash (24 hours)

## HTTP contract

All routes use the existing authentication, ownership/admin bypass and node identity forwarding.
Base: `/api/servers/{id}/files` (or `/api/nodes/{nodeId}/proxy/api/servers/{id}/files`).

| Request | Permission | Result |
| --- | --- | --- |
| `DELETE ?path=...` | `files.delete` | Move to trash; `200 {"message":"Deleted successfully"}` |
| `DELETE ?path=...&permanent=true` | `files.delete` | Permanently remove, without a trash entry |
| `POST /delete-bulk {"paths":[...]}` | `files.delete` | Move to trash, preserving `200 {"deleted":[...],"failed":[{"path":"...","error":"..."}]}` |
| `POST /delete-bulk {"paths":[...],"permanent":true}` | `files.delete` | Same bulk response, permanent removal |
| `GET /trash` | `files.view` AND `files.delete` | `200 {"items":[...]}`, newest first, empty array when empty |
| `POST /trash/{trashId}/restore {}` | `files.view` AND `files.edit` AND `files.delete` | Original-path restore, `200 {"message":"Restored successfully"}` |
| `DELETE /trash/{trashId}` | `files.delete` | Permanent purge, `200 {"message":"Permanently deleted"}` |

List items have exactly `id`, `original_path`, `name`, `is_dir`, `size`, `deleted_at`, `expires_at`.
IDs are opaque lowercase 32-character cryptographically random hex strings. Paths are server-relative,
normalized with `/`, without a leading slash. Times are RFC3339/RFC3339Nano UTC strings. `size` is
filesystem entry size at deletion (for directories, **not** a recursively computed byte total).

Restore takes no destination option; the body is unused. It never creates missing parent directories
and never overwrites any destination, including a dangling symlink. Collision is `409`; expired entry
is `410` until physically swept, then `404`. Missing parent, unknown/invalid ID and an ID from another
server return `404`; invalid metadata or forbidden symlink ancestors return `400`. Listing hides
expired entries immediately; it does not extend their lifetime. Ordinary single-delete errors retain
the existing `400` validation / `500` filesystem-error behavior. Purging an absent entry returns `404`.

Restore acquires the server-operation gate before opening the tenant root; active reinstall, restore,
backup or other gated maintenance returns `409` without consuming the Trash entry. Non-UTF-8 source
paths are rejected before moving anything, so JSON metadata cannot silently change the original name.

Missing ordinary deletion targets are still successful and create no trash entry. Bulk validates the
whole batch before mutation, keeps original request paths/order/duplicates in `deleted`, and preserves
partial filesystem failure reporting, 1 MiB body cap, 1000-path cap, and one audit summary per request.
Overlapping paths are idempotent; restoring separately deleted child/parent entries may require the
parent entry first. No `trashed` response field was added.

## Storage and safety

For configured servers directory `/data/servers`, payloads live under
`/data/servers-trash/<sha256(server-id)>/<entry-id>/payload`, with adjacent `metadata.json`.
This is outside every tenant working root and therefore outside rooted file/SFTP APIs and ordinary
server backups. Keep this panel-owned directory when migrating/restarting a node. Store directories
are created `0700` and metadata `0600`; Windows deployments must also protect the panel data directory
with appropriate service-account ACLs (Go Unix permission bits do not establish Windows ACLs).

Metadata is written exclusively to a fresh entry and file-synced **before** the move. A metadata failure
leaves the original in place; a move failure removes the unused entry and reports failure. Restore moves
the payload before metadata cleanup; leftover metadata without a payload cannot be restored again.
Normal daemon restart preserves entries/timestamps. This is not a backup or a power-loss transaction:
there is no multi-directory fsync journal, and open game-process handles or hardlinks can continue
changing the same inode after it is moved. Corrupt metadata is hidden and retained for operator recovery
rather than guessing a destination/expiry or deleting potentially recoverable data.

All entry lookups are server-hash scoped; client IDs are validated, not concatenated as arbitrary paths.
Metadata paths are lexically validated and re-resolved through pinned tenant parent directories.
Parents are opened component-by-component with identity checks. Symlink ancestors/root modification
are refused. Atomic Linux `renameat2(RENAME_NOREPLACE)` and Windows handle-relative NT rename protect
restore publication against collision races; no payload inode is opened for content writes. Directory
contents (including links) move without traversal, and rooted recursive purge never follows links.

Linux and Windows support atomic moves. Other Go platforms compile but return a safe unsupported-move
error. Cross-filesystem moves (including separately mounted individual server roots) fail without a
copy/delete fallback; provision the trash sibling on the same filesystem. Linux filesystems must
support `RENAME_NOREPLACE`. Trash is not a storage quota mechanism and does not reclaim bytes until purge.

## Cleanup lifecycle

The manager attempts a 1024-directory-entry sweep on startup, then continues with 128-entry sweeps in
its existing 20-second idle watchdog. Directory cursors ensure large unexpired collections do not
starve later entries. There are no new background goroutines. Manager shutdown closes cursor handles
and prevents reopening them. Sweep errors are logged. Successful server deletion makes a best-effort
immediate trash cleanup; failures are logged and normal expiry remains available.

The sweep bounds **entries scanned**, not elapsed time or bytes: recursive removal of one large expired
directory may take longer. Large backlogs can take multiple watchdog ticks to physically disappear;
restore/list enforce expiry independently. Retention is exactly `deleted_at + 24h`, based on node wall
clock, never on list time. Operators should maintain a synchronized clock. One manager/daemon should
own a given store; in-process operations are serialized, while no-replace publication is OS-atomic.

## Verification

`go test ./internal/trash ./internal/api ./internal/runner` covers default/permanent single and bulk,
list/restore/purge, persistence and exact expiry, missing parent/collision, malicious paths and metadata,
symlinks/hardlinks, cross-server IDs/permissions, real master-to-worker proxy routes, bounded sweep
progress and Windows failed-move metadata rollback. Existing bulk tests remain unchanged.

## Panel UI and deployment

Open **Files → Trash** to restore or permanently purge an entry. The delete dialog defaults to
**Move to Trash**; **Delete permanently instead (skip Trash)** is unchecked. The panel verifies Trash
support on the destination worker before recoverable deletion. Update the panel and workers together.
Restore needs the original parent folder to exist; restore deleted parents first, or recreate them.

Only file-manager single/bulk deletion uses this bin. SFTP deletes, plugin uninstalls, server deletion,
and reinstall cleanup are not redirected into it. Trash remains a recovery convenience, not a backup.

Browser drag/drop moves a selected set with one non-overwriting request; holding a folder/breadcrumb
opens it after 650 ms without losing the original source paths. External file drops use the upload
queue and target folder. External **directory uploads** remain unsupported and are explicitly rejected.

The accompanying bubblewrap fix mounts the host's public OS-release file read-only at `/etc/os-release`
(falling back to `/usr/lib/os-release`), not all of `/etc`. This addresses OSHI's missing-file warning
inside that sandbox. Rebuild/update the node and restart the game server. Live Linux verification is
still required; a genuinely missing/unreadable host release file is a host configuration issue.
