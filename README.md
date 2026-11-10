# NeoPanel 🚀

> **Pterodactyl from Temu** — A single static Go binary that sips **~15–25 MB of idle RAM**, requires **zero Docker / PHP / Redis / MariaDB / Wings**, and delivers **~80% of Pterodactyl's Minecraft workflow**.

[![Linux Multi-OS](https://img.shields.io/badge/Linux-Debian%20%7C%20Ubuntu%20%7C%20RHEL%20%7C%20Arch%20%7C%20Alpine-E95420?logo=linux&logoColor=white)](#)
[![Idle RAM](https://img.shields.io/badge/Idle%20RAM-~15--25%20MB-10b981)](#)
[![Go 1.22+](https://img.shields.io/badge/Go-1.22%2B-00ADD8?logo=go&logoColor=white)](#)
[![Database](https://img.shields.io/badge/Database-SQLite%20WAL%20(Pure%20Go)-003B57?logo=sqlite&logoColor=white)](#)
[![SSL](https://img.shields.io/badge/SSL-Caddy%20%2B%20Let's%20Encrypt-003A70?logo=letsencrypt&logoColor=white)](#)

---

## 🛒 The "Pterodactyl from Temu" Reality

Running full **Pterodactyl Panel + Wings** for Minecraft means maintaining **PHP-FPM, Nginx, Redis, MariaDB, Composer worker queues, Docker daemons, and Wings** — consuming **1–2 GB of RAM** before a single Minecraft world chunk loads.

**NeoPanel takes the opposite approach:** one compiled Go executable contains the HTTP dashboard, WebSocket multiplexer, embedded SQLite WAL engine (`modernc.org/sqlite`), chroot SFTP server, multi-node cluster reverse proxy, subuser RBAC, cron scheduler, pure-Go token-bucket rate limiter, and Modrinth/mcjarfiles downloaders. 

Your VPS RAM is spent running Minecraft servers, not administrative scaffolding.

| Feature / Metric | Pterodactyl + Wings | NeoPanel |
| :--- | :--- | :--- |
| **Daemon Architecture** | PHP 8.x + Nginx + MariaDB + Redis + Docker + Wings | **1 Static Go Binary** + SQLite WAL + Java |
| **Idle Memory Footprint** | ~1,024 – 2,048 MB | **~15 – 25 MB** |
| **Process Isolation** | Heavy Docker container layers | **Bubblewrap (`bwrap`) Jail + Systemd Slices** |
| **Setup & Dependencies** | Docker images + Compose + SQL schemas + Node.js | **Single `curl \| sudo bash` script** |
| **Multi-Node Clustering** | Wings tokens + TLS per node | **Daemon `--node` mode + Token / API key pairing** |
| **Subuser Access** | Granular ACL checkboxes | **Per-server RBAC subuser permissions** |
| **Automated Billing API** | External Wings API wrappers | **Native Client API (`/api/client-api/v1/provision`)** |
| **Resource Limits** | Docker cgroups | **`systemd-run` PID (`TasksMax=512`) & CPU caps** |

---

## 🛡️ Security & Sandboxing Architecture

### 1. Filesystem Isolation (Bubblewrap `bwrap`)
Untrusted Minecraft plugins can execute arbitrary Java code and read `/etc/passwd`, `/etc/shadow`, or neighbor server files on uncontainerized hosts. NeoPanel invokes server instances inside unprivileged user namespaces via `bwrap`:
- **Read-only system binds**: `/usr`, `/lib`, `/lib64`, `/bin`, `/sbin`, `/etc/alternatives`, and JVM runtime directories.
- **Isolated scratch mounts**: Empty private `tmpfs` mounts for `/tmp` and `/run`.
- **Server jail**: The server's working directory is the **only read-write mount**. All neighboring server directories, `/root`, and sensitive system paths are completely unreachable.
- **Network access**: Net namespace is shared (`--share-net`) for multiplayer networking while isolation boundaries stay locked.

### 2. Fork-Bomb & CPU Starvation Defense (Systemd Slices)
Malicious plugins or runaway threads cannot exhaust host process tables:
- Process groups run inside dedicated systemd scopes (`systemd-run --scope`).
- **PID Limit**: Hard-capped via `-p TasksMax=512` to instantly mitigate fork-bombs.
- **CPU Quota**: Optional `-p CPUQuota=...%` to prevent a rogue instance from pinning 100% of all VPS cores.

### 3. Pure Go Token-Bucket Rate Limiter
NeoPanel protects endpoints from brute force and scraper spam without Redis or external dependencies:
- In-memory thread-safe rate limiter (`internal/api/ratelimit.go`).
- **Global API rate**: 60 req/sec (burst 100).
- **Authentication login rate**: 1 req/sec (burst 15).
- Background cleanup routines automatically prune inactive IP states to keep memory usage negligible.

---

## ⚡ What You Get: The 80% for Minecraft

### 🎮 Live Console & Monitoring
- **Real-Time WebSocket Stream**: Live log streaming with ANSI & Minecraft `§` formatting, interactive input bar with command history (`↑`/`↓`), quick-command menu, and standalone popout terminal.
- **Slide-Out Drawer (`Ctrl + \``)**: Access the live console from any sub-page or file manager without leaving your current work.
- **Interactive SVG Telemetry**: Real-time crosshair hover tooltips displaying exact timestamps and values for CPU %, Memory MiB/GiB, and Network In/Out.
- **1-Click Copyable IP**: Click any host address chip to copy `fqdn:port` instantly.

### 📁 Pterodactyl-Style File Manager
- **Full File Operations**: Upload (multi-file drag-and-drop with progress), rename, delete, download, and in-browser editor (`Tab` indentation, keyboard shortcuts).
- **Batch Multi-Select**: Select multiple items with shift-click support to **Move** (with `/home/container/` auto-completion), **Archive / Zip** (`.zip` / `.tar.gz`), or **Extract / Unzip** (`.zip`, `.tar.gz`, `.tar`, `.jar`).
- **View Switcher**: Switch between **List**, **Icons**, and **Tiles** views.
- **Smart Sorting**: Folders always stay on top, with sorting by Name, Size, or Extension.

### ⏰ Database-Backed Automated Schedules
- **Go Cron Scheduler**: Recurring automation evaluated every 30 seconds by an internal worker.
- **Supported Actions**: Console commands (`say`, `save-all`), graceful restarts, world backups (`.tar.gz`), and graceful shutdowns.
- **Interactive Management**: Modal editor with human-readable cron presets, active toggles, duplicate-minute safeguards, and instant **"Run Now"** triggers.

### 👥 Subusers & Granular RBAC Permissions
- **Invite Teammates**: Add staff members or builders by email or username.
- **Granular Permissions**:
  `console.view`, `console.command`, `server.power`, `files.view`, `files.edit`, `files.delete`, `settings.view`, `settings.edit`, `database.manage`, `backup.manage`, `subusers.manage`, `schedules.manage`.
- **Node-Level Policy Padlocks**: Cluster administrators can lock RAM allocation, port numbers, or Java runtimes on specific nodes. Restricted settings display padlock chips in the UI and enforce `403 Forbidden` checks on the backend.

### 🔌 Client Area Billing Integration (`POST /api/client-api/v1/provision`)
- Seamlessly connects storefronts (WHMCS, Blesta, Paymenter, custom billing portals).
- Automatically provisions user accounts, allocates ports, targets specific nodes, and sets memory/CPU limits.
- **Admin Area Integration Hub**: Live copyable code snippets in PHP (WHMCS/Blesta), Node.js, Python, and cURL.
- **Sidebar Visibility**: Toggle whether the "Client Area" navigation button is visible.

### 💤 Free Node Idle Player Watchdog
- Designate any cluster node as a **Free Node** (`is_free: true`).
- Running servers with **0 connected players** automatically shut down after a configurable timeout (default 15 minutes) to conserve VPS resources.
- Fully audited with clean server log notifications before shutdown.

### ⚙️ Quick Configs & Server Icon Tools
- **Server Icon Uploader**: Upload 64×64 PNG icons directly (`server-icon.png`) for the in-game multiplayer list.
- **Quick Configs Editor**: Sliders for view distance (2–32), simulation distance (2–32), spawn protection, toggles for flight, PvP, whitelist, command blocks, and 1-click EULA acceptance.
- **Presets**: 1-click presets for *Extreme Performance* (low render), *Balanced*, and *Ultra Visual*.

### 🌐 Global Timezone Customization
- Display all activity logs, console timestamps, file modification dates, and schedule timers in your preferred timezone (UTC, America, Europe, Asia, Australia, etc.).

---

## 📦 Multi-OS Installation

The unified installer (`install.sh`) automatically detects your Linux distribution and package manager (`apt`, `dnf`, `yum`, `pacman`, `apk`):

Supported OS: **Ubuntu 20.04/22.04/24.04 LTS, Debian 11/12, CentOS 8/9, RHEL 8/9, AlmaLinux, Rocky Linux, Fedora, Arch Linux, Alpine Linux**.

### 1. Install Master Panel
```bash
curl -sSL https://raw.githubusercontent.com/Pekly/NeoPanel/main/install.sh | sudo bash
```

Unattended non-interactive installation:
```bash
curl -sSL https://raw.githubusercontent.com/Pekly/NeoPanel/main/install.sh | sudo bash -s -- \
  --install --domain panel.yourdomain.com --email admin@yourdomain.com -y
```

### 2. Connect Dedicated Cluster Nodes (`--node` Mode)
Install a lightweight worker node on another VPS without serving any frontend web code:

**Step 1:** On the Master Panel, create an Admin API key (`neo_...`) via `Admin Area -> API Keys` or via SSH:
```bash
sudo neopanel api-key create --name "eu-node-01"
```

**Step 2:** On the remote worker VPS:
```bash
curl -sSL https://raw.githubusercontent.com/Pekly/NeoPanel/main/install.sh | sudo bash -s -- \
  --node \
  --master https://panel.yourdomain.com \
  --api-key neo_your_admin_api_key_here \
  --name eu-node-01 \
  --location "Frankfurt, DE"
```

---

## 🛠️ CLI Management

| Task | Command |
| :--- | :--- |
| **Update Panel / Node to Latest Release** | `sudo neopanel-update` |
| **Interactive Installer Menu** | `sudo neopanel-installer` |
| **Show / Rotate Node Secret Token** | `sudo neopanel node-token [--rotate] [--raw]` |
| **Create / List / Revoke API Keys** | `sudo neopanel api-key create\|list\|delete ...` |
| **Connect Remote Node via CLI** | `sudo neopanel node connect --name n1 --url http://IP:8080 --token neo_...` |
| **Check Daemon Status** | `systemctl status neopanel` |
| **Follow Live Daemon Logs** | `journalctl -u neopanel -f` |

---

## 🏗️ Building From Source

Requires **Go 1.22+**. All frontend assets in `web/` are embedded directly into the Go binary at compile time via `//go:embed`:

```bash
git clone https://github.com/Pekly/NeoPanel.git
cd NeoPanel

# Compile binary -> ./bin/neopanel
make build

# Cross-compile static Linux binaries (amd64 / arm64)
make build-linux

# Run directly
make run
```

---

## 📄 License

MIT License — Crafted for Minecraft admins who prefer spending RAM on players instead of PHP workers.
