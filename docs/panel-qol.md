# Panel quality-of-life controls

These features live in the shared web panel. They do not remove or change Windows worker/hosting support, and do not depend on the desktop shell.

## Workspace and server list

- Opening the panel root restores the last server page when that server is still in your authorized inventory. Explicit links take priority. Visiting **Servers** makes the dashboard your saved landing page.
- Refreshing Files restores the last directory and open text file after checking access. Existing unsaved-draft recovery asks before restoring a draft, including a warning when the server copy changed.
- Pin servers with the star on their card. Pins sort first; **Pinned only** filters the list. Sort by name, status, or location. Search, location, sorting, and pins are saved for your account in this browser. The location selector lists only hosts containing your servers and stays locked when they are all on one host. Layout buttons keep a stable toolbar position.
- The address button in the server header copies the Minecraft player connection address, not the panel URL. Its label stays consistent during live updates. Card addresses remain copyable.

## Files and plugin browsing

- Trash is the first navigation item in the root Files listing, outside normal file selection and bulk operations.
- Open **Files → World tools → Convert world** for world conversion and existing-job monitoring.
- Plugin results open an in-panel project page; **Load more** extends the search list. Default Install selects the newest release for the actual server software without requiring an exact Minecraft-version tag. Explicit release selection remains available. See the [plugin browser guide](plugin-browser.md) for compatibility caveats and details.

## Text editor

- **Save** is unavailable when there are no changes, a save is pending, or the file is read-only. Ctrl+S also avoids unchanged writes.
- The **Saved** label shows the file's acknowledged modification time, with the full date in its tooltip. Opening the editor does not invent a new save time; failed saves do not advance it.
- Typing during a save remains unsaved work after that request completes; only the submitted contents are acknowledged.
- **Ctrl+F / Cmd+F** opens search inside the active editor. Outside it, normal browser find remains available.

## Console and connection status

- Scroll upward to stop following new output. The position survives UI redraws; **Jump to latest** resumes following.
- Open **Console tools ▾** in the console toolbar for output filtering, **Jump to latest**, and **Clear command history**. These controls stay collapsed until needed; an active filter is indicated on the button. Filtering does not delete output.
- Up/Down recall the last 100 commands for the selected server.
- Command history and console filters are account/server/node scoped and kept in the browser tab's session storage. Closing the tab clears that storage. File draft contents also remain tab-local; regular workspace preferences use local storage.
- A disconnected/reconnecting banner pauses power and console-command controls. **Reconnect & refresh** retrieves server status; reconnecting the console also refreshes it. A lost request is never automatically replayed.
- Errors include relevant next steps for connection problems, access restrictions, operation conflicts, missing resources, Java, and disk space.

## Maintenance access

Maintenance shows a notice but still lets users sign in, select their server, use its power controls, and open its console. Account access remains available. Other management pages show the maintenance notice; administrators retain full access.

Existing permissions continue to apply:

- `server.power`: start, graceful stop, restart, and existing force-stop control.
- `console.view`: console output.
- `console.command`: console commands.

A power-only user can open the console page's power controls without gaining access to console output. Maintenance does not grant permissions or create public/anonymous server controls.

## Warn players before stopping or restarting

On Console, open **Schedule ▾** beside the power buttons. Choose the action, enter a countdown of **10–3600 seconds**, and optionally add a single-line player message. An active countdown is shown on the button; reopen the dropdown to cancel it. Escape or clicking outside closes the dropdown.

Scheduling requires both `server.power` and `console.command`. The server must be online and support the `say` command. The backend writes the initial warning and further warnings at applicable 60-, 30-, and 10-second milestones. Successful command delivery means it reached server input, not that a plugin or proxy displayed it.

The visible countdown has a **Cancel countdown** action for users with power permission. Another direct power action is refused while a countdown is pending. Cancellation cannot interrupt a shutdown/restart already executing.

Countdowns continue if you close the browser, but are cancelled if the hosting panel/worker process restarts, the original game process stops or changes, the server is deleted, or the originating user's required permissions are no longer valid on that host. Failed final actions are logged and are not automatically retried.

Update the panel and remote workers together to use countdowns. Remote workers must have synchronized originating accounts and permissions. Master-side permission changes are effective on a worker once synchronized.

If countdown status cannot load, the panel shows the HTTP error or compatibility problem and offers **Retry countdown status**. Unsupported endpoints and denied access pause automatic polling; transient failures back off. Countdown scheduling stays disabled until a valid status response returns. An older backend lacking countdowns does not by itself disable ordinary power controls. Updating frontend files alone is not enough: rebuild/redeploy and restart the hosting backend as well.

## Panel hostname

Use **Admin → Appearance → Public URL** to change the public panel FQDN. Check DNS, trusted HTTPS, same-panel identity, and passkey compatibility before applying the verified address. Supported installer-managed Linux/Caddy setups can prepare a Let's Encrypt certificate while retaining the old hostname. Email and administrator passkey settings link to the same workflow. The [panel address guide](panel-address.md) covers setup, migration, and recovery. Game and SFTP addresses are not changed.

## Admin setup and appearance

- **Getting started** is available only to administrators during the first three days after their account was created, never to ordinary users.
- Under **Admin → Nodes**, choose **Rename** for any node, including the local host. This changes only its display name, not its addresses, credentials, policies, or running servers.
- Local node cards distinguish the saved **Panel URL** from the **Game / SFTP host**. A panel-domain change updates the panel label without silently moving game connections. Use the local node's **Edit address** button to open Appearance → Server defaults and change the advertised game/SFTP host. Successful saves refresh cached local server addresses; **Refresh Nodes** reloads both the node details and saved panel URL.
- The saved panel palette and your selected light/dark/OLED/automatic appearance are applied before the page's styles render. The server supplies the current palette for new browsers; a validated browser cache is the fallback when that response is unavailable.

## Server state after panel updates and restarts

After this version is installed, a clean panel or worker shutdown saves which locally managed servers were online. The next successful startup loads that one-shot snapshot and starts those servers again through the normal process and isolation checks. Servers that were offline, deliberately stopped, stopping, or startup-halted stay off. This is independent of the per-server crash auto-restart option.

- Use the normal updater or a graceful service stop/restart. On Windows, stop the hosting process with Ctrl+C and wait for cleanup before replacing the executable; do not force-kill it.
- A master-panel update does not restart remote workers or their games. Each worker preserves its own local running state when that worker is updated/restarted.
- Servers with a conflicting backup/restore/reinstall/deletion operation may be skipped conservatively. Finish those operations before updating. Missing servers and failed launches are logged in the daemon output/activity log and are not replayed on every boot; resolve the error and start them manually.
- Pending player-warning countdowns are cancelled, not resumed. A server restored from the previous snapshot may still be starting if another shutdown arrives; that pending recovery is carried forward.
- The snapshot is stored in the existing SQLite database and bound to the server data directory. Forced termination, power loss, corrupt/unwritable storage, and moving to a different data path are not guaranteed recovery paths.

**First-upgrade limitation:** the old running binary must support saving the snapshot. When first upgrading from an older version, note which servers are online and start them manually once after the update. Subsequent clean updates/restarts restore their running state automatically. Back up worlds and configuration before updating.
