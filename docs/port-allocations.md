# Port allocations

**Server → Network** shows **Used / Allowed**, including the primary port. Administrators set the allowance there. New servers default to one total port; upgrading preserves existing allocations and allows at least the number already assigned. Existing legacy collisions are preserved for administrator review, not silently reassigned.

## Member allocation

1. An administrator sets a random port pool for each hosting node in **Admin → Appearance → Member port pool**.
2. The node must permit member port changes, and the member needs `settings.edit` (owners already have it).
3. Members choose protocol and notes, not a port number. The API selects a random free port from that node's pool, within the server allowance. An unset/exhausted pool refuses allocation. Submitted member port numbers do not override random selection.
4. Administrators can choose explicit port numbers, but must still respect the server allowance and collision checks. Increase the allowance first if necessary.

A number counts once, whether its protocol is TCP, UDP, or both. New reservations exclude other servers' primary and additional ports on the same hosting node. Concurrent allocations are checked in SQLite write transactions. This does not reserve a socket against non-panel software or prevent untrusted plugins from binding other ports.

## Does assigning a port apply it?

- **Additional port:** reserves the number in the panel. Configure the plugin or service to listen on that number. Allocation itself does not create a listener.
- **Primary port:** the server must be stopped. NeoPanel updates `server.properties`, the Steel configuration, or the relevant proxy configuration together with the allocation change. It applies on the next manual start; the panel does not restart the server automatically.
- A newly added primary retains the previous allocation as an extra port, consuming another allowance slot.
- The primary-port field in general settings is read-only. Generic server PUT cannot bypass allocation rules or edit the allowance; use the dedicated Network endpoints.

Configuration failure rolls back the database change and restores the previous file contents. If file rollback itself fails, further operations remain blocked in that daemon until an administrator repairs the configuration. Filesystem publication and SQLite are not one crash-durable transaction.

Firewall helpers remain best-effort. Allocation does not prove that NAT, host firewall, DNS, or an external connection is working. No live firewall/network reachability checks are performed by allocation tests.

## API

- `GET /api/servers/{id}/ports` remains an array.
- `GET /api/servers/{id}/ports/summary` returns `{used,allowed}`.
- Admin `PUT /api/servers/{id}/ports/allowance` accepts `{allowed:3}`.
- Admin `GET/PUT /api/admin/port-pool` uses `{start:30000,end:30100}` on the hosting daemon. Use the normal node proxy for remote workers.
- `POST /api/servers/{id}/ports` accepts `protocol`, `notes`, `is_primary`, and an administrator's optional explicit `port`.

Ownership transfer remains administrator-only in server Settings and the API. Port permissions do not grant ownership-transfer rights.
