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
.\NeoPanel-node.exe java install-all
.\NeoPanel-node.exe -init
.\NeoPanel-node.exe node-token --raw
.\NeoPanel-node.exe -host 127.0.0.1 -port 8081
```

Keep the node token private. Default storage is `%LOCALAPPDATA%\NeoPanel-node`,
separate from `%LOCALAPPDATA%\NeoPanel-Desktop`. The last command runs the worker
in the foreground. Use Ctrl+C for graceful shutdown rather than killing it.

Provision Java before starting the worker. `java install-all` verifies all supported
families (8, 11, 16, 17, 21, 25) and exits nonzero if any cannot be installed.
`java list` reports local availability without downloads. Use the same hosting
account and `NEOPANEL_JAVA_DIR`, if configured, for provisioning and hosting.
Alternatively, run `scripts\install-java.ps1 -BinaryPath <existing-backend.exe>`
from the repository. Neither command nor helper starts a service or game server.
Older runtime availability can differ on Windows ARM64; failures are not ignored.

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
- Java servers default to **Auto**, using verified JAR/build requirements rather
  than always Java 25. Manual runtimes remain unchanged. Unknown JARs may require
  an explicit choice; see [Java and custom-JAR identity](new-server-defaults.md).
  Review the Minecraft EULA before starting Minecraft.
- Back up the database, configuration, worlds, credentials, and sibling
  `servers-trash` folder before upgrading. Gracefully stop the worker with Ctrl+C
  and wait for cleanup, replace the EXE, then restart. With restart recovery
  installed, previously online local servers resume; manually stopped servers
  stay off. The first upgrade from an older binary may require manual starts.
  Stop local servers explicitly before moving data if they should remain offline.
  Windows does not use the Linux auto-updater.
- Force-closing a process kills its owned jobs; this is crash cleanup, not a
  guaranteed world save. Windows sleep, logoff, and power loss interrupt hosting.
- Windows ARM64 cross-builds do not replace testing on ARM64 hardware. Local
  Windows tests do not establish hostile-tenant isolation or Linux compatibility.
