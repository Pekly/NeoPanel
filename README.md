# NeoPanel

[![MIT license](https://img.shields.io/badge/license-MIT-blue.svg)](#license)
[![Go 1.26+](https://img.shields.io/badge/Go-1.26%2B-00ADD8?logo=go&logoColor=white)](https://go.dev/dl/)
[![Linux first](https://img.shields.io/badge/focus-Linux-FCC624?logo=linux&logoColor=black)](#install-on-linux)
[![Windows beta](https://img.shields.io/badge/Windows-beta-0078D4)](https://github.com/Pekly/NeoPanel/tree/windows-desktop-release)

NeoPanel is a Minecraft server panel written in Go. It ships as one binary with an embedded web interface and SQLite database, so there’s no PHP runtime or Docker stack to set up for the panel.

Run servers on the panel’s machine, or connect workers on other hosts. You get a live console, file editor, Java management, plugins and mods, backups, schedules, and team access. It supports the usual Java servers and proxies, including Paper, Fabric, Forge, and Velocity, as well as [Steel](https://steelmc.dev/).

**Linux is the main focus.** The [Windows desktop branch](https://github.com/Pekly/NeoPanel/tree/windows-desktop-release) is a secondary project and will receive updates occasionally. Don’t expect every Linux update to get an immediate Windows build.

## Install on Linux

Ubuntu 24.04 LTS is the tested installer target. On a machine you control:

```bash
curl -sSL https://raw.githubusercontent.com/Pekly/NeoPanel/main/install.sh | sudo bash
```

The installer builds NeoPanel, installs the required Go/Java tools, configures a systemd service and firewall rules, and can set up Caddy for HTTPS. Review [install.sh](install.sh) before running it as root if you want to see exactly what it changes.

For an unattended install:

```bash
curl -sSL https://raw.githubusercontent.com/Pekly/NeoPanel/main/install.sh | sudo bash -s -- \
  --install --domain panel.example.com --email you@example.com -y
```

Open the address printed by the installer. The first page creates your administrator account; the **Getting started** guide then walks through account security, choosing a host, creating a server, and console/backups. You can skip it and reopen it from the dashboard. Guide progress is saved for your account in that browser, not shared across devices.

The panel’s web address and the Minecraft address are different. Players connect to the **game host and port** shown on the server card.

### Add another host

Create an administrator API key under **Account → API Keys**, then run this on the worker machine:

```bash
curl -sSL https://raw.githubusercontent.com/Pekly/NeoPanel/main/install.sh | sudo bash -s -- \
  --node --master https://panel.example.com \
  --api-key neo_your_admin_api_key_here --name eu-node-01 --location "Frankfurt, DE"
```

Workers don’t serve a web interface. Keep their APIs behind HTTPS or a trusted private/VPN network. You don’t need a worker just to host on the panel’s own machine—choose **Local**.

### Updates and troubleshooting

```bash
sudo neopanel-update
sudo neopanel-installer          # reopen the installer menu
systemctl status neopanel
journalctl -u neopanel -f
```

Configuration lives in `/etc/neopanel/config.yaml`; data and worlds are under `/var/lib/neopanel`. Installer logs are in `/var/lib/neopanel/installer/install.log`.

Back up your database, configuration, and worlds before upgrading. Update the panel and workers together, and read the [security upgrade notes](docs/security-upgrade.md) when updating an older installation.

## Windows desktop

[Download the portable beta](https://github.com/Pekly/NeoPanel/releases/tag/v1.0.0-windows-beta.1), extract the ZIP, and run `NeoPanel-Desktop.exe`. It opens its own window and needs Microsoft Edge WebView2 Runtime. No Go installation is needed to run the download.

Desktop development stays on **`windows-desktop-release`**. It is not the primary focus and gets occasional updates; the published beta may lag behind branch changes. See the [desktop guide on that branch](https://github.com/Pekly/NeoPanel/blob/windows-desktop-release/docs/windows-desktop.md) for building, storage, and safe shutdown. The desktop EXE is the Windows panel edition—there is no separate console/browser package or startup script. An optional [API-only Windows worker](docs/windows.md) can host servers on another PC.

Minimize the desktop app to keep local servers running. Closing with active local servers asks whether to save/stop them or keep running. Windows local hosting is **trusted-only**: Job Objects manage processes and memory, not filesystem or network isolation.

## Accounts

**Admins** manage the panel and create servers. **Users** access their own or shared
servers. Set their role in **Admin → Users**; team permissions stay per server.
The getting-started guide is shown only to admins, not normal users.

## Before inviting players

- Use configured Linux isolation for untrusted workloads. The installer requires a successful Bubblewrap preflight; it won’t quietly bypass a failed check. Select **Bubblewrap (`bwrap`)** on the node to require filesystem isolation. Systemd resource limits alone are not a filesystem sandbox.
- Stop a server before taking a consistent world backup. Restore replaces active files, so keep a separate backup too.
- Don’t run two panels against the same database or server folders, or remove ownership locks to bypass a running instance.
- Configure SMTP and the public panel URL before relying on invitation or password-reset emails. Use HTTPS and configure trusted proxies explicitly.
- Give teammates their own accounts and only the permissions they need. Extensions run with panel-origin access; install only ones you trust.

More details: [invited accounts](docs/invitation-account-lifecycle.md), [file-manager Trash](internal/trash/README.md), [server deletion](docs/server-deletion.md), [file previews](docs/file-previews.md), and [memory limits](internal/runner/MEMORY_POLICY.md).

See [panel quality-of-life controls](docs/panel-qol.md) for workspace recovery, console filters, pinned servers, maintenance access, and cancellable restart countdowns.

## Working on the code

Use Go 1.26 or newer with current security patches. The web UI doesn’t need a bundler or npm install.

```bash
git clone https://github.com/Pekly/NeoPanel.git
cd NeoPanel
make build                       # bin/neopanel
go test ./...
go vet ./...
node --test web/tests/*.test.cjs
bash tests/install-isolation.test.sh
bash tests/install-build.test.sh
```

`main` is the Linux/headless line. The desktop branch keeps Wails in a separate `desktop/go.mod`; Linux install/update checkouts exclude desktop-only files and build the root module. Linux isolation tests need a Linux host—cross-compiling is not a sandbox test.

Found a problem? [Open an issue](https://github.com/Pekly/NeoPanel/issues) with your OS, branch/version, steps to reproduce, and relevant logs. Remove tokens, passwords, and personal data first.

## License

[MIT](https://opensource.org/license/mit). The Steel logo is supplied by [SteelMC](https://steelmc.dev/), credited upstream to colonthreeing.
