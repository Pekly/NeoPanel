# NeoPanel

NeoPanel is a web panel for running Minecraft servers. Start a server, watch its console, edit files, install plugins or mods, and manage backups from your browser. You can keep everything on one machine or connect workers on other hosts.

The backend is written in Go, with the web interface and a SQLite database built in. You don't need PHP, a separate database server, or Docker to run the panel.

## What you can do

- Run Paper, Purpur, Fabric, Forge, NeoForge, Vanilla, proxies such as Velocity, and [Steel](https://steelmc.dev/).
- Manage Java versions, server settings, files, plugins, and mods without leaving the panel.
- Create backups and schedules, or warn players before a planned restart with a countdown you can cancel.
- Give friends or teammates access to individual servers, with control over what each person can do.
- Pin your usual servers, filter console output, and pick up where you left off after refreshing the page.

During panel maintenance, users can still access their servers' power controls and console. Their normal permissions still apply.

## Install on Linux

The installer is tested on Ubuntu 24.04 LTS. On your server, run:

```bash
curl -sSL https://raw.githubusercontent.com/Pekly/NeoPanel/main/install.sh | sudo bash
```

It builds NeoPanel, installs the required Go and Java tools, sets up the service and firewall rules, and can configure HTTPS with Caddy. You can [read the script](install.sh) first to see what it changes.

For an unattended install:

```bash
curl -sSL https://raw.githubusercontent.com/Pekly/NeoPanel/main/install.sh | sudo bash -s -- \
  --install --domain panel.example.com --email you@example.com -y
```

Open the address shown at the end of the install and create your admin account. The getting-started guide will help you create your first server. Choose **Local** if you want to run it on the same machine as the panel; you don't need to set up a separate worker.

Players join using the **Minecraft host and port** shown on the server card, not the panel's web address. Click the address to copy it.

### Add another host

Create an administrator API key under **Account → API Keys**, then run this on the Linux worker machine:

```bash
curl -sSL https://raw.githubusercontent.com/Pekly/NeoPanel/main/install.sh | sudo bash -s -- \
  --node --master https://panel.example.com \
  --api-key neo_your_admin_api_key_here --name eu-node-01 --location "Frankfurt, DE"
```

Workers run the game servers; you manage them from the same panel. Use HTTPS or a trusted private/VPN network between the panel and its workers, and keep your API keys private.

### Hosting on Windows

You can also connect a [Windows worker](docs/windows.md#optional-api-only-windows-worker) to host Minecraft servers on a Windows PC. The guide covers building it, connecting it to the panel, and where it stores your servers. It doesn't require the desktop app.

Windows hosting is for trusted servers: its process and memory controls don't isolate files or network access. For untrusted workloads, use a Linux host with the supported isolation configured.

## Updates and troubleshooting

```bash
sudo neopanel-update
sudo neopanel-installer          # reopen the installer menu
systemctl status neopanel
journalctl -u neopanel -f
```

Back up your worlds, configuration, and database before updating. Update the panel and workers together so their APIs stay compatible. If you're upgrading an older installation, check the [security upgrade notes](docs/security-upgrade.md).

On Linux, configuration is in `/etc/neopanel/config.yaml`, server data is under `/var/lib/neopanel`, and installer logs are in `/var/lib/neopanel/installer/install.log`.

## Sharing access and keeping data safe

Admins manage the panel and create servers. Other users can access servers they own or have been invited to. Set account roles in **Admin → Users** and manage each server's permissions from its **Team** page.

A few things worth setting up before inviting people:

- Configure SMTP and the public panel URL if you want invitation and password-reset emails to work.
- Give each person their own account and only the permissions they need.
- Stop a server before taking a consistent world backup. Restoring a backup replaces its current files.
- Don't run two panel instances against the same database or server folders.
- For untrusted Linux workloads, use Bubblewrap isolation. Resource limits alone aren't a filesystem sandbox.

More help: [invitations](docs/invitation-account-lifecycle.md), [file previews](docs/file-previews.md), [Trash](internal/trash/README.md), [server deletion](docs/server-deletion.md), and [workspace and console controls](docs/panel-qol.md).

## Working on NeoPanel

You'll need Go 1.26 or newer. The frontend is plain JavaScript and CSS, so there's no bundler or npm install step. Node.js is used for the frontend tests.

```bash
git clone https://github.com/Pekly/NeoPanel.git
cd NeoPanel
make build                       # writes bin/neopanel
go test ./...
go vet ./...
node --test web/tests/*.test.cjs
```

On Linux, you can also run the installer checks:

```bash
bash tests/install-isolation.test.sh
bash tests/install-build.test.sh
```

Found a bug? [Open an issue](https://github.com/Pekly/NeoPanel/issues) with what you were doing, what happened, your OS and version, and any useful logs. Please remove passwords and tokens before posting.

## License

[MIT](https://opensource.org/license/mit). The Steel logo is supplied by [SteelMC](https://steelmc.dev/), credited upstream to colonthreeing.
