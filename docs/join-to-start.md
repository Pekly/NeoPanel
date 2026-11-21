# Minecraft Java join to start

Join to start is **off by default**, separately for Local and every connected remote node. In **Admin → Nodes**, choose **Edit** on the target node, then **Configure join to start**. This is available for Local and remote nodes in both list and card layouts. **Back** returns to the node settings without discarding unsaved node edits; wake settings are saved separately. Enabling it opts in that node’s **entire local Minecraft Java inventory**, not an individual server or every node in the cluster. Bedrock is not supported.

## Requirements and setup

- Update and restart the panel **and each participating worker** to a release supporting join to start. Updating only the master cannot add listeners to an older worker.
- The node daemon must remain running while game servers are stopped. Java and the server’s normal start configuration must already work; this feature does not install or repair them.
- Players must be able to reach the server’s normal Java TCP port on the node. Configure firewalls/NAT as usual. The selected bind IP must exist on the node; use an empty address for all interfaces. A hostname, URL, or address with a port is not a bind address.
- A stopped server’s port must be available to the node daemon. Resolve duplicate/conflicting ports and other processes occupying them. The real server must be able to bind its normal port after handoff.
- Only locally assigned, recognized Java-protocol server types with primary ports 1024–65535 are eligible. Startup-halted, restore-blocked, active, conflicting, and busy servers are skipped. Listeners do not count as running servers for restart recovery.
- Configuration and inventory are reconciled roughly once per second. There are at most 1024 offline listeners, 64 concurrent wake connections, and 12 admitted connections per source IP per 10 seconds. Protocol connections have a five-second deadline; excessive traffic is dropped. Shared/NAT addresses share that connection budget.

Open the settings, wait for the saved configuration to load, edit it, then select **Save settings**. There is no autosave. The GET response supplies authoritative defaults; the browser does not invent defaults when loading fails.
The separate dialog groups controls into **Startup behavior**, **Offline appearance**, and **Player messages**. Desktop sections sit side by side; small screens stack them. Status cards distinguish saved settings, worker acknowledgement, and live listener readiness. **Check listeners** reads runtime state without starting a game; it is not an external connectivity or game-startup test.

## What players experience

1. While a Java server is stopped, the node can answer the normal server-list status ping with its configured offline MOTD, version label, maximum-player display, and icon. **Status pings never request a start.** Refreshing the multiplayer list is safe.
2. A Java **login attempt** (joining the server) can request a start, subject to runtime eligibility and the cooldown. The player receives a disconnect message telling them to reconnect, or the unavailable message when a start cannot be requested.
3. The wake listener releases the game port for the real server. The player must reconnect after boot completes. An early reconnect may fail or receive connection refused during the handoff/boot window.
4. Once online, the real server handles connections and its own MOTD/icon. These settings do not modify its online presentation.

The sleeping listener advertises protocol **-1** so vanilla Java renders its **red-X incompatible ping sprite**, rather than healthy game-server ping bars, alongside the configured version label. The default label says **Sleeping • Join to start**, and real login requests are still accepted. Minecraft controls the sprite and tooltip; modded clients/resource packs may display it differently. This is not the server's 64×64 PNG icon. Once the real game takes over, its normal version/ping behavior returns.

This is **not a proxy**, does not hold or transfer a player connection, and does not display Minecraft 1.21.6 dialogs. There is no separate “starting MOTD” setting or guarantee of status replies while the real server is starting. The starting setting below is a **disconnect message**, not a boot-time status listener.

## Settings and current defaults

Limits on text are **UTF-8 bytes**. Text supports `{server}`, `{version}`, newlines, legacy `&`/`§` colors and styles, and exact hex colors such as `&#387aff`. MOTD and disconnect formatting is encoded as real Minecraft JSON text components; pre-1.16 clients receive the closest legacy color for hex values. The admin form keeps editable formatting codes; text is never treated as user-supplied JSON or HTML.

| API field | Current default | Meaning / validation |
| --- | --- | --- |
| `enabled` | `false` | Node-wide Java opt-in. |
| `bind_address` | `""` | Empty = all interfaces; otherwise a literal IPv4 or IPv6 address, without brackets or a port. |
| `offline_motd` | Blue/bold **NeoPanel » {server}**, then green/bold **JOIN TO START • Your adventure is one click away** | Colored two-line offline status text, at most 2048 bytes. |
| `version_label` | `&cSleeping &8• &aJoin to start` | Offline red/green label, at most 128 bytes; does not change the installed server version. |
| `max_players` | `20` | Offline display only, integer 0–100000; does not change the real server’s player limit. |
| `icon_data_url` | `""` | Optional custom 64×64 PNG; `data:image/png;base64,...`, at most 65536 bytes including prefix/base64. Empty removes the custom icon. |
| `use_server_icon` | `true` | Use a valid server `server-icon.png` when available. |
| `starting_message` | Branded header, green/bold **WAKE REQUEST ACCEPTED**, and yellow reconnect guidance | Colored multiline login disconnect text, at most 4096 bytes. |
| `unavailable_message` | Branded header, gold/bold **NOT READY YET**, and yellow retry guidance | Colored multiline unavailable text, at most 4096 bytes. |
| `cooldown_seconds` | `30` | Start request cooldown, integer 10–600 seconds; not a promised boot duration. |

Use the PNG file picker to preview a custom icon or **Remove custom icon** to clear it; save to persist the change. The data-URL limit includes base64 overhead, so a 64 KiB PNG file is too large. Server-side validation checks the actual PNG and dimensions. The server-icon option and custom offline icon do not overwrite the server’s files or affect its online icon.
The server icon is read when its offline listener opens; reopen the listener (for example, save a changed wake setting or stop the server after a run) to reload a changed file. A valid custom icon is the fallback when the server icon is missing or invalid. Expanded disconnect text is also subject to a bounded wire-packet limit; oversized output uses a safe reconnect fallback.

**Use panel branding template** reads the panel's saved brand name and accent and replaces only the four text fields in your draft. It never changes enabling, bind address, cooldown, player count, or icons, and never saves automatically. Review and save to apply. Existing saved custom messages are preserved on upgrade; new configurations get the colorful NeoPanel defaults. Branding changes do not silently rewrite worker settings—apply the template again explicitly.

## Live listener indicator

On opening the dialog, after saving/retrying/refreshing, or with **Check listeners**, the panel requests a fresh runtime report. **Ready — N listening** is green only for a recent matching configuration with at least one listener waiting for a join. Cooldown/busy, disabled, blocked, bind-failed, conflicting-port, and standby states remain distinct. Expand **Per-server listener status** for the reason on each local server. Readiness expires after ten seconds rather than displaying a stale success indefinitely.

The authenticated endpoints are `GET /api/admin/nodes/{id}/wake/status` and worker `GET /api/cluster/wake/status`. They report actual reconciliation time, configuration hash/match, bound socket count, and up to 1024 local server states. Reading status never binds ports or sends a player join. Remote reports must be recent, valid, and match the master's current desired configuration; an old/unreachable worker shows unavailable, not ready. A bound port still does not prove firewall/NAT reachability, whitelist access, or a successful game launch. Actual Java joins remain the end-to-end check.

## Saved desired settings versus worker acceptance

The administrator API is `GET /api/admin/nodes/{id}/wake` and `PUT /api/admin/nodes/{id}/wake`, with `local` for Local and the actual remote node ID otherwise. Both return an envelope:

```json
{
  "config": { "enabled": false, "...": "the complete settings object" },
  "applied": false,
  "error": "worker delivery/application error, or an empty string"
}
```

The abbreviated `config` above is illustrative, not a valid PUT body. **PUT sends the complete configuration object only**, with the ten fields in the table; do not send the envelope or `applied`/`error` fields.

- **Saved desired settings** describe what the master wants the node to use. Expand that section in the modal to inspect the last confirmed saved configuration independently of unsaved form edits.
- **`applied: true` means configuration was accepted by the node runtime**, not that every game port is listening or every server can start. Bind failures are logged by the runtime; inspect the node logs and test actual connectivity.
  Remote acceptance is the last matching acknowledgement, not a live probe of worker state. Local acceptance means the configuration was persisted; the listener coordinator applies it on its next reconciliation. Manually changing the worker independently can make the master's last acknowledgement stale; explicitly resend the desired settings to restore agreement.
- **`applied: false` means delivery/application is pending or failed.** Read the visible error, restore node connectivity or update the worker, and use **Retry saved settings** to explicitly resend the last confirmed desired configuration. Older worker settings may still be active.
- A **pending disable is not confirmation that wake is disabled**. Do not begin maintenance on that assumption. Verify acceptance on the worker; if it is unreachable, use host/network controls to prevent unwanted starts until it is reachable again.
- **Refresh status** reloads saved desired settings, acknowledgement, and live listener state without discarding edits. It does not retry delivery. **Check listeners** reads runtime state only. **Save settings** submits current edits; **Retry saved settings** resends the last saved configuration, leaving edits untouched.
- Request failures preserve edited fields. If no valid envelope was received, the UI reports unknown status: a network failure can occur after a save/delivery completed. Refresh before deciding what to resend. Closing and reopening the modal loads saved settings, not the abandoned draft.

### HTTP 405 / 404 from a worker

The panel sends **PUT `/api/cluster/wake`** to the node URL. A 405 means the responding server or proxy does not allow that method; a 404 means that route is unavailable. This can happen when the worker is still running an older binary, the node URL points to the wrong service, or a reverse proxy is not forwarding the route/method. The response alone does not identify which cause applies.

1. Update **the affected worker**, not just the master panel, to a join-to-start capable release and restart its daemon so it runs the new binary.
2. In the node's Edit dialog, check that its saved URL points to the intended NeoPanel daemon (not a game port, static website, or obsolete URL).
3. If a proxy fronts that daemon, verify it forwards `PUT /api/cluster/wake` with the normal authentication header. Do not remove authentication or expose the node token to diagnose this.
4. Open join-to-start and choose **Retry saved settings**. Wait for acceptance, then use **Check listeners** for runtime readiness. **Refresh status** reloads delivery/runtime information but never retries delivery. Proxies must also forward authenticated `GET /api/cluster/wake/status` for the new readiness indicator.

HTTP 401/403 instead calls for checking the saved Node Token / administrator API key and proxy access rules. Redirects are intentionally not followed to protect credentials. Pending enable/disable settings are not proof of what is currently active on the worker.

## Security and maintenance

The offline listener operates **before Minecraft authentication**. An unauthenticated login attempt, including one from an automated client, can request a start. A supplied player name is **not proof of identity**, and join to start is **not authorization**. The real server must still enforce its normal authentication, whitelist, bans, and permissions after reconnect; those checks do not protect the wake request itself.

Only enable this behavior where anyone able to reach the port may request consumption of server resources. Restrict network access if that is not acceptable. Cooldowns limit repeated requests but are not an access-control boundary. Disable join to start on the node for maintenance and confirm acceptance before stopping servers; otherwise a login attempt may start them again.

For problems, check desired versus accepted state first, then worker version/connectivity, runtime bind-error logs, port conflicts, bind IP, firewall/NAT, and the server’s normal Java/start logs. Test a status ping separately from a login attempt, and allow boot to finish before reconnecting.
