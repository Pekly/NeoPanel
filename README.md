# NeoPanel

A lightweight Minecraft server panel. Think "Pterodactyl from Temu": it does most of what you'd actually use day to day, but it's one Go binary instead of a whole stack.

I built it because running Pterodactyl for a Minecraft-only setup meant PHP, Nginx, Redis, MariaDB, Docker and Wings, and that eats a lot of RAM before a single block is placed. NeoPanel is one executable with an embedded SQLite database. It idles at roughly 15 to 25 MB on my setup, so the memory goes to your server instead.

## How it compares

| | Pterodactyl + Wings | NeoPanel |
|---|---|---|
| Stack | PHP, Nginx, MariaDB, Redis, Docker, Wings | One Go binary, SQLite, Java |
| Idle RAM | around 1 to 2 GB | around 15 to 25 MB |
| Install | Docker images, Compose, database setup | One install script |
| Built for | Any game via Eggs | Minecraft and proxies only |
| Multi-node | Wings tokens and SSL per node | Same binary in `--node` mode, paired with an API key |

## What you get

- **Live console** over WebSocket with colors, command history, a popout terminal, and a slide-out drawer you can open from any server page with `Ctrl + \``. CPU, memory and network charts sit alongside it.
- **File manager** with an in-browser editor, unsaved-change prompts, per-user/server draft recovery within the same browser tab, and `Ctrl/Cmd + S`. Drag-and-drop uploads have a queue with per-file errors, cancel/retry controls and a completion summary. Bulk select/move and zip/tar.gz archiving/extraction are also included.
- **Built-in SFTP** on port 2022. Log in with your panel credentials or an SSH key, and each user is jailed to their server's directory.
- **11 server engines**: Paper, Purpur, LeafMC, Folia, Vanilla, Fabric, Forge, NeoForge, Velocity, Waterfall and BungeeCord. Downloads come from mcjarfiles.com, the Paper, Purpur and Fabric APIs, and Mojang's manifest. Reinstalling lets you keep your worlds, configs and plugins.
- **Java manager** for Java 25, 21, 17 and 8. It detects what's installed, installs what's missing, and picks the right version for each Minecraft release. If a server fails to start twice in a row, it stops retrying instead of crash-looping.
- **Modrinth store** for plugins and mods. Installation verifies both Minecraft version and loader, offers compatible release selection, and refuses incompatible releases or existing filenames. Enable/disable switches persist as `.jar` / `.jar.disabled`; restart the server to apply changes.
- **Per-server SQLite databases** with a small SQL studio in the browser, and `.sql` or `.db` export.
- **Backups, ports and players**: `.tar.gz` world backups, port allocations with automatic UFW rules, and a player manager (OP, kick, ban, whitelist). There's also a searchable audit log you can export as CSV or JSON.
- **Schedules** with a simple UTC time picker or validated five-field cron (lists, ranges, steps, month/day names), next-run previews, and persisted execution status/errors. Overlapping runs of the same schedule are rejected.
- **Server sharing** with per-action subuser permissions enforced on server HTTP APIs and the WebSocket console. Owners and administrators retain full server access.
- **MOTD rendering** for legacy `§` and `&` codes, hex colors, Iridium gradients and MiniMessage tags, so your server header and the settings preview look right.
- **Multi-node clustering**. Run a master panel and add worker VPSes with the same binary.
- **Branding and themes**: six built-in themes, custom accent colors and CSS, uploadable logo, favicon, wallpaper and avatars, and a maintenance mode that still lets admins in.
- **Auth and security**: bcrypt passwords, JWT sessions, API keys (`neo_...`), optional Cloudflare Turnstile, and automatic HTTPS through Caddy.
- **Client area link** for connecting an external billing or storefront site (Paymenter, WHMCS, Tebex and similar) as a sidebar page or an iframe.
- **Mods and extensions** through a small JavaScript SDK. More on that below.

## What it doesn't do

- No Docker eggs. Servers run as plain JVM processes, so no SteamCMD, Rust, Python or Discord bot eggs.
- No hosted MySQL or MariaDB. Databases are embedded SQLite.
- No payment processing. You can link your own storefront, but NeoPanel won't handle checkout.

## Install

Tested on Ubuntu 24.04 LTS.

### Master panel

```
curl -sSL https://raw.githubusercontent.com/Pekly/NeoPanel/main/install.sh | sudo bash
```

For a non-interactive install:

```
curl -sSL https://raw.githubusercontent.com/Pekly/NeoPanel/main/install.sh | sudo bash -s -- \
  --install --domain panel.yourdomain.com --email you@yourdomain.com -y
```

The script will:

1. Install the basics (`curl`, `jq`, `tar`, `ufw`, `git`) and Go 1.24+ if needed.
2. Install Java 25, 21, 17 and 8.
3. Build `/usr/local/bin/neopanel` and write `/etc/neopanel/config.yaml`.
4. Optionally set up Caddy with a Let's Encrypt certificate.
5. Configure UFW for ports 22, 80, 443, 2022 and 25565 to 25600.
6. Optionally set up Cloudflare Turnstile.
7. Install a systemd service plus the `neopanel-update` and `neopanel-installer` helpers.

### Adding a node

On the master panel, create an admin API key under Admin Area, Cluster Nodes (or Account, API Keys). You can also do it over SSH:

```
sudo neopanel api-key create --name "eu-node-01"
```

Then run this on the new VPS:

```
curl -sSL https://raw.githubusercontent.com/Pekly/NeoPanel/main/install.sh | sudo bash -s -- \
  --node \
  --master https://panel.yourdomain.com \
  --api-key neo_your_admin_api_key_here \
  --name eu-node-01 \
  --location "Frankfurt, DE"
```

It builds the binary, installs Java, opens the firewall ports, starts the service, and registers itself with your master panel. The node doesn't serve a web UI. Everything shows up in the master.

## Command line

| Task | Command |
|---|---|
| Update from GitHub | `sudo neopanel-update` |
| Installer menu | `sudo neopanel-installer` |
| Show or rotate node token | `sudo neopanel node-token [--rotate] [--raw]` |
| Manage API keys | `sudo neopanel api-key create\|list\|delete ...` |
| Connect a node from the CLI | `sudo neopanel node connect --name n1 --url http://IP:8080 --token neo_node_...` |
| List, test or remove nodes | `sudo neopanel node list\|test\|remove ...` |
| Service status | `systemctl status neopanel` |
| Live logs | `journalctl -u neopanel -f` |

Running the binary directly:

```
neopanel -config /etc/neopanel/config.yaml -host 0.0.0.0 -port 8080 -role panel
neopanel -config /etc/neopanel/config.yaml -role node   # or just: -node
```

In the web UI, `Ctrl + K` opens the command palette and `Ctrl + \`` toggles the console drawer.

## Configuration

The config lives at `/etc/neopanel/config.yaml`:

```yaml
server:
  host: "127.0.0.1"             # 127.0.0.1 behind Caddy, 0.0.0.0 for standalone or nodes
  port: 8080
  mode: "behind_proxy"          # "behind_proxy" or "standalone"
  role: "panel"                 # "panel" (master + web UI) or "node" (daemon only)
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

## Mods and extensions

You can extend the panel without rebuilding it. Mods can add pages, sidebar items, dashboard widgets, command palette actions, and themes.

There are three ways to install one:

- **In the browser.** Go to Admin Area, Mods & Extensions (`/admin/mods`), pick a template, edit the JS and CSS, and hit Save & Apply Live.
- **Upload.** Drop in a `.zip` (with an optional `manifest.json`), a `.js` or `.mjs` file, or a `.css` file.
- **On disk.** Put an unpacked mod in `/var/lib/neopanel/mods/<mod-id>/`:

```
my-custom-mod/
├── manifest.json   # { id, name, version, author, description, enabled, entry_js, entry_css }
├── mod.js          # runs on panel load, with window.NP available
└── mod.css         # attached while the mod is enabled
```

Each mod runs in its own try/catch, so a broken one shouldn't take the panel down. If something goes wrong, add `?safemode=1` to any URL (like `/admin/mods?safemode=1`) to skip mod execution while you fix it.

### The `window.NP` API

| Method | What it does |
|---|---|
| `NP.registerPage({ id, title, icon, scope, render })` | Adds a page at `/mod/<id>` and a sidebar entry. `scope` is `'global'`, `'server'` or `'both'`. |
| `NP.addSidebarItem({ id, label, icon, scope, pageId, url, onClick })` | Adds a sidebar item. |
| `NP.registerSlot(slotName, (el, ctx) => ...)` | Renders into built-in slots: `nav:bottom`, `dashboard:top`, `dashboard:bottom`, `console:below-stats`, `console:below-charts`, `files:top`, `plugins:top`, `settings:bottom`. |
| `NP.addPaletteCommand({ id, group, label, subtitle, action })` | Adds an action to the `Ctrl + K` palette. |
| `NP.on(event, fn)` | Listens for `'render'`, `'route'` or `'server:switch'`. |
| `NP.api(url, opts)` / `NP.srvApi(subPath, opts)` | Calls the panel or active-server API with the current session. |
| `NP.injectCSS(id, css)` / `NP.removeCSS(id)` | Adds or removes a stylesheet at runtime. |
| `NP.storage(modId)` | A namespaced `localStorage` wrapper with `get`, `set` and `remove`. |
| `NP.navigate(target)` / `NP.toast(msg, kind)` / `NP.modal` | Navigation, notifications, and alert/confirm/prompt dialogs. |

A small example, adding a "Save World" button below the console charts:

```js
NP.registerSlot('console:below-charts', (el, ctx) => {
  if (!ctx.server) return;
  el.innerHTML = `<div class="card"><div class="row sb">
    <b>Server Macros (${ctx.server.name})</b>
    <button class="btn sm pr" id="macro-save">Save World</button>
  </div></div>`;
  el.querySelector('#macro-save').onclick = () =>
    NP.srvApi('/command', { method: 'POST', body: JSON.stringify({ command: 'save-all' }) })
      .then(() => NP.toast('Saved world!', 'ok'));
});
```

## Client area and maintenance mode

**Client area.** Under Admin Area, Panel Look & Branding, Client Area Storefront Portal, you can set a sidebar label, point it at your external storefront URL, and choose whether it opens as a card with a new tab or embeds as an iframe. Turn it off and visitors get a 503 page. Leave it unconfigured and they get a "not configured" page.

**Maintenance mode.** You can enable it from the same branding page, with your own title, message and optional ETA. Regular users see a 503 lock screen, and admins can still sign in and get a banner reminding them it's on.

Server pages use short routes like `/server/{id}/console` (with `?node={nodeId}` for remote nodes). Unknown routes and missing servers show a proper 404.

## HTTP and WebSocket API

For scripts and automation, authenticate with `Authorization: Bearer neo_<api_key>` or a JWT from login.

- **Cluster and nodes:** `GET /api/cluster/servers`, `GET|POST /api/nodes`, `ALL /api/nodes/{id}/proxy/*`, `WS /api/nodes/{id}/ws/{serverId}`
- **Servers and power:** `GET|POST /api/servers`, `GET|PUT|DELETE /api/servers/{id}`, `POST /api/servers/{id}/power`, `POST /api/servers/{id}/command`
- **Console:** `WS /api/servers/{id}/ws`
- **Files:** `GET|DELETE /api/servers/{id}/files`, plus `/content`, `/rename`, `/move`, `/compress`, `/unzip`, `/upload` and `/download` under `/api/servers/{id}/files/`
- **Software and Java:** `GET /api/versions`, `POST /api/servers/{id}/software`, `POST /api/servers/{id}/reinstall`, `GET /api/system/java`, `POST /api/system/java/install`
- **Plugins, backups, databases:** `/api/plugins/search`, `/api/servers/{id}/plugins*`, `/api/servers/{id}/backups*`, `/api/servers/{id}/databases*`
- **Branding:** `GET /api/branding`, `GET|PUT /api/admin/branding`, `POST /api/admin/branding/upload`, `POST /api/admin/branding/reset`, `GET /branding/{filename}`
- **Mods:** `GET /api/mods`, `GET /mods/{id}/{path...}`, `GET|POST /api/admin/mods`, `POST /api/admin/mods/upload`, `PUT|DELETE /api/admin/mods/{id}`

## Building from source

You need Go 1.22 or newer. There's no Node.js and no bundler, since everything in `web/` is embedded with `//go:embed`.

```
git clone https://github.com/Pekly/NeoPanel.git
cd NeoPanel

make build         # builds ./bin/neopanel
make build-linux   # static Linux binaries for amd64 and arm64
make run           # run in dev mode
```

## License

MIT. Made for Minecraft admins who would rather spend their RAM on players than on PHP workers.

## Regression checks

Run `go test ./...` for backend safeguards and plugin compatibility tests, and `node --test web/tests/qol.test.cjs` for editor, upload queue, plugin toggle, and cron-preview state tests. JavaScript syntax can be checked with `node --check web/static/js/app.js` and `node --check web/static/js/qol.js`.

Backups are published only after archive creation and closure succeed, and existing archive names are not overwritten. These are file archives, not application-consistent live snapshots: stop the Minecraft server before backing up when a fully consistent world snapshot is required. Browser upload retries restart the transfer; cancellation cannot retract a file the server has already received.

When deploying this update, update the panel and all worker nodes. Non-admin remote access fails closed on older workers that do not enforce forwarded user permissions, or when the user has not been synchronized to the worker. Cron aliases such as `@daily` must be converted to five fields (for example `0 0 * * *`).
