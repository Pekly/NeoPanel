# Panel quality-of-life controls

These features live in the shared web panel. They do not remove or change Windows worker/hosting support, and do not depend on the desktop shell.

## Workspace and server list

- Opening the panel root restores the last server page when that server is still in your authorized inventory. Explicit links take priority. Visiting **Servers** makes the dashboard your saved landing page.
- Refreshing Files restores the last directory and open text file after checking access. Existing unsaved-draft recovery asks before restoring a draft, including a warning when the server copy changed.
- Pin servers with the star on their card. Pins sort first; **Pinned only** filters the list. Sort by name, status, or location. Search, location, sorting, and pins are saved for your account in this browser.
- **Join** in the server header copies the Minecraft player connection address, not the panel URL. Card addresses remain copyable.

## Console and connection status

- Scroll upward to stop following new output. The position survives UI redraws; **Jump to latest** resumes following.
- **Filter output** filters buffered console text. Filtering does not delete output.
- Up/Down recall the last 100 commands for the selected server. **Clear command history** clears that server's history.
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

On Console, expand **Plan a safe restart or stop**. Choose the action, enter a countdown of **10–3600 seconds**, and optionally add a single-line player message.

Scheduling requires both `server.power` and `console.command`. The server must be online and support the `say` command. The backend writes the initial warning and further warnings at applicable 60-, 30-, and 10-second milestones. Successful command delivery means it reached server input, not that a plugin or proxy displayed it.

The visible countdown has a **Cancel countdown** action for users with power permission. Another direct power action is refused while a countdown is pending. Cancellation cannot interrupt a shutdown/restart already executing.

Countdowns continue if you close the browser, but are cancelled if the hosting panel/worker process restarts, the original game process stops or changes, the server is deleted, or the originating user's required permissions are no longer valid on that host. Failed final actions are logged and are not automatically retried.

Update the panel and remote workers together to use countdowns. Remote workers must have synchronized originating accounts and permissions. Master-side permission changes are effective on a worker once synchronized.

If countdown status cannot load, the panel shows the HTTP error or compatibility problem and offers **Retry countdown status**. Unsupported endpoints and denied access pause automatic polling; transient failures back off. Countdown scheduling stays disabled until a valid status response returns. An older backend lacking countdowns does not by itself disable ordinary power controls. Updating frontend files alone is not enough: rebuild/redeploy and restart the hosting backend as well.
