# 1-Click Modpack Installers

NeoPanel includes native support for Modrinth modpack (`.mrpack`) installation and management.

## How it works

1. **Browse & Search:** Browse popular Modrinth modpacks directly in **Server → Plugins & Mods → 1-Click Modpacks**.
2. **Server-Safe Mod Filtering:** Parses `modrinth.index.json` inside the `.mrpack` archive. Mods with `env.server == "unsupported"` (such as Sodium, Iris, client-only HUDs, and shaders) are automatically omitted to prevent server crashes on boot.
3. **Override Extraction:** Extracts server configuration overrides, datapacks, and custom scripts from `overrides/` and `server-overrides/`.
4. **Automatic Loader Alignment:** Detects the required loader (Fabric, Forge, NeoForge, or Quilt) and Minecraft version from modpack dependencies, automatically aligning the server configuration.
5. **Hash Verification:** Validates each downloaded mod JAR against its published SHA-512 and SHA-1 checksums before placement.

## API Endpoints

- `GET /api/modpacks/search`: Search Modrinth modpack catalog.
- `GET /api/servers/{id}/modpack`: Retrieve currently installed modpack metadata (`.neopanel_modpack.json`).
- `POST /api/servers/{id}/modpack/install`: Download, parse, and install a modpack archive into the server directory.
- `DELETE /api/servers/{id}/modpack`: Unlink modpack metadata.
