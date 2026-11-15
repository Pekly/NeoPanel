# NeoPanel Desktop for Windows

The `windows-desktop` branch adds a standalone Windows application using Wails and
Microsoft Edge WebView2. Double-click **NeoPanel-Desktop.exe**: the app opens its own
native window and starts its local Go backend. No console launcher or browser tab
is needed. The familiar panel interface is embedded, not rewritten into WinUI controls.

The same app is the head/master, can host trusted Minecraft servers locally, and
can manage separate web-less Linux or Windows nodes. Players still connect directly
to each game server's host/port; the head does not tunnel game traffic.

## Build and run

On Windows with Go 1.26+:

```powershell
powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\build-windows-desktop.ps1
.\dist\windows-desktop-amd64\NeoPanel-Desktop.exe
```

The script builds a GUI-subsystem EXE with an icon, version resources, per-monitor
DPI support and an as-invoker manifest, plus a curated ZIP and SHA-256 checksums.
Use `-Architecture arm64` to cross-build for Windows ARM64. The pinned resource
builder is downloaded through Go modules on the first build. Node/npm and the Wails
CLI are not required: the current interface has no frontend compilation step.

To run a package you only need Windows 10/11 and Microsoft Edge WebView2 Runtime
(normally already present). Go is not needed. If WebView2 is missing, install it
from Microsoft's official WebView2 download page; no runtime is bundled in the ZIP.
Java is needed only for local Java servers and can be installed through the panel.

### Optional installer

Install Inno Setup 6, then:

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

**NeoPanel → Open app data folder** opens this folder. A second launch for the same
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

## Closing and background operation

Minimizing the app keeps the backend and local servers running in the taskbar.
There is no tray-only mode in this first desktop version. Closing with active local
servers asks whether to stop/save and quit (**Yes**) or minimize/keep running (**No**).
The window remains responsive while shutting down. Remote workers are not stopped.
Without active local servers, closing exits normally. Window size/maximized state
is saved. The View menu can reset window size or reload the interface.

The existing runner sends each engine's normal stop command and waits before its
bounded force-stop fallback. A crashed/hung server may require force termination;
no application can guarantee a world save after a crash or forced Windows shutdown.
Do not kill the app in Task Manager to stop worlds. Windows sleep/logoff/power loss
can interrupt hosting; this app is not an always-on Windows service.

## Security and compatibility

- The native shell exposes no panel-callable Go bindings. The panel continues to
  use its existing authenticated HTTP/WebSocket APIs on loopback.
- Windows local workloads are **trusted-only**. Job Objects manage process trees
  and memory but do not isolate filesystem, account credentials or network access.
- No automatic elevation, firewall opening, node deployment or EULA approval is
  added by the desktop shell.
- The existing `cmd/neopanel` and `-tags node` builds remain available. Desktop-only
  Wails code is Windows-gated; Linux headless builds do not require GTK/WebKit.
- Branch creation carries uncommitted changes onto this branch; it does not create
  a commit or push a release automatically.
