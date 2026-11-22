# NeoPanel Desktop for Windows

## Get started

1. Open **NeoPanel-Desktop.exe**. New installs sign in as Guest automatically.
2. Click **Create server**, choose **Local**, and pick your Minecraft version.
3. Install Java if prompted, review the EULA, then click **Start**.

Minimize the app to keep servers running. To quit, close it and choose to save
and stop your servers. Only run servers and plugins you trust.

Already have an account? Sign in as usual. Want a password for Guest?
Set one in **Admin → Users** to turn off automatic sign-in.

## Start fresh

Close the app, then rename `%LOCALAPPDATA%\NeoPanel-Desktop` to
`NeoPanel-Desktop-backup`. Open the EXE again for a fresh start.
Your old worlds and settings stay in the backup folder. If you chose a custom
data folder, rename that folder instead.

Windows gets occasional updates; Linux remains the main focus. The published
beta may be older than this branch.

<details>
<summary>Advanced: building, storage, and security</summary>


## Build and run

On Windows with Git and Go 1.26+, check out `windows-desktop-release`, then run
these commands from the repository root:

```powershell
powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\build-windows-desktop.ps1
.\dist\windows-desktop-amd64\NeoPanel-Desktop.exe
```

The script builds a GUI-subsystem EXE with an icon, version resources, per-monitor
DPI support and an as-invoker manifest, plus a curated ZIP and SHA-256 checksums.
Use `-Architecture arm64` to cross-build for Windows ARM64. The pinned resource
builder is downloaded through Go modules on the first build. Node/npm and the Wails
CLI are not required: the current interface has no frontend compilation step.

Desktop dependencies live in the nested `desktop/go.mod` module, the application
in `desktop/internal/desktop`, and its entry point in `desktop/cmd/neopanel-desktop`.
The root PowerShell script handles the separate build. To run desktop tests on Windows:

```powershell
Push-Location .\desktop
go test -tags desktop ./...
Pop-Location
```

To run a package you only need Windows 10/11 and Microsoft Edge WebView2 Runtime
(normally already present). Go is not needed. If WebView2 is missing, install it
from [Microsoft's WebView2 page](https://developer.microsoft.com/en-us/microsoft-edge/webview2/);
no runtime is bundled in the ZIP.
Java is needed only for local Java servers. Before first portable use, run the
included `NeoPanel-Runtime.exe java install-all` from PowerShell under the hosting
account. This helper provisions and verifies Java 8, 11, 16, 17, 21 and 25 without
starting a service or game. `java list` reports local availability. New servers use
Auto to match verified JAR/build requirements; explicit runtimes remain unchanged.

### Portable beta and optional installer

This beta ships the portable EXE/ZIP only. Inno Setup was unavailable for this
release, so no setup EXE is included or claimed as tested. The portable build does
not need Inno Setup. For your own optional installer build, install Inno Setup 6:

```powershell
.\build-windows-desktop.ps1 -Installer
# Or specify the compiler:
.\build-windows-desktop.ps1 -Installer -InnoCompiler 'C:\Tools\Inno Setup 6\ISCC.exe'
```

This adds `NeoPanel-Desktop-Setup.exe`. It installs for the current user under
`%LOCALAPPDATA%\Programs\NeoPanel Desktop`, adds a Start menu shortcut, and offers
an optional desktop shortcut. There is no service, scheduled startup, firewall
change or administrator elevation. Close NeoPanel safely before upgrading.
Setup runs the included `NeoPanel-Runtime.exe java install-all` synchronously before
offering application launch. A missing family, failed download, or unsupported
platform makes setup incomplete and suppresses launch. Retry that command in the
installation folder after resolving the error. Old Java families may be unavailable
on Windows ARM64. Use the same account and `NEOPANEL_JAVA_DIR` for setup and hosting.
The helper is a backend binary; run it only with the documented Java subcommands
for this purpose. It is included in package checksums.
Uninstall removes application files, **not** databases, credentials or worlds.
These builds are unsigned; production code signing is a separate release step.

## First start and storage

An empty desktop database gets a local **Guest** administrator and opens straight
to the dashboard, without setup or login. Existing accounts are never converted
or bypassed. Guest has no usable password; its authenticated session is handed
from the native shell to the panel, never exposed in an HTTP endpoint or URL.

Anyone able to open this Windows profile can use Guest, including its connected
nodes. To require login, edit Guest under **Admin → Users** and set a password.
Password changes, security enrollment, role changes, or deleting Guest permanently
disable automatic Guest sign-in. Sign-out lasts until the app restarts; Reload
does not sign you back in. Linux panel authentication is unchanged.

Desktop storage stays separate from legacy installations and worker data:

| Data | Location |
| --- | --- |
| Configuration | `%LOCALAPPDATA%\NeoPanel-Desktop\config.yaml` |
| Database | `%LOCALAPPDATA%\NeoPanel-Desktop\neopanel.db` |
| Local server files | `%LOCALAPPDATA%\NeoPanel-Desktop\servers` |
| Browser profile/sign-in | `%LOCALAPPDATA%\NeoPanel-Desktop\webview` |
| Desktop log | `%LOCALAPPDATA%\NeoPanel-Desktop\desktop.log` |
| Window preferences | `%LOCALAPPDATA%\NeoPanel-Desktop\window.json` |

The toolbar's **Data folder** button opens this folder. A second launch for the same
data directory focuses the existing window rather than starting another backend.

An alternate instance can be launched with an explicit data directory:

```powershell
.\NeoPanel-Desktop.exe -data-dir 'D:\NeoPanel Desktop'
```

The backend is always loopback-only and SFTP is disabled in desktop mode, even if
those settings were changed in YAML. It selects a free HTTP port on first launch
and remembers it. If another application occupies that port later, NeoPanel selects
another; you may need to sign in again because browser storage is origin-specific.
Accounts, server files and node configuration remain in the database.

To migrate an older Windows panel installation, first **stop that panel and all
its local servers**, back up its data/configuration, then explicitly select its
data directory. Never run two processes against the same database or server
directories. The desktop does not import, move or erase existing data automatically.
Use the Linux panel when you need a publicly reachable web interface, SFTP, or a
separately managed background service.

## Hosting and nodes

- Create a server with **Local** selected to host it on this Windows PC.
- Select/install a suitable Java runtime, then install the desired engine. Review
  the applicable Minecraft EULA before running Minecraft.
- Connect your game client to `localhost:<game port>`; HTTP and game ports differ.
- Add remote workers under **Admin → Nodes**, using their reachable API URL and
  node token. Protect worker APIs with TLS or a trusted private/VPN network.
- An optional [API-only Windows worker](https://github.com/Pekly/NeoPanel/blob/windows-desktop-release/docs/windows.md) can host servers on another PC.
  Select Automatic/Host Direct process management for it, not a Linux isolation
  engine. No worker is needed for desktop-local hosting.

## Window controls

The frameless window has a custom toolbar with **Dashboard**, **Reload**, and
**Data folder** actions, plus minimize, maximize/restore, and safe close. Drag the
empty title-bar area to move the window. Reload refreshes the interface, not the
backend or game servers; save editor changes before reloading.

## Closing and background operation

Minimizing the app keeps the backend and local servers running in the taskbar.
There is no tray-only mode in this first desktop version. Closing with active local
servers asks whether to stop/save and quit (**Yes**) or minimize/keep running (**No**).
The window remains responsive while shutting down. Remote workers are not stopped.
Without active local servers, closing exits normally. Window size/maximized state
is saved. Use the toolbar to return to the dashboard or reload the interface.

The existing runner sends each engine's normal stop command and waits before its
bounded force-stop fallback. A crashed/hung server may require force termination;
no application can guarantee a world save after a crash or forced Windows shutdown.
Do not kill the app in Task Manager to stop worlds. Windows sleep/logoff/power loss
can interrupt hosting; this app is not an always-on Windows service.

## Security and compatibility

- Panel operations continue to use the existing authenticated HTTP/WebSocket APIs
  on loopback; the native toolbar does not replace panel authentication.
- Windows local workloads are **trusted-only**. Job Objects manage process trees
  and memory but do not isolate filesystem, account credentials or network access.
- No automatic elevation, firewall opening, node deployment or EULA approval is
  added by the desktop shell.
- Linux panel and API-only worker builds retain the shared `cmd/neopanel` entry
  point. Wails belongs only to `desktop/go.mod`, not the root module. Linux install/update checkouts omit
  desktop code, packaging, Windows scripts, and Windows documentation, while keeping
  shared Windows build-tag code. Linux builds do not install Wails, GTK, or WebKit.
- Read the [security upgrade notes](https://github.com/Pekly/NeoPanel/blob/main/docs/security-upgrade.md)
  before upgrading an existing panel or workers.

</details>
