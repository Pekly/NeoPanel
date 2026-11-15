# Windows master panel with local servers and web-less nodes

NeoPanel can run as the **head/master on Windows**, serve the existing browser UI,
run trusted Minecraft workloads on that Windows PC, and manage separate worker
nodes. A desktop/WebView wrapper is not required.

```text
Browser -> Windows NeoPanel.exe (master + local daemon)
                     |-- Local Windows Minecraft server(s)
                     |-- Linux web-less worker -> Minecraft server(s)
                     `-- Windows web-less worker -> trusted local workloads
```

The panel does not tunnel Minecraft traffic: players connect directly to the
machine and port hosting their game server.

## Build/package

Install Go 1.26+ to build from source (Go is not required to run a packaged EXE).
In PowerShell, from the repository:

```powershell
powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\build-windows.ps1
```

This produces `dist\windows-amd64\` and a ZIP/checksum beside it, containing:

- `NeoPanel.exe`: master/browser UI plus local daemon.
- `NeoPanel-node.exe`: node-only build; no embedded browser UI, and node role enforced.
- `start-windows.ps1` / `start-windows.cmd`.
- These instructions and `SHA256SUMS.txt`.

Use `-Architecture arm64` for an ARM64 build. Architecture-specific Java/runtime
availability is separate from whether the panel itself compiles. Local builds are
unsigned; signing and a Windows installer/service are not provided by this script.
The package includes only these build files, never configuration or server data.

## Start the Windows master

Extract the package, then double-click `start-windows.cmd`, or run:

```powershell
.\start-windows.ps1
```

Open **http://127.0.0.1:8080**, create the first administrator, and keep the console
running. Defaults are intentionally local-only:

| Setting | Windows panel default |
| --- | --- |
| HTTP binding | `127.0.0.1:8080` |
| Configuration | `%LOCALAPPDATA%\NeoPanel\config.yaml` |
| SQLite database | `%LOCALAPPDATA%\NeoPanel\neopanel.db` |
| Server directories | `%LOCALAPPDATA%\NeoPanel\servers` |
| SFTP listener | Disabled until explicitly enabled in config |

Use **Ctrl+C** for graceful shutdown and wait for the completion message. Killing
the process, closing the console, logging off, or powering off can interrupt world
saves. Disable sleep if this PC must keep acting as the head. This launcher is a
foreground process, **not** a Windows service. Unattended startup/service management
must be arranged separately under an appropriate dedicated Windows account.

Custom instance or port:

```powershell
.\start-windows.ps1 -DataDirectory 'D:\NeoPanelData' -Port 8085
```

`-DataDirectory` chooses the initial/default configuration location. Existing
explicit YAML paths are preserved. `-ConfigPath 'D:\settings\panel.yaml'` selects
another file. Relative Windows YAML paths resolve against that config file's
folder, not the shell's working directory. Use different data/config directories
for separate panel/node processes; never point both at the same database.

To create/validate configuration without running servers, or persist explicit
host/port choices before startup:

```powershell
.\start-windows.ps1 -InitializeOnly -Port 8085
```

The launcher does not overwrite unspecified host/port settings. Normal `-Port` and
`-ListenHost` flags are runtime overrides; edit YAML or use `-InitializeOnly` to
persist them. `NEOPANEL_DATA_DIR` remains available for controlled deployments.

## Host a Minecraft server directly on Windows

1. Open the master, create a server, and select **Local** as its node/location.
2. Install/select the Java major recommended for the Minecraft version. The Java
   manager discovers supported system installs and managed per-user runtimes.
   A full custom path such as `C:\Program Files\Eclipse Adoptium\jdk-21...\bin\java.exe`
   is also supported. The head needs Java only when hosting Java servers locally;
   remote Linux nodes use their own Java installation.
3. Download/install the chosen engine through the panel. Review Minecraft's EULA
   and the engine's requirements before running it.
4. Start it, then connect Minecraft to `localhost:25565` (or its allocated port).
   Panel HTTP port 8080 and Minecraft game port 25565 are different services.

The panel does not automatically create Windows Firewall rules. Minecraft's own
bind address also determines whether a server accepts only localhost or other
interfaces. Only allow the game port on the networks you intend to serve.

## Add a web-less Linux node to the Windows head

Install the existing node daemon on the Linux machine, or build with:

```text
go build -tags node -o neopanel-node ./cmd/neopanel
./neopanel-node -config ./node.yaml -host 0.0.0.0
```

On that worker, retrieve its credential with:

```text
./neopanel-node -config ./node.yaml node-token --raw
```

In the Windows master's **Admin -> Nodes**, add the worker's reachable API URL and
node token. Keep that token private. The Windows head must be able to reach the
worker's API; worker firewall rules should restrict it to the head/VPN. Use HTTPS
or a trusted private/VPN network, not a publicly exposed unencrypted node token.
The node serves control APIs, not the panel web interface. Creating a server with
that node selected makes it run on the worker, not on the Windows head.

A head used only from its own browser can remain loopback-only while controlling
remote nodes through outbound connections. Sharing the panel with other users,
or registering nodes back to it through an installer, requires a reachable panel
address instead of `localhost`.

## Optional Windows web-less node

On a separate Windows machine:

```powershell
.\start-windows.ps1 -Role node -ListenHost 0.0.0.0
.\NeoPanel-node.exe node-token --raw
```

Default node storage is `%LOCALAPPDATA%\NeoPanel-node`, separate from the master.
If testing master and node on one PC, use `-Port 8081` for the node and register
`http://127.0.0.1:8081` on the head. Use its real LAN/VPN IP when the head is on
another machine; `localhost` always refers to the machine making the connection.
A node-only binary cannot be switched into panel mode with `-role panel`.
Choose **Automatic** or **Host Direct** for a Windows node's process isolation
setting. Explicit Linux isolation engines are rejected on Windows.

## Network, security and operational limits

- Complete first-admin setup on loopback **before** making the head reachable.
  `-ListenHost 0.0.0.0` listens on network interfaces but does not configure firewall,
  NAT, DNS or TLS. Set the correct public panel URL in Appearance for emails/passkeys.
  Prefer a trusted VPN or TLS reverse proxy; configure only actual trusted proxies.
- No firewall/elevation changes are automatic. An administrator can add scoped
  Windows Firewall rules deliberately; do not open all panel/node/game ports globally.
- **Local Windows workloads are trusted-only.** Process/job ownership is not a
  filesystem, credential, user, or network sandbox. Do not host hostile tenant code
  on the Windows head. Keep untrusted workloads on properly isolated Linux workers.
- Windows Job Objects own the complete process tree and apply an aggregate
  committed-memory ceiling derived from the server memory policy. Closing the
  daemon kills owned jobs; this is crash cleanup, not a graceful world save.
  CPU quotas and task-count limits are not enforced on Windows.
- Per-process network accounting is unavailable on Windows; do not treat zero or
  unavailable network values as measured usage. CPU and RAM reporting use host APIs.
- Keep server data, database, config, credentials, and the sibling `servers-trash`
  directory when moving/upgrading an instance. Back up before replacing binaries.
  Windows auto-update via the Linux updater is not supported: stop, replace, restart.
- Absolute paths/configuration persist across launcher upgrades. Managed Java lives
  under the Windows user's application-data runtime folder (or the configured Java
  runtime override), not inside each server. Windows ACLs and account permissions
  remain the administrator's responsibility.

## Verification scope

Windows amd64 testing covers a real master EXE, API-only node EXE, administrator
setup, node pairing/proxying, and a real Java 21 fixture hosted at both locations.
The fixture verifies localhost TCP, CPU/RAM reporting, and graceful server stop;
it is not a Minecraft engine compatibility test and uses no worlds. Java installer
tests cover metadata, checksum rejection, safe extraction and runtime discovery.
Linux and Windows ARM64 cross-builds do not replace runtime testing on those hosts.
