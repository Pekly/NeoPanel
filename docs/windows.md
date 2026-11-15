# Windows desktop and optional worker

**NeoPanel Desktop is the Windows panel edition.** Double-click
`NeoPanel-Desktop.exe`; it opens the interface and starts its own local backend.
No console launcher or separate browser-panel package is needed. See the
[desktop guide](https://github.com/Pekly/NeoPanel/blob/windows-desktop-release/docs/windows-desktop.md)
for downloads, building, storage, and safe shutdown. Use the Linux panel when you
need a publicly reachable web interface or a managed background service.

## Optional API-only Windows worker

A worker is not another panel edition. It runs trusted game servers on a separate
Windows PC and is controlled from the desktop app or a Linux panel. It serves
control APIs only, has no panel UI, and cannot be switched into panel mode.
You do not need it to host servers locally in the desktop app.

### Build

With Go 1.26+ on the Windows worker machine, run from the repository root:

```powershell
New-Item -ItemType Directory -Force .\dist\windows-node | Out-Null
go build -trimpath -tags node -o .\dist\windows-node\NeoPanel-node.exe .\cmd\neopanel
```

Go is needed to build, not to run the resulting EXE. Local builds are unsigned.
Windows receives occasional updates; Linux remains the primary hosting platform.

### Configure and run

```powershell
cd .\dist\windows-node
.\NeoPanel-node.exe -init
.\NeoPanel-node.exe node-token --raw
.\NeoPanel-node.exe -host 127.0.0.1 -port 8081
```

Keep the node token private. Default storage is `%LOCALAPPDATA%\NeoPanel-node`,
separate from `%LOCALAPPDATA%\NeoPanel-Desktop`. The last command runs the worker
in the foreground. Use Ctrl+C for graceful shutdown rather than killing it.

For a head on another machine, replace `127.0.0.1` with the worker's LAN/VPN
interface address and deliberately restrict its firewall to the head. Use HTTPS
or a trusted private/VPN network; do not expose an unencrypted node token publicly.
No firewall rules, elevation, or Windows service installation happen automatically.

In the head's **Admin → Nodes**, add the worker's reachable API URL and token.
`localhost` always means the machine making the connection. Choose **Automatic**
or **Host Direct** for Windows process management; Linux isolation engines are
not supported on Windows. Players connect directly to the game host/port, not
the head's HTTP address.

## Hosting limits and data safety

- Windows workloads are **trusted-only**. Job Objects own process trees and limit
  memory, but do not isolate files, credentials, accounts, or network access.
  Host untrusted workloads on properly isolated Linux workers.
- CPU quotas and task-count limits are not enforced on Windows. Per-process
  network accounting is unavailable; CPU and RAM reporting use host APIs.
- Java is required for Java servers. Choose a compatible runtime through the head
  and review the Minecraft EULA before starting Minecraft.
- Stop local servers and the worker before upgrading or moving data. Back up the
  database, configuration, worlds, credentials, and sibling `servers-trash` folder.
  Windows does not use the Linux auto-updater: stop, replace the EXE, then restart.
- Force-closing a process kills its owned jobs; this is crash cleanup, not a
  guaranteed world save. Windows sleep, logoff, and power loss interrupt hosting.
- Windows ARM64 cross-builds do not replace testing on ARM64 hardware. Local
  Windows tests do not establish hostile-tenant isolation or Linux compatibility.
