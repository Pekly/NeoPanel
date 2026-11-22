# Optional features

Administrators manage tools in **Admin → Appearance → Optional features**. Each server also has **Settings → Administration → Optional features**.

All tools are enabled by default on upgrade. A server can inherit, enable, or disable a tool; a **panel-wide off switch always wins**. These are administrator-only settings. Owners and team members can read effective capabilities, but cannot change policy.

The switches cover new backups, schedules, SFTP access, SFTP import, plugins/mods, database tools, player tools, the proxy builder, additional port allocations, software/reinstall actions, subdomains, power countdowns, and Chunker world conversion. They control panel tools, not arbitrary commands in uploaded server software. Authentication, permissions, console, basic files, and direct power remain core functionality. Existing node idle/wake settings remain independently configurable.

## Disabling a tool

- The UI hides the relevant tools; APIs also check the saved policy, including administrator requests.
- Disabling new backups retains listing, download, deletion and restore. It does not disable mandatory safety-backup creation for recovery or explicit world replacement; those safeguards cannot be opted out.
- Schedules remain saved. New runs and subsequent steps are blocked. An already executing operation may finish safely.
- SFTP policy is checked on authentication and each authorized filesystem operation, including open handles.
- Active imports/conversions may finish; status, cancellation and results remain accessible. New conversion/replacement/import requests are blocked.
- Existing ports, files, plugins, databases and DNS records are not deleted. Database connections and already-running game plugins are not stopped by hiding their panel tools.
- Pending power countdowns recheck policy and cannot execute after disabling countdowns.

## Workers

Global settings are delivered to each registered hosting worker and persisted there. The admin card reports pending nodes; **Save features** retries delivery, even for unchanged settings. Newly proxied requests synchronize changed policy before proceeding. A pending synchronization blocks those requests rather than falling back to permissive settings.

An unreachable/offline worker retains its last applied policy, including its local SFTP and background tasks. Do not treat a pending switch as applied to that worker. Restore connectivity and retry; update older workers that lack the policy endpoint. Direct administrator changes on a worker can differ from the panel until the next panel save. Per-server overrides are saved on the hosting worker.

These switches are not a substitute for permissions, filesystem isolation, database/network credentials, or firewall policy.

## API

- `GET /api/features` — global effective capabilities (authenticated).
- `GET/PUT /api/admin/features` — admin global policy.
- `GET/PUT /api/servers/{id}/features` — member read/admin write.
- `PUT /api/cluster/features` — authenticated administrator/node delivery.

PUT replaces a sparse object of known feature keys and boolean values. Missing keys inherit enabled defaults; omit a server key to restore inheritance. Responses contain `policy`, `global`, `effective`, and administrator-only node acknowledgement state. Invalid settings fail closed at enforcement points.
