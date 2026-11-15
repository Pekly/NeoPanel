# NeoPanel Desktop for Windows

The `windows-desktop-release` branch contains the Windows-only prerelease
`v1.0.0-windows-beta.1`; `main` and the Linux release path remain unchanged.
Double-click **NeoPanel-Desktop.exe** to open the app's own window and local Go
backend. No console launcher or browser tab is needed. The familiar panel interface
is embedded using Wails and Microsoft Edge WebView2.

The same app is the head/master, can host trusted Minecraft servers locally, and
can manage separate web-less Linux or Windows nodes. Players still connect directly
to each game server's host/port; the head does not tunnel game traffic.

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
Java is needed only for local Java servers and can be installed through the panel.

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
Uninstall removes application files, **not** databases, credentials or worlds.
These builds are unsigned; production code signing is a separate release step.

## First start and storage

Create the first administrator inside the app. Desktop defaults are intentionally
separate from the console panel so both cannot accidentally use the same database:

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

To reuse an existing console-panel installation, first **stop that panel and all
its local servers**, back up its data/configuration, then explicitly select its
data directory. Never operate desktop and CLI processes against the same database
or server directories. The desktop does not import, move or erase existing data
automatically. Keep using the normal CLI panel when you need a publicly reachable
browser-based head, SFTP, or a separately managed background service.

## Hosting and nodes

- Create a server with **Local** selected to host it on this Windows PC.
- Select/install a suitable Java runtime, then install the desired engine. Review
  the applicable Minecraft EULA before running Minecraft.
- Connect your game client to `localhost:<game port>`; HTTP and game ports differ.
- Add remote workers under **Admin → Nodes**, using their reachable API URL and
  node token. Protect worker APIs with TLS or a trusted private/VPN network.
- For Windows workers select Automatic/Host Direct process management, not a Linux
  isolation engine. Linux workers keep their existing isolation mechanisms.

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
- The existing `cmd/neopanel` and `-tags node` builds remain available. Wails belongs
  only to `desktop/go.mod`, not the root module. Linux install/update checkouts omit
  desktop code, packaging, Windows scripts, and Windows documentation, while keeping
  shared Windows build-tag code. Linux builds do not install Wails, GTK, or WebKit.
- Read the [security upgrade notes](https://github.com/Pekly/NeoPanel/blob/main/docs/security-upgrade.md)
  before upgrading an existing panel or workers.
