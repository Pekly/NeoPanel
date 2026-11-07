# NeoPanel 🚀

> **Hella Light Minecraft Server Management Panel & Daemon**  
> Tailored for **Ubuntu 24.04 LTS**, low-spec VPS, and power users who want blazing speed with near-zero overhead.

[![Ubuntu 24.04 LTS](https://img.shields.io/badge/Ubuntu-24.04%20LTS-E95420?logo=ubuntu&logoColor=white)](#)
[![Memory Usage](https://img.shields.io/badge/Idle%20RAM-%3C%2025%20MB-10b981)](#)
[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white)](#)
[![Database](https://img.shields.io/badge/Database-SQLite%20WAL-003B57?logo=sqlite&logoColor=white)](#)
[![SSL](https://img.shields.io/badge/SSL-Automated%20Let's%20Encrypt-003A70?logo=letsencrypt&logoColor=white)](#)

---

## ⚡ Why NeoPanel?

Traditional Minecraft panels (like Pterodactyl) require **PHP-FPM, Nginx, Redis, MariaDB, and Wings**, consuming **1 to 2 GB of RAM** just to display the dashboard before running a single server.

**NeoPanel does things differently:**
- **Single Compiled Static Binary**: The daemon, web server, and dashboard UI are compiled into one single executable.
- **Embedded Static UI**: No Node.js runtime, no PHP, no Python needed on the server.
- **Embedded SQLite with WAL Mode**: Zero external database overhead.
- **Total Idle RAM**: **~15 – 25 MB** for the whole panel.
- **Sub-Second Page Loads**: Instantaneous navigation and real-time WebSocket console.

---

## 🚀 Quick Install (Ubuntu 24.04 LTS)

Run the all-in-one interactive installer script with a single command (similar to the unofficial Pterodactyl installer):

```bash
curl -sSL https://raw.githubusercontent.com/Pekly/NeoPanel/main/install.sh | sudo bash
```

*(Or clone the repository and run `sudo bash install.sh`)*

### What the Installer Does Automatically:
1. **Validates OS**: Verifies Ubuntu 24.04 LTS (`noble`) and root privileges.
2. **Installs System Dependencies**: Installs `curl`, `jq`, `tar`, `ufw`, and system utilities.
3. **Installs Minecraft Java Runtimes**: Sets up OpenJDK 21 and 17 (supporting modern Minecraft 1.18 – 1.21+).
4. **Installs NeoPanel**: Installs `/usr/local/bin/neopanel` and generates `/etc/neopanel/config.yaml`.
5. **Configures Domain & Let's Encrypt SSL**:
   - Prompts for your domain (e.g. `panel.yourdomain.com`).
   - Sets up **Caddy** reverse proxy.
   - Automatically provisions and renews SSL certificates via Let's Encrypt.
6. **Configures Firewall (UFW)**:
   - Preserves SSH access (port `22`).
   - Opens Web ports (`80`, `443`).
   - Opens Minecraft server port allocations (`25565–25600` TCP/UDP).
7. **Systemd Service**:
   - Installs and enables `/etc/systemd/system/neopanel.service`.
   - Starts NeoPanel automatically on system boot.

---

## 🎮 Key Features

- **Real-Time Interactive Console & Popout View**:
  - Live WebSocket stream with ANSI and Minecraft (`§`) color code rendering.
  - Interactive command input with command history (`↑` and `↓` arrow keys).
  - Live connection health indicator dot (🟢 Connected, 🟡 Auto-Reconnecting, 🔴 Offline).
  - **Standalone Popout Console**: Pop out the terminal into a focused, distraction-free window or dedicated screen.
  - Power controls: Start, Stop, Restart, and Force Kill.
- **Modrinth Plugin Downloader**:
  - Search and install from 10,000+ plugins directly using the official Modrinth v2 API.
  - Category filters (Optimization, Administration, Chat, World Management, Economy, Security).
  - 1-click install and uninstall management.
- **Built-in Pure-Go SFTP Server (Port 2022)**:
  - Connect with FileZilla, Cyberduck, WinSCP or any standard SFTP client.
  - Chroot directory jailing isolates sessions directly into the Minecraft server directory.
- **One-Click Server Jar Downloader**:
  - Direct downloads for **PaperMC**, **Purpur**, **Vanilla (Mojang)**, and **Fabric**.
  - Automatically writes `eula.txt` (`eula=true`).
- **Web File Manager**:
  - In-browser code editor for `server.properties`, plugins configs, and YAML files.
  - File upload, creation, and deletion with breadcrumb navigation.
  - Quick-copy SFTP credentials modal.
- **Crash Detection & Auto-Restart**:
  - Automatically restarts Minecraft instances if they crash unexpectedly.
- **Host & Process Metrics**:
  - Live CPU %, Memory (RSS), and Disk usage monitoring.
- **Security**:
  - Bcrypt password hashing.
  - Cryptographically secure JWT authentication tokens.
  - Directory traversal prevention.

---

## 🛠️ Management Commands

Once installed, manage NeoPanel with standard system commands:

| Action | Command |
| :--- | :--- |
| **Instant Update from GitHub** | `sudo neopanel-update` |
| **Check Panel Status** | `systemctl status neopanel` |
| **View Live Panel Logs** | `journalctl -u neopanel -f` |
| **Restart Panel** | `systemctl restart neopanel` |
| **Stop Panel** | `systemctl stop neopanel` |
| **Check Caddy SSL Status** | `systemctl status caddy` |
| **View Caddy Logs** | `journalctl -u caddy -f` |
| **Interactive Installer Menu** | `bash <(curl -sSL https://raw.githubusercontent.com/Pekly/NeoPanel/main/install.sh)` |

### 🔄 Seamless Updating

Whenever you push any commit or change to your repository, simply run on your Ubuntu server:

```bash
sudo neopanel-update
```

It automatically fetches the freshest code, recompiles the binary with zero downtime risk, preserves your databases and world files, and restarts the service!

---

## ⚙️ Configuration (`/etc/neopanel/config.yaml`)

```yaml
server:
  host: "127.0.0.1"   # 127.0.0.1 behind Caddy/Nginx, or 0.0.0.0 standalone
  port: 8080
  mode: "behind_proxy"

paths:
  data_dir: "/var/lib/neopanel"
  servers_dir: "/var/lib/neopanel/servers"
  db_path: "/var/lib/neopanel/neopanel.db"

security:
  jwt_secret: "<auto-generated-random-hex>"
  token_expiration_hours: 72

java:
  default_path: "java"

sftp:
  enabled: true
  port: 2022
  key_path: "/var/lib/neopanel/sftp_host_key"
```

---

## 🏗️ Building From Source

```bash
# Clone the repository
git clone https://github.com/Pekly/NeoPanel.git
cd NeoPanel

# Compile local binary
make build

# Cross-compile for Linux (amd64 / arm64)
make build-linux

# Run directly
make run
```

---

## 📄 License

MIT License. Designed with ❤️ for Minecraft server administrators and low-latency hosting.

