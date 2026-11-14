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
- **File dragging and Trash:** drag files or a selection onto a folder; hold over folders or breadcrumbs to browse while dragging. Moves do not overwrite existing items. File-manager deletions default to a recoverable 24-hour Trash bin, with an explicit permanent-delete checkbox. See [Trash storage and deployment notes](internal/trash/README.md). SFTP and plugin-uninstall deletions retain their existing behavior.
- **Built-in SFTP** on port 2022. Log in with your panel credentials or an SSH key, and each user is jailed to their server's directory.
- **12 server engines**: Paper, Purpur, LeafMC, Folia, Vanilla, Fabric, Forge, NeoForge, Velocity, Waterfall, BungeeCord and Steel. Java downloads come from mcjarfiles.com, the Paper, Purpur and Fabric APIs, and Mojang's manifest; Steel uses official GitHub native releases. Reinstalling lets you keep your worlds, configs and plugins.
- **Java manager** for Java 25, 21, 17 and 8. It detects what's installed, installs what's missing, and picks the right version for each Minecraft release. If a server fails to start twice in a row, it stops retrying instead of crash-looping.
- **Modrinth store** for plugins and mods. Installation verifies both Minecraft version and loader, offers compatible release selection, and refuses incompatible releases or existing filenames. Enable/disable switches persist as `.jar` / `.jar.disabled`; restart the server to apply changes.
- **Per-server SQLite databases** with a small SQL studio in the browser, and `.sql` or `.db` export.
- **Backups, ports and players**: `.tar.gz` world backups, port allocations with automatic UFW rules, and a player manager (OP, kick, ban, whitelist). There's also a searchable audit log you can export as CSV or JSON.
- **Schedules** with a UTC time picker or validated five-field cron, next-run previews, and persisted execution results. Build up to 20 ordered command/backup/power steps, add delays before steps, and optionally continue after a failure. Legacy single-action schedules still work; overlapping runs are rejected.
- **Team & invitations** in a dedicated server tab. Add existing accounts or invite a username and email to register through a single-use, 72-hour link. Pending invitations, resend/revoke controls, and delivery failures are visible. Subuser permissions are enforced on server HTTP APIs and the WebSocket console; owners and administrators retain full access.
- **MOTD rendering** for legacy `§` and `&` codes, hex colors, Iridium gradients and MiniMessage tags, so your server header and the settings preview look right.
- **Multi-node clustering**. Run a master panel and add worker VPSes with the same binary.
- **Branding and themes**: six built-in themes, custom accent colors and CSS, uploadable logo, favicon, wallpaper and avatars, and a maintenance mode that still lets admins in.
- **Auth and security**: bcrypt passwords, JWT sessions, API keys (`neo_...`), optional Cloudflare Turnstile, passkey registration/sign-in through WebAuthn, and automatic HTTPS through Caddy. User-verified passkeys complete sign-in without a separate 2FA code; password sign-in still requires configured 2FA.
- **Client area link** for connecting an external billing or storefront site (Paymenter, WHMCS, Tebex and similar) as a sidebar page or an iframe.
- **Mods and extensions** through a small JavaScript SDK. More on that below.

## What it doesn't do

- No Docker eggs. Java servers run as JVM processes and Steel runs as a native executable; arbitrary SteamCMD, Rust game-server, Python or Discord bot eggs are not supported.
- No hosted MySQL or MariaDB. Databases are embedded SQLite.
- No payment processing. You can link your own storefront, but NeoPanel won't handle checkout.

## Install

Before upgrading an existing installation, read the [security upgrade notes](docs/security-upgrade.md), especially public URL and trusted-proxy configuration. Rebuild and update both the panel and workers.

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

1. Install the basics (`curl`, `jq`, `tar`, `ufw`, `git`) and Go 1.26+ if needed, verifying official compiler downloads with SHA-256.
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

### Steel (native Rust server)

Choose **Steel** when creating a server, or from Software / Reinstall. Select a Steel release tag (for example `v0.17.0+mc26.3`) or `latest`, not a Minecraft version. NeoPanel downloads the official executable for the **server node's** platform: Linux x64 (preferring the static musl build), Windows x64, or macOS ARM64. Other platforms require a self-built executable and automatic download disabled. See [Steel installation](https://steelmc.dev/getting-started/installation/).

Steel starts directly without Java, JVM flags, `nogui`, or Java heap arguments. The existing `jar_file` API field stores `steel` (`steel.exe` on Windows). Its full default `config/config.toml` is downloaded from the selected release; NeoPanel synchronizes `[server].server_port` with the assigned port before each launch. Use Files to edit TOML settings. For manual installs, upload the matching release configuration as well as the executable. macOS may require the development-mode setup described by Steel.

Reinstall preservation includes Steel's `saves/` and `config/` directories. Back up before changing between Steel and Java engines: their plugins, mods, configuration and world layouts are not interchangeable. Java quick configuration and the Modrinth JAR catalog are unavailable for Steel. Memory maximum is enforced by systemd when that isolation backend is available; unrestricted processes and bubblewrap alone do not impose a native memory cap.

Steel logo supplied by [SteelMC](https://steelmc.dev/), credited upstream to colonthreeing.

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

`Ctrl/Cmd + F` (or the editor's **Find** button) opens find **only inside an open text file**. Use Enter / Shift+Enter (or F3 / Shift+F3) to move between matches and Escape to close it. Search runs locally in a worker, scans up to the editor's 5 MiB text limit, and caches up to 1,000 matches. If workers are unavailable, a labelled fallback searches the first 128 Ki characters. Outside the editor, the browser's own find shortcut is untouched. **Files → Reload** refreshes the current folder without discarding its filter/selection.

Text-editor syntax colors are chosen by extension: JSON/JSONC/`.mcmeta`, YAML, TOML, INI/config, Properties/`.lang`, JavaScript/TypeScript, Java, Python, shell/PowerShell, SQL, XML/HTML, CSS, Markdown, Go, Rust, Lua, and Dockerfile. Pinned local Prism assets run only in a worker; text, caret, undo, save, and draft behavior remain native. Files above 200,000 characters, unsupported formats, or coloring timeouts remain editable as plain text with a status explanation.

Page and subtab entrances play only on the first visit in a browser-tab session (including across reloads); signing out clears that history. Browsing folders or returning from the editor does not replay the Files page entrance. Layout-switch and action feedback remain responsive, with reduced-motion preferences respected.

## Configuration

The config lives at `/etc/neopanel/config.yaml`:

```yaml
server:
  host: "127.0.0.1"             # 127.0.0.1 behind Caddy, 0.0.0.0 for standalone or nodes
  port: 8080
  mode: "behind_proxy"          # "behind_proxy" or "standalone"
  role: "panel"                 # "panel" (master + web UI) or "node" (daemon only)
  public_url: "https://panel.example.com" # fixed origin for account emails/passkeys
  startup_timeout_seconds: 180

paths:
  data_dir: "/var/lib/neopanel"
  servers_dir: "/var/lib/neopanel/servers"
  db_path: "/var/lib/neopanel/neopanel.db"

security:
  jwt_secret: "<auto-generated-hex>"
  token_expiration_hours: 72
  node_token: "neo_node_<auto-generated-hex>"
  trusted_proxies: ["127.0.0.1/32", "::1/128"] # only for local reverse proxy; otherwise []
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

You need Go 1.26 or newer for the patched security dependencies and rooted filesystem APIs. There's no Node.js build step and no bundler, since everything in `web/` is embedded with `//go:embed`.

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

Run `go test ./...` for backend safeguards, SMTP invitation flows, WebAuthn security checks, restore, and schedule-chain tests. Run `node --test web/tests/*.test.cjs` for frontend state, validation, navigation, passkey, and panel-address regressions. JavaScript syntax can be checked with `node --check` on `web/static/js/app.js`, `qol.js`, `team.js`, `automation.js`, and `passkeys.js`.

Backups are published only after archive creation and closure succeed, and existing archive names are not overwritten. These are file archives, not application-consistent live snapshots: stop the Minecraft server before backing up when a fully consistent world snapshot is required. Browser upload retries restart the transfer; cancellation cannot retract a file the server has already received.

When deploying this update, update the panel and all worker nodes. Non-admin remote access fails closed on older workers that do not enforce forwarded user permissions, or when the user has not been synchronized to the worker. Cron aliases such as `@daily` must be converted to five fields (for example `0 0 * * *`).

## Team invitations, restore, and schedule chains

- **Invitations:** configure and test SMTP under Admin → Email & SMTP on the main panel. In a server’s **Team** tab, choose **Invite by email**, enter the requested username/email, and select permissions. The recipient follows the email link, sets a password, and signs in normally. Existing accounts must sign in as the matching email identity; invitation links never reset an existing password or bypass 2FA. Resending rotates the link and has a one-minute cooldown. Tokens are stored only as hashes.
- **Remote servers:** update the panel and worker nodes together and synchronize the inviter’s account. Registration and email are master-owned. Offline/old workers fail explicitly instead of granting partial access silently. If a remote grant was interrupted, the recipient can retry the same valid link; any account already created by that invite keeps its original password. Revoking an unresolved remote grant requires the worker to be reachable so the grant cannot reappear on a delayed retry.
- **Public invite URL:** set the browser-facing panel origin under **Admin → Appearance → Panel address**. It is used for invitation and password-reset links as well as passkeys, independently of the game/SFTP FQDN. The database setting overrides the `server.public_url` configuration fallback.
- **Themed emails:** invitations, password resets, and SMTP tests share one email-safe layout, with a 24 px logo **before** the 18 px panel name. Saved global Appearance colors (background, card, header, text, accent, and light/dark/OLED default) are applied; `auto` uses dark defaults, since recipient browser preferences are unavailable. Custom browser CSS, blur, and transparency are not copied. Uploaded/relative logos use the configured public panel URL. Prefer a publicly reachable PNG/JPEG; clients may block remote images or override colors. Logos are linked, not fetched/attached by the daemon.
- **Password recovery:** use **Reset Password** on the sign-in screen. The email supplies a one-hour, single-use 64-character code/link. The form handles request failures, validates matching passwords and the 72-byte bcrypt limit, and returns to sign-in after success. Reset tokens are never returned by the public request API or saved by the browser; link tokens are removed from the address bar after opening.
- **Invitation account cleanup:** on local-only installations, removing the last team membership or deleting the last associated server also deletes an account **created through an invitation**, along with its credentials. Existing independent accounts, administrators, protected panel roles, and accounts with other ownership/memberships are preserved. Shared/exported identities and accounts on panels with remote nodes are retained rather than guessing global orphanhood. See [account lifecycle and remote reconciliation](docs/invitation-account-lifecycle.md) before upgrading; ambiguous legacy provenance is not assumed.
- **Restore:** stop the server, open **Backups**, and choose **Restore**. Confirming performs a full replacement: files not present in the archive disappear from the active directory. Archives are validated/extracted into staging before the swap; traversal, links/devices, truncation, and extraction-budget violations are rejected. Concurrent power/reinstall/deletion operations are excluded. The previous server directory is retained as a safety copy, with its relative path shown after success; verify the restored server before manually removing it to reclaim disk space. Two directory renames with rollback are used, not a crash-atomic filesystem exchange. Do not modify files through another process during restoration.
- **Multi-step schedules:** add/reorder up to 20 steps. Each delay is 0–3600 seconds before that step. Actions are console command, backup, start, stop, and restart. A failure stops the sequence unless that step has **Continue to the next step if this one fails** enabled; any failure still marks the overall run failed. Edits affect subsequent runs, not a sequence already executing. Console-command success means successful delivery to stdin, not game-level acknowledgment.

## Passkeys and invitation delivery

To enable passkeys and choose the public origin for invitation and password-reset links, open **Admin → Appearance → Panel address** and save an address such as `https://panel.example.com`. The normalized address is persisted in SQLite and takes effect immediately without a restart. This database override takes precedence over `server.public_url`; saving does not edit YAML, and the address cannot be cleared in the UI/API.

For initial deployment, `public_url` under the existing `server` section remains a fallback when no database override is set (restart after changing YAML):

```yaml
server:
  public_url: "https://panel.example.com"
```

- Use the exact browser-facing panel origin, without a path, credentials, query, or fragment, and preserve that Host through your reverse proxy. HTTPS is required in production; an explicitly configured `http://localhost:PORT` is supported for development. Saving does not provision DNS records, HTTPS certificates, or reverse-proxy routing: configure and verify these first. The application does not choose its passkey identity from untrusted request headers. Without either a database address or `server.public_url`, password sign-in continues to work but passkeys remain unavailable.
- Open **Account → Security → Add passkey**, provide a name and your current password, then approve the request on your device or FIDO2 security key. Up to ten passkeys can be registered. NeoPanel stores public credential data, not private keys. Removal also requires your current password.
- Choose **Sign in with a passkey** on the sign-in screen. A valid passkey assertion with user verification completes sign-in without a separate authenticator/recovery code. This does not disable the account's 2FA setting: password sign-in still requires it, and Turnstile checks remain in effect. Keep the panel hostname stable: passkeys are bound to it and the RP hostname is pinned after first registration. Before migrating to another hostname, remove **all passkeys from every account**, change the panel address, then register new passkeys at the new address. Saving a different hostname is blocked while any passkeys exist. `security.passkeys_disabled: true` disables ceremonies without deleting registered keys and does not bypass this safeguard.
- Team invitations reject conflicting username/email identities and people who already have access to the target server. A matching existing account may still be invited to another server and must sign in as that account. If someone is added manually while their invite is pending, accepting the old invitation cannot overwrite the membership or its permissions.
- Invitation emails contain a branded HTML card, readable permission names, an acceptance button, explicit expiry, and a plain-text alternative. SMTP settings are still required; delivery errors are reported rather than labeled successful. Update the panel and workers together for remote invitation-overlap checks.

Passkey tests include signed assertions and replay/origin/user-verification checks. Browser integration was exercised with Chromium’s virtual authenticator against a real temporary panel, including direct passkey sign-in for a 2FA-enabled account, continued TOTP enforcement for password sign-in, removal, and password fallback. Invitation delivery tests use a local SMTP receiver; delivery through your own email provider and behavior on physical authenticators depend on your deployment.

### Process ownership, isolation, and unavailable server pages

- **Stop/restart:** Stop waits for the owned run to exit; Restart refuses to proceed if shutdown or cleanup failed. Linux launches use a dedicated process group and unique systemd scope when available. Startup timeouts and forced stops clean up that owned run instead of killing all Java processes.
- **One owner per directory:** Panel processes share an OS lock under the server root's `.neopanel-locks` directory, outside each server's writable files. Start, reinstall/deletion maintenance, and restore refuse conflicting ownership. Do not delete lock files to bypass a running instance.
- **Isolation is explicit:** `systemd` provides resource limits, **not filesystem isolation**. `auto` prefers Bubblewrap but reports a clear warning label if only resource limits or an unrestricted process are available. Choose `bwrap` in the node's isolation setting to require the filesystem sandbox; missing Bubblewrap now fails the launch instead of silently downgrading. Install the `bubblewrap` package on that Linux host and ensure its user namespaces are permitted. Changes apply on the next start. The sandbox still shares the host network for Minecraft connectivity; it is not a VM or a separate network namespace.
- **Sandbox visibility:** Only the server directory is writable, alongside private temporary/device filesystems. Host `/etc` and `/opt` are not mounted wholesale; Java runtime, DNS, and public certificate/configuration files are mounted as needed, and inherited host environment variables are cleared. Custom plugins requiring extra host paths or environment variables may need configuration changes rather than disabling isolation.
- **Existing world lock:** `world/session.lock: already locked` means Minecraft could not acquire its world lock. On the host, use `sudo lsof /var/lib/neopanel/servers/<server-id>/world/session.lock` (or `sudo fuser -v ...`) to identify the owner, then stop that specific instance through its owning service/panel. Do not delete `session.lock` or use a blanket Java kill. New Linux starts check the configured world's locks before launching; already-orphaned processes from older versions are not automatically adopted or killed.
- **Server-page recovery:** Incomplete inventory, network errors, and offline remote nodes show a recoverable unavailable page, not a false 404. The requested URL is retained, two automatic retries are bounded by network timeouts, and manual Retry remains available. A complete authorized inventory can still return a genuine missing-server 404.

Regression checks: `go test ./...`, `go vet ./...`, and `node --test web/tests/*.test.cjs`. Linux-specific lock/process-group tests require execution on Linux; cross-compilation alone does not validate host systemd or Bubblewrap behavior.

### Installer isolation preflight

Fresh **panel and node installs**, and **updates of existing installations**, now install missing `bubblewrap` and `timeout` tools and run a bounded sandbox probe before building/replacing NeoPanel or requesting a service restart. Existing tools are reused; apt, dnf, yum, pacman, and apk are supported. Package failures and unsupported package managers are no longer silently ignored in this prerequisite path.

The probe tests namespace creation, a writable sandbox directory, hidden host configuration/canary files, and clearing inherited environment variables. Updates inspect the installed service's user/group and test under that identity; they do not assume root for a custom service user. Missing identity-launch tools, unreadable units, and `DynamicUser=yes` stop with an explanation rather than report a false success.

**A failed probe blocks the install/update.** Check `/tmp/neopanel-install.log` for the actual Bubblewrap error. A package can be installed while AppArmor, kernel user-namespace policy, or the VPS/container provider still prevents sandbox creation. Resolve the host restriction or package issue and rerun; the installer does not disable AppArmor, relax sysctls, or automatically switch to an unrestricted runtime. Probe temporary files are removed on success and failure.

Passing this probe means the tested host/identity supports the basic sandbox, not that every systemd service restriction or Java/plugin workload has been tested. The installer preserves node isolation settings: select **Bubblewrap (`bwrap`)** on the node to require filesystem isolation. Systemd-only mode remains resource limiting, not a filesystem sandbox.

Run `bash -n install.sh` and `bash tests/install-isolation.test.sh` to validate installer syntax and the mocked prerequisite/probe/update regressions. These tests do not install packages, invoke real systemd, or establish that Bubblewrap works on your production host.

### Team access, panel roles, and community settings

- **SFTP is opt-in for invited members.** Grant `sftp.access` explicitly. SFTP then also enforces `files.view` for reads/listing, `files.edit` for writes, and `files.delete` for removal. Owners and built-in administrators retain full server access. Existing file permissions alone do not grant SFTP. Authentication and subsequent requests/open handles recheck current identity and permissions, so revocation applies to established sessions.
- **Every SFTP login selects a server jail.** There is no shared servers-directory fallback, even for administrators. An account with exactly one permitted server may use its username; otherwise use `username#server-id` or the server-specific username shown by the panel. Ambiguous server aliases are rejected; shell/exec/forwarding and symlink creation are not supported.
- **Team managers cannot grant permissions they do not hold.** This applies to member edits and invitation creation/resend/acceptance. Server owners and built-in administrators can grant the full permission catalog.
- **Ordinary members cannot create servers by default**, including accounts that already own a server. Administrators can create named custom panel roles and grant the global `servers.create` capability, then assign a role to a user. Panel roles do not confer administrator privileges or access to other servers. Assigned roles must be unassigned before deletion. Role changes, admin demotion, and deleted users are resolved from current database state rather than stale session roles.
- **Update panel and workers together.** User proxy requests require the current permission-policy version; an older worker is refused instead of being allowed to apply weaker rules. The master explicitly authorizes remote server creation; a synchronized worker account alone is not a creation grant.
- **Community settings** control a custom Discord join card and a plain-text announcement bar. Supply a Discord server ID and enable its public Server Widget in Discord so the panel can obtain the invite. NeoPanel does not embed Discord's native widget, invent an invite from the numeric ID, or expose the widget's member list. Failed lookups show unavailable status. Announcement changes get a new dismissal identity.

The backend checks remain authoritative: gray/disabled navigation is a usability aid, not the access-control boundary. Real loopback SSH/SFTP tests and API regressions cover denied/allowed access, live revocation, and role changes; host-specific filesystem confinement still requires Linux runtime verification in addition to cross-compilation.

**Security upgrade note:** Older provisioning code samples could include the configured node token in a response available to authenticated members. The examples endpoint is now admin-only and uses a placeholder, never the node token. If members could access an older deployment, rotate its node token(s), update the corresponding cluster credentials, and review access logs after upgrading. This change does not rotate credentials automatically because that would disconnect existing workers. SFTP rename/overwrite now requires both `files.edit` and `files.delete`.
