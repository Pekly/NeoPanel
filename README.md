# NeoPanel

A lightweight Minecraft server panel: one Go binary, an embedded SQLite database, and no Docker stack to maintain. It runs Java servers and proxies, plus the native [Steel](https://steelmc.dev/) server. On my setup the panel idles at roughly 15–25 MB, leaving the memory for Minecraft.

## What's included

- Live console, resource charts, file editor, SFTP, and recoverable file-manager Trash.
- Paper, Purpur, LeafMC, Folia, Vanilla, Fabric, Forge, NeoForge, Velocity, Waterfall, BungeeCord, and Steel.
- Java management, Modrinth plugins/mods, backups and restore, schedules, and player management.
- Team permissions and email invitations, passkeys and 2FA, API keys, and audit logs.
- Remote worker nodes, themes and branding, and JavaScript extensions.

NeoPanel is for Minecraft, not arbitrary game-server eggs. It doesn't host MySQL/MariaDB or process payments. Only install extensions you trust: they run with panel-origin access.

## Linux setup

The installer is tested on Ubuntu 24.04 LTS:

```bash
curl -sSL https://raw.githubusercontent.com/Pekly/NeoPanel/main/install.sh | sudo bash
```

For a non-interactive panel install:

```bash
curl -sSL https://raw.githubusercontent.com/Pekly/NeoPanel/main/install.sh | sudo bash -s -- \
  --install --domain panel.example.com --email you@example.com -y
```

It installs Go and Java as needed, builds NeoPanel, configures a systemd service and firewall rules, and optionally sets up Caddy HTTPS. Configuration is in `/etc/neopanel/config.yaml`; data and worlds are under `/var/lib/neopanel`.

Installs and updates require a successful Bubblewrap preflight. If it fails, check `/var/lib/neopanel/installer/install.log` and resolve the host's namespace/package restrictions; the installer does not silently disable isolation. Select **Bubblewrap (`bwrap`)** on the node to require filesystem isolation. Systemd alone provides resource limits, not a filesystem sandbox.

### Add a worker

Create an administrator API key on the master under **Account → API Keys**, then run on the worker:

```bash
curl -sSL https://raw.githubusercontent.com/Pekly/NeoPanel/main/install.sh | sudo bash -s -- \
  --node --master https://panel.example.com \
  --api-key neo_your_admin_api_key_here --name eu-node-01 --location "Frankfurt, DE"
```

Workers have no web UI. Protect their APIs with HTTPS or a trusted private/VPN network; players connect directly to each game server's host and port.

### Updates and logs

```bash
sudo neopanel-update
sudo neopanel-installer          # installer menu
systemctl status neopanel
journalctl -u neopanel -f
```

Before upgrading, back up the database, configuration, and worlds, and read the [security upgrade notes](docs/security-upgrade.md). Update the panel **and all workers** together. Configure the public panel URL and trusted proxies explicitly; don't expose the backend behind an unrestricted proxy trust setting.

## Windows desktop (beta)

The desktop release line is `windows-desktop-release`, with Windows-only prerelease `v1.0.0-windows-beta.1`. It does not change `main` or the Linux release path.

[Download the Windows beta](https://github.com/Pekly/NeoPanel/releases/tag/v1.0.0-windows-beta.1), extract the ZIP, and run `NeoPanel-Desktop.exe`. Or build it from that branch:

```powershell
powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\build-windows-desktop.ps1
.\dist\windows-desktop-amd64\NeoPanel-Desktop.exe
```

The app opens its own window and backend—no console or browser tab needed. It can host trusted local servers and manage remote workers. This beta uses a portable EXE, requires WebView2, and stores data in `%LOCALAPPDATA%\NeoPanel-Desktop`. See the [desktop guide](docs/windows-desktop.md) for requirements, storage, and safe shutdown.

Prefer the browser UI on Windows? Use `build-windows.ps1` and `start-windows.cmd`; see the [Windows setup guide](docs/windows.md). Windows local workloads are **trusted-only**, not filesystem-isolated. Neither launcher automatically elevates privileges or opens the firewall.

## Before hosting real worlds

- Stop a server before taking a backup when you need a consistent world snapshot. Restore replaces the active files; keep independent backups.
- Never run two panel instances against the same database or server directories, or delete world/ownership locks to bypass a running instance.
- Use Linux with properly configured isolation for untrusted workloads. Windows Job Objects manage processes and memory, not filesystem or network access.
- Configure SMTP and the browser-facing panel address before using invitations, password resets, or passkeys. Keep the hostname stable once passkeys are registered.
- SFTP team access is opt-in. Grant `sftp.access` and the appropriate file permissions; use the server-specific login shown in the panel.

More detail: [security and deployment](docs/security-upgrade.md) · [invited-account lifecycle](docs/invitation-account-lifecycle.md) · [Trash](internal/trash/README.md) · [server deletion](docs/server-deletion.md) · [file previews](docs/file-previews.md) · [memory limits](internal/runner/MEMORY_POLICY.md).

## Build from source

Use Go 1.26 or newer with current security patches. The web interface is embedded; no Node.js build step or bundler is needed.

```bash
git clone https://github.com/Pekly/NeoPanel.git
cd NeoPanel
make build         # ./bin/neopanel
make build-linux   # Linux amd64 and arm64
make run           # development mode
```

The Windows desktop has its own module at `desktop/go.mod`. Linux installs and updates exclude desktop source, packaging, and Windows launchers/docs from their checkout, and build only `cmd/neopanel` (with `-tags node` for workers). They do not install Wails, WebView2, GTK, or WebKit. A local source install leaves your checkout untouched.

Regression checks:

```bash
go test ./...
go vet ./...
node --test web/tests/*.test.cjs
bash -n install.sh
bash tests/install-isolation.test.sh
bash tests/install-build.test.sh
```

Linux isolation/process tests need a Linux host; cross-compilation isn't a runtime sandbox test. Desktop builds and tests are separate from the root module; see the [desktop guide](docs/windows-desktop.md).

## License

MIT. Made for Minecraft admins who would rather spend their RAM on players than on PHP workers.

Steel logo supplied by [SteelMC](https://steelmc.dev/), credited upstream to colonthreeing.
