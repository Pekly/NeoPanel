# NeoPanel on Windows

[![Repository size](https://img.shields.io/github/repo-size/Pekly/NeoPanel?style=flat-square)](https://github.com/Pekly/NeoPanel)

Use a Windows PC to host Minecraft servers and manage them from your NeoPanel web panel. This guide sets up the Windows worker, which runs the game servers and connects to an existing panel. The worker doesn't have its own web interface.

## What you'll need

- A Windows PC with enough memory and disk space for your servers.
- An existing NeoPanel panel and an administrator account on it.
- [Git for Windows](https://git-scm.com/downloads/win) and [Go 1.26 or newer](https://go.dev/dl/) to build the worker.
- A Java version compatible with the Minecraft server software you want to run. Steel doesn't need Java.

Go and Git are only needed for building and updating. You can copy the finished EXE to another Windows PC to run it.

## 1. Build the worker

Open PowerShell and run:

```powershell
git clone --branch windows-desktop-release --single-branch https://github.com/Pekly/NeoPanel.git
cd NeoPanel
New-Item -ItemType Directory -Force .\dist\windows-node | Out-Null
go build -trimpath -tags node -o .\dist\windows-node\NeoPanel-node.exe .\cmd\neopanel
```

The executable is now in `dist\windows-node`. Local builds are unsigned.

## 2. Create the configuration

```powershell
cd .\dist\windows-node
.\NeoPanel-node.exe -init
.\NeoPanel-node.exe node-token --raw
```

The `node-token` command prints your worker's connection token. You'll need it when adding the worker to your panel. Keep it private.

By default, the worker stores its configuration and data in `%LOCALAPPDATA%\NeoPanel-node`, not beside the EXE.

## 3. Start the worker

If the panel is on the same PC, use:

```powershell
.\NeoPanel-node.exe -host 127.0.0.1 -port 8081
```

If the panel is on another machine, listen on this PC's LAN or VPN address instead. For example, if this PC's address is `192.168.1.50`:

```powershell
.\NeoPanel-node.exe -host 192.168.1.50 -port 8081
```

Replace that example address with your own. Allow the panel to reach TCP port `8081` through Windows Firewall, restricting access to the panel's address. Use a trusted private/VPN network or HTTPS; don't expose the worker's unencrypted control API to the internet.

Leave PowerShell open while the worker is running. This command doesn't install a Windows service or configure firewall rules for you.

## 4. Connect it to your panel

1. Open **Admin → Nodes** in NeoPanel and add a node.
2. Enter the worker's reachable API URL, such as `http://192.168.1.50:8081` on your private network.
3. Paste the token from step 2.
4. Choose **Automatic** or **Host Direct** for process management.
5. Save the node, then select it when creating a Minecraft server.

Use `127.0.0.1` only when the panel and worker are on the same PC. From another machine, it points back to that machine—not to your worker.

Players connect to the Minecraft host and game port shown in the panel, not to port `8081`. Allow the game port through Windows Firewall separately; the usual Java Edition port is TCP `25565`.

## Daily use

Use the panel to start, stop, and restart servers, read the console, edit files, and manage backups. Choose the appropriate Java runtime for each Java server and review the Minecraft EULA before starting it.

To shut down the worker, press **Ctrl+C** in its PowerShell window and wait for it to finish. Don't close the window or end the process in Task Manager while worlds are being saved. Sleep, sign-out, and shutting down the PC interrupt hosting too.

Windows hosting is intended for trusted servers. Process and memory controls are supported, but they don't isolate files, credentials, or network access. CPU quotas, task-count limits, and per-process network accounting aren't available.

## Updating

Stop your game servers, then stop the worker with **Ctrl+C**. Back up `%LOCALAPPDATA%\NeoPanel-node`, along with any server folders you've configured elsewhere.

From your `NeoPanel` source folder, pull the update and rebuild:

```powershell
git pull --ff-only
go build -trimpath -tags node -o .\dist\windows-node\NeoPanel-node.exe .\cmd\neopanel
```

Start the worker again using the same address and port. Update your panel alongside the worker so their APIs stay compatible. Don't run two worker instances against the same data directory.

## Troubleshooting

- **The panel can't reach the worker:** check its listening address, port, firewall rule, and token. Make sure the PC is awake and the worker is still running.
- **Players can't join:** check the Minecraft port and game address. The worker's API port isn't the game port.
- **A Java server won't start:** check the selected Java version and the server console for the actual error.
- **A new panel feature isn't available:** rebuild and restart both the panel and worker; updating the browser files alone isn't enough.

For panel controls, see the [workspace and console guide](docs/panel-qol.md). If you're stuck, [open an issue](https://github.com/Pekly/NeoPanel/issues) with your Windows version, what happened, and relevant logs. Remove tokens and passwords before posting.

## License

[MIT](https://opensource.org/license/mit).
