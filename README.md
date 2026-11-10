# NeoPanel 🚀

> **Pterodactyl from Temu** — except it runs in a **single static Go binary**, sips **~15–25 MB of idle RAM**, needs **zero Docker / PHP / Redis / MariaDB / Wings**, and still gives you **~80% of Pterodactyl's Minecraft workflow**.

[![Ubuntu 24.04 LTS](https://img.shields.io/badge/Ubuntu-24.04%20LTS-E95420?logo=ubuntu&logoColor=white)](#)
[![Idle RAM](https://img.shields.io/badge/Idle%20RAM-~15--25%20MB-10b981)](#)
[![Go 1.22+](https://img.shields.io/badge/Go-1.22%2B-00ADD8?logo=go&logoColor=white)](#)
[![Database](https://img.shields.io/badge/Database-SQLite%20WAL%20(Pure%20Go)-003B57?logo=sqlite&logoColor=white)](#)
[![SSL](https://img.shields.io/badge/SSL-Caddy%20%2B%20Let's%20Encrypt-003A70?logo=letsencrypt&logoColor=white)](#)

---

## 🛒 The "Pterodactyl from Temu" Pitch

Running full **Pterodactyl Panel + Wings** for Minecraft means juggling **PHP-FPM, Nginx, Redis, MariaDB, Composer queues, and Docker/Wings** — burning **1–2 GB of RAM** before a single block is placed.

**NeoPanel cuts straight to the chase:** one single Go executable holds the web dashboard, REST + WebSocket APIs, embedded SQLite database (`modernc.org/sqlite`), chroot SFTP server, multi-node cluster proxy, process supervisor, and Modrinth/mcjarfiles downloaders. That leaves almost 100% of your VPS RAM for Minecraft chunks and plugins.

| Comparison | Pterodactyl + Wings | NeoPanel |
| :--- | :--- | :--- |
| **Runtime Stack** | PHP 8.x + Nginx + MariaDB + Redis + Docker + Wings | **1 Go binary** + SQLite WAL + Java |
| **Idle RAM Usage** | ~1.0 – 2.0 GB | **~15 – 25 MB** |
| **Install Footprint** | Docker images + Compose + database setup | **One `curl \| sudo bash` script** |
| **Target Workload** | Generic game containers (Eggs) | **Purpose-built for Minecraft & Proxies** |
| **Multi-Node Clustering** | Wings tokens + SSL per node | **Daemon `--node` mode + Admin API key pairing** |

### ✅ The ~80% of Pterodactyl You Get for Minecraft

Everything you touch day-to-day when running Survival, Modded, or Bungee/Velocity networks:

- **Live WebSocket Console + Fly-Out Drawer**: Real-time log streaming with ANSI & `§` colors, plugin/timestamp badges, command history (`↑`/`↓`), quick-command dropdown, standalone popout terminal, and a global slide-out console drawer (`Ctrl + \``) accessible from any page.
- **Pterodactyl-Style File Manager**:
  - In-browser editor (`Tab` indenting, 5 MB file editing), drag-and-drop multi-file upload with live progress bar, create file/folder, rename, copy path, download.
  - **Bulk Multi-Select**: Move files & folders across directories with `/home/container/` path completion + quick folder chips & overwrite option; **Archive / Zip** (`.zip` & `.tar.gz`); **Extract / Unzip** (`.zip`, `.tar.gz`, `.tgz`, `.tar`, `.jar`).
  - **View Switcher**: Switch between **List**, **Icons**, and **Tiles** views across Files, Plugins, Versions, and Dashboard.
- **Built-In Chroot SFTP Server (Port `2022`)**: Connect directly with FileZilla, WinSCP, or Cyberduck using your panel credentials or SSH public keys, jailed inside each server directory.
- **11 Server Engines & Complete MC Release Catalog (`26.3` → `1.0`)**:
  - **Engines**: Paper, Purpur, LeafMC, Folia, Vanilla, Fabric, Forge, NeoForge, Velocity, Waterfall, BungeeCord.
  - Powered by **mcjarfiles.com**, Paper Fill v3, Purpur API, Fabric Meta, and Mojang manifest — with safe 1-click Reinstall (choose whether to preserve Worlds, Configs, and Plugins/Mods) + automatic `eula.txt`.
- **Smart Java Runtime Manager (Java `25`, `21`, `17`, `8`)**: Auto-detects installed JVMs, installs missing Eclipse Temurin / OpenJDK runtimes from the Admin Area or CLI, matches the recommended Java major version per Minecraft release, and **halts crash-loops after 2 failed startup attempts** (`180s` startup timeout).
- **Modrinth Plugin & Mod Store**: Search Modrinth v2 with category, loader, sort, and MC version filters, plus 1-click install with optional backup confirmation, enable/disable toggle, and uninstall.
- **Per-Server Databases + SQL Studio**: Provision isolated pure-Go SQLite databases per server, rotate credentials, browse table schemas, run interactive SQL queries in-browser, and export `.sql` dumps or `.db` binaries.
- **Backups, Allocations & Players**: Manual `.tar.gz` world backups with download/delete, port allocations + automatic UFW rules, subdomain SRV/A helper, online player manager (OP/De-OP, Kick, Ban/Pardon, Whitelist), and rich searchable Activity/Audit logs with CSV/JSON export.
- **Full Minecraft MOTD Color & Gradient Support**: Renders legacy `§` / `&` codes, Java `.properties` `\u00A7` & `\n` escapes, `§x§R§R§G§G§B§B` / `&#RRGGBB`, Iridium `{#RRGGBB>}...{#RRGGBB<}`, and Paper MiniMessage `<gradient:...>` / `<rainbow>` / named tags directly in the server header & Settings Live Preview.
- **Multi-Node Clustering (API-Key Paired)**: Run master panel + remote game nodes using the exact same binary (`--node`), managing all servers seamlessly through clean routes (`/server/{node}/{id}/...`) via reverse proxy.
- **White-Label Panel Customizer + Security**: 6 built-in themes, custom accent/glass blur/radius/CSS, Dark/Light/OLED/Compact modes, Activity & Audit Log filtering with JSON/CSV export, API keys (`neo_...`), bcrypt auth, JWT sessions, optional Cloudflare Turnstile CAPTCHA, and Caddy automated HTTPS.

### ⚠️ The ~20% Left Out (So You Know Before Switching)

Because it's Pterodactyl *from Temu*, here is what NeoPanel deliberately does **not** try to do:

- **No Arbitrary Docker Eggs**: Runs native JVM processes on bare metal/VPS rather than Docker containers (no SteamCMD / Rust / Python / Discord bot Eggs).
- **No Per-Server Subuser Permission Checkboxes**: Role-based panel users (`admin` and `user`) without granular 50-toggle per-server ACLs.
- **No Built-In MariaDB/MySQL Server Hosting**: Databases are embedded zero-config SQLite databases (with SQL Studio + export), not provisioned MySQL daemon instances.
- **No Automated WHMCS / Paymenter Billing Provisioner**: The Billing / Invoices view is a static VPS reference placeholder.

---

## ⚡ Quick Install (Ubuntu 24.04 LTS)

### 1. Install Master Panel (Full Dashboard + Local Daemon + Caddy SSL + UFW)

```bash
curl -sSL https://raw.githubusercontent.com/Pekly/NeoPanel/main/install.sh | sudo bash
```

You can also pass flags directly for non-interactive setup:

```bash
curl -sSL https://raw.githubusercontent.com/Pekly/NeoPanel/main/install.sh | sudo bash -s -- \
  --install --domain panel.yourdomain.com --email you@yourdomain.com -y
```

**What the installer sets up automatically:**
1. Installs system tools (`curl`, `jq`, `tar`, `ufw`, `git`) + **Go 1.24+** toolchain if needed.
2. Installs **Java 25, 21, 17, and 8** runtimes (`/usr/lib/jvm/...`).
3. Builds `/usr/local/bin/neopanel` and generates `/etc/neopanel/config.yaml`.
4. Optionally configures **Caddy** reverse proxy + automatic **Let's Encrypt SSL**.
5. Configures **UFW firewall** (`22`, `80`, `443`, `2022` SFTP, and Minecraft range `25565:25600`).
6. Optionally configures **Cloudflare Turnstile** protection.
7. Installs `systemd` service (`neopanel.service`) and helper commands (`neopanel-update`, `neopanel-installer`).

---

### 2. Connect Dedicated Cluster Nodes (`--node` Mode)

Add extra VPS machines as pure worker nodes **without serving any web UI** on the worker — everything appears seamlessly inside your Master Panel.

**Step A:** On the Master Panel (`Admin Area -> Cluster Nodes` or `Account -> API Keys`), create an Admin API key (`neo_...`), or run via SSH on the Master host:

```bash
sudo neopanel api-key create --name "eu-node-01"
```

**Step B:** On the new worker VPS (Ubuntu 24.04 LTS), run:

```bash
curl -sSL https://raw.githubusercontent.com/Pekly/NeoPanel/main/install.sh | sudo bash -s -- \
  --node \
  --master https://panel.yourdomain.com \
  --api-key neo_your_admin_api_key_here \
  --name eu-node-01 \
  --location "Frankfurt, DE"
```

That compiles `neopanel`, installs Java 25/21/17/8, configures `role: "node"` + `security.node_token`, opens firewall ports, starts `neopanel.service`, and auto-registers the node with your Master Panel!

---

## 🛠️ CLI & Management Commands

The `neopanel` binary doubles as an admin CLI alongside standard systemd tools:

| Action | Command |
| :--- | :--- |
| **Update Panel / Node from GitHub** | `sudo neopanel-update` |
| **Open Installer Menu** | `sudo neopanel-installer` |
| **Show / Rotate Node Token** | `sudo neopanel node-token [--rotate] [--raw]` |
| **Create / List / Revoke API Keys** | `sudo neopanel api-key create\|list\|delete ...` |
| **Connect Remote Node via CLI** | `sudo neopanel node connect --name n1 --url http://IP:8080 --token neo_node_...` |
| **List / Test / Remove Nodes** | `sudo neopanel node list\|test\|remove ...` |
| **Check Service Status** | `systemctl status neopanel` |
| **Live Daemon Logs** | `journalctl -u neopanel -f` |
| **Caddy SSL Status / Logs** | `systemctl status caddy` / `journalctl -u caddy -f` |

### Daemon Runtime Flags

```bash
neopanel -config /etc/neopanel/config.yaml -host 0.0.0.0 -port 8080 -role panel
neopanel -config /etc/neopanel/config.yaml -role node   # or shorthand: -node
```

### Keyboard Shortcuts in the Web UI

- **`Ctrl + K`**: Open Global Command Palette (switch servers, jump pages, run quick actions)
- **`Ctrl + \``**: Toggle the Slide-Out Live Console Drawer on any server page

---

## ⚙️ Configuration (`/etc/neopanel/config.yaml`)

```yaml
server:
  host: "127.0.0.1"             # 127.0.0.1 behind Caddy, or 0.0.0.0 standalone / node
  port: 8080
  mode: "behind_proxy"          # "behind_proxy" or "standalone"
  role: "panel"                 # "panel" (Master + Web UI) or "node" (Cluster Daemon only)
  startup_timeout_seconds: 180

paths:
  data_dir: "/var/lib/neopanel"
  servers_dir: "/var/lib/neopanel/servers"
  db_path: "/var/lib/neopanel/neopanel.db"

security:
  jwt_secret: "<auto-generated-hex>"
  token_expiration_hours: 72
  node_token: "neo_node_<auto-generated-hex>"
  turnstile:
    enabled: false
    site_key: ""
    secret_key: ""

java:
  default_path: "java"

sftp:
  enabled: true
  port: 2022
  key_path: "/var/lib/neopanel/sftp_host_key"
```

---

## 🌐 Core HTTP & WebSocket Endpoints (For Automation)

Authenticate external automation scripts using `Authorization: Bearer neo_<api_key>` or JWT login:

- **Cluster & Nodes**: `GET /api/cluster/servers`, `GET|POST /api/nodes`, `ALL /api/nodes/{id}/proxy/*`, `WS /api/nodes/{id}/ws/{serverId}`
- **Servers & Power**: `GET|POST /api/servers`, `GET|PUT|DELETE /api/servers/{id}`, `POST /api/servers/{id}/power`, `POST /api/servers/{id}/command`
- **Console Stream**: `WS /api/servers/{id}/ws`
- **Files & Archives**: `GET|DELETE /api/servers/{id}/files`, `GET|POST /api/servers/{id}/files/content`, `POST /api/servers/{id}/files/rename`, `POST /api/servers/{id}/files/move`, `POST /api/servers/{id}/files/compress`, `POST /api/servers/{id}/files/unzip`, `POST /api/servers/{id}/files/upload`, `GET /api/servers/{id}/files/download`
- **Software & Java**: `GET /api/versions`, `POST /api/servers/{id}/software`, `POST /api/servers/{id}/reinstall`, `GET /api/system/java`, `POST /api/system/java/install`
- **Plugins, Backups, DBs**: `/api/plugins/search`, `/api/servers/{id}/plugins*`, `/api/servers/{id}/backups*`, `/api/servers/{id}/databases*`

---

## 🏗️ Building From Source

Requires **Go 1.22+** (zero Node.js or asset bundler step needed — everything in `web/` is embedded via `//go:embed`):

```bash
git clone https://github.com/Pekly/NeoPanel.git
cd NeoPanel

# Build native binary -> ./bin/neopanel
make build

# Cross-compile static Linux binaries (amd64 & arm64)
make build-linux

# Run directly in dev mode
make run
```

---

## 📄 License

MIT License — crafted for Minecraft admins who prefer spending RAM on players instead of PHP workers.
