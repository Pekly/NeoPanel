# Idle auto-shutdown

In **Admin → Nodes → Edit**, enable **Idle auto-shutdown** and set **1–1440 minutes** (default 15). For remote nodes, **Save Node** saves connection settings and attempts to apply the idle policy to the worker. For Local, use **Save idle policy**; this does not change addresses, isolation, or user permissions.

## Upgrade and apply

Update and restart the panel **and the affected workers**, then save/retry the policy once. Older versions saved remote policy only on the master, so an old saved checkbox is not proof that the worker received it. The UI now distinguishes saved policy from the last confirmed worker application.

- **Last worker delivery accepted**: the worker persisted the matching policy at its configured URL/token. It is historical acknowledgement, not a live timer readout.
- **Worker application pending**: the master saved the setting, but the worker has not confirmed it. Read the error, update/reconnect the worker, and choose **Retry saved policy**. This resends the last saved policy, not unsaved form edits.
- **Refresh policy status** reloads saved desired state and acknowledgement without resending or discarding edited fields.
- **Unknown after request failure**: delivery may have completed before the connection failed. Refresh before retrying.
- A pending disable is **not** a confirmed disable; the worker can still use its previous policy.
- HTTP 404/405 means the worker endpoint or method is unavailable: verify the worker version, node URL, and authenticated forwarding of **PUT `/api/cluster/idle-policy`**. HTTP 401/403 calls for checking the saved Node Token/admin API key and proxy access rules.

## Timer behavior

The host watchdog checks approximately every **20 seconds**. Once a process reaches online status, time with zero tracked players contributes to the idle interval. A join clears the interval; the last player leaving starts a new one. Repeated empty player-list replies and duplicate disconnect messages do **not** restart it.

Only an enabled policy is enforced. A positive timeout value does not enable a disabled toggle. Existing legacy `is_free` is the enable flag; legacy timeout zero means the 15-minute default. Invalid stored settings fail closed rather than stopping games.

The watchdog reads the host's effective **local** policy and acts only on locally assigned, online process instances. Before stopping, it checks the latest saved policy and inventory, uses the normal per-server operation gate, and rechecks the same process generation, player state, and elapsed interval atomically with the transition to stopping. Busy backups/restores/maintenance, blocked servers, occupied servers, and restarted processes are not stopped by an obsolete scan.

Idle shutdown sends the normal **`stop`** command and waits for graceful exit, using the normal bounded fallback if the process refuses to exit. Failed shutdown is logged separately and is never audited as a successful stop. Auto-restart does not undo an intentional idle stop. A new process starts with a fresh idle interval.

Player tracking uses supported console join/leave/login/list formats, not server-list pings. Positive list counts with missing or incomplete names do not count as empty. Custom plugins, suppressed/rewritten logs, or unsupported output formats can prevent accurate tracking; confirm the player state in the console before relying on this policy. Unresolved occupancy is treated conservatively until a valid empty list confirms it.

## Join-to-start interaction

Join-to-start applies **only after a successful idle auto-shutdown**. Both node policies and the server's wake preference must be enabled. Manual stops, crashes, fresh servers, maintenance/restores, and failed launches do not create eligibility. A successful idle stop records a durable marker; a launch consumes it, while manual Stop/maintenance/restore invalidates it. Status pings never wake a server. Existing offline servers are not retroactively marked on upgrade.

Users can change **Join to start after idle shutdown** in **Server Settings → Startup & resources → Server automation**. The node-enforced idle policy is displayed as a locked checkbox and cannot be toggled off by users. Changing the server wake preference never alters the idle timeout or node enable flag; only node administrators can change those.

## API

Administrator endpoints: `GET` / `PUT /api/admin/nodes/{id}/idle-policy`; use `local` for Local. PUT body:

```json
{"enabled":true,"timeout_minutes":15}
```

Responses contain `{config, applied, error}`. A remote delivery failure returns HTTP 502 with the saved desired configuration and `applied:false`. Generic node create/update still reports a successful node save (200/201), with `idle_policy_applied:false` and `idle_policy_error` if policy delivery failed; the UI retains the modal and offers retry instead of treating the entire node creation as failed. Worker application uses authenticated `PUT /api/cluster/idle-policy` and writes local idle policy only. No node token, address, isolation setting, or permission policy is overwritten by the dedicated idle endpoint.
