# Proxy settings

Velocity, BungeeCord and Waterfall now have their own editor under **Server → Settings → Proxy settings**. Java, memory and startup controls remain available; world options such as gamemode, difficulty and PvP do not apply to proxies.

## First setup

1. Create a server using the proxy software you want. NeoPanel creates `velocity.toml` for Velocity or `config.yml` for BungeeCord/Waterfall with the assigned port. Proxy downloads do not create `server.properties` or `eula.txt`.
2. Keep the proxy stopped while configuring it. For an older install with no config, use **Create config** in Proxy settings. Existing files are not replaced.
3. Add your backend servers, for example `lobby` and `survival`, using addresses reachable **from the proxy**. An address on another machine needs that machine's reachable host/IP—not `localhost`.
4. Choose the join/fallback order. The first available backend in that order receives players unless a domain route selects another server.
5. Set the MOTD and advertised player count. The displayed count is not a connection cap.
6. Save, configure forwarding on the backends, and start the proxy manually. Saving does not restart or reload it.

Starter configs have **no invented backend addresses**. Add at least one backend and one fallback entry before starting; an empty BungeeCord/Waterfall network cannot start successfully. The panel checks configuration structure, not whether backend servers are reachable.

## Velocity forwarding

New Velocity configs use authenticated player logins and **Modern** forwarding, with a unique random `forwarding.secret`. Existing configs and existing secrets are preserved. Viewing the config does not reveal the secret; use the explicit **Show forwarding secret** action when setting up the backend.

For current Paper backends:

- Enable Velocity support in `config/paper-global.yml` under `proxies.velocity`.
- Set its `secret` to the proxy's forwarding secret and match its `online-mode` to the proxy's login mode.
- Restrict direct backend access to the proxy before setting the backend's `server.properties` `online-mode=false`. Keep the public proxy's player authentication enabled.
- Restart the backend after changing its forwarding settings.

Other server implementations may require a forwarding mod or different settings. Use that implementation's documentation. Modern forwarding requires compatible Minecraft 1.13+ backends; older networks may need Legacy or BungeeGuard. Legacy forwarding alone does not authenticate the proxy-to-backend connection, so protecting direct backend access is essential.

Disabling player authentication on the public proxy requires an explicit confirmation. Do not disable it simply because backend servers use offline mode behind the proxy.

## BungeeCord and Waterfall

The editor uses the primary listener in `config.yml` for the MOTD, displayed slots, fallback order and hostname routes. IP forwarding, default-server behavior and ping passthrough are available too. A domain route selects one backend for these proxy types.

For Spigot/Paper backends using legacy Bungee forwarding, configure the matching backend support and protect direct backend access. NeoPanel does not make those changes automatically.

Edit advanced multi-listener configs in **Files**; the builder does not flatten them into a single listener. Startup port synchronization changes the primary listener only and preserves additional listeners, whose allocations remain the operator's responsibility. YAML aliases, anchors, merges and custom tags are rejected by config handling, including startup port synchronization. Convert those constructs to plain values before starting; unsupported files are never silently reset.

## Existing files and reinstalls

- Installation and **Create config** only create missing files. They do not reset a customized config.
- Builder saves preserve unknown setting values, sections and extra fields on retained backend entries. Saving or changing a listening/query port can reformat the file and remove comments; already-correct port fields leave the file unchanged. Keep a backup if exact layout/comments matter, or use Files for manual editing.
- A changed revision is rejected. Refresh before saving after edits made through Files, another tab, or another client; the builder does not silently overwrite a stale draft.
- The assigned listening port comes from NeoPanel's server allocation. Change it through the existing permitted network controls, not a backend-address field.
- **Preserve config** during reinstall also preserves proxy configs and Velocity's configured relative secret path. If that path is nested, its top-level directory is retained. A destructive reinstall without config preservation can create a new forwarding secret, which must be copied to the backends again.
- Invalid, oversized or unsafe files are rejected, not automatically replaced. Config reads are limited to 1 MiB, secret reads to 4 KiB. Symlinks, special files and paths outside the server directory are not accepted.

## Access and limits

Reading the editor requires `settings.view`; saving and initialization require `settings.edit`. Showing the forwarding secret additionally requires `files.view`. Owners and admins retain their normal access. Worker requests use the same permissions as local requests. Host-managed port, Java and memory restrictions still apply.

The builder supports up to 128 backend servers and 128 domain routes, with bounded references and request sizes. Backend names must be unique (ignoring case), use letters, digits, `_` or `-`, and cannot be `try`. Routes and fallback entries must refer to configured servers.

Update the panel and workers together. This feature does not alter DNS, open firewall ports, connect to or reconfigure backends, or automatically start/restart games.
