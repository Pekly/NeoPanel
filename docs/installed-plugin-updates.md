# Installed plugin identification and bulk updates

Open **Server → Plugins → Installed** (or **Mods → Installed** on a modded server) and choose **Check for updates**. Update the panel and workers together. Discovery uses `files.view`; source confirmation and updating require `files.edit`, and the Plugins feature policy still applies.

## How identification works

1. NeoPanel reads eligible JARs from the server's plugin/mod folder and calculates SHA-512 locally.
2. Modrinth's batch hash lookup identifies exact published files, including manually uploaded or renamed copies. NeoPanel then retrieves the public project information and the latest release for the server's actual software.
3. If the hash is not recognized, NeoPanel reads bounded embedded metadata, where supported, and searches using the plugin name or cleaned filename. It does not execute the JAR or extract it into the server's files.
4. A name search produces **possible matches**, not an automatic trusted association. Review the project and explicitly confirm its source before replacement is allowed.
5. Files without a useful match remain **Unknown**. Provider failures are shown as lookup errors, not interpreted as proof that a plugin is unknown or current.

The panel sends hashes and plugin search names to Modrinth, not JAR contents, configuration files, or credentials. A hash match identifies a provider release; it is not a malware scan or a runtime-compatibility guarantee. Projects not published on Modrinth cannot be reliably updated through this provider.

### Confirming a possible match

The Installed tab distinguishes an exact hash match from a user-confirmed source. An embedded name/version can help identify a file but is author-controlled and does not establish its origin.

Confirming a source means: **replace this file with a release from this project when I approve an update**. A custom or patched JAR may lose its modifications. The association is tied to the current server, filename and file hash; replacing or renaming the file requires discovery again. An unverified installed version must not be presented as an exact Modrinth release.

## Updating selected plugins

1. Scan the installed files and inspect each identified project, installed version and proposed target.
2. Stop the server yourself. NeoPanel does not stop it automatically.
3. Select up to ten eligible items and choose **Review selected updates**. Only exact-hash or explicitly confirmed sources may be updated; suggestions and unknown files are not selected automatically.
4. Confirm the specific files and target releases. The backend rechecks file hashes and source associations, validates the target project/software, and excludes conflicting starts, restores and managed file writes while replacing files.
5. Review the individual results. Successful replacements preserve the original filename and enabled/disabled state. Plugin configuration files are not changed. The old JARs are retained in a separate backup directory outside the active plugin/mod folder; results provide relative backup paths.
6. Start the server manually when ready and inspect its startup log. No automatic dependency installation or game restart occurs.

The default proposed target follows the server **software**, not an exact Minecraft-version tag. Latest can require a different Minecraft version, Java version, dependency or configuration. Review the author's advertised compatibility before approving. An exactly identified newer release is not automatically downgraded to an older eligible release.

## Failures and recovery

Bulk updating reports each file separately; it is not an all-or-nothing transaction for the whole selection. A failed download must not remove the old JAR. A file changed since discovery is rejected rather than silently overwritten. Existing backups are not overwritten.

If the browser loses the response, do not assume the request failed before changing files. NeoPanel blocks repeated submissions until a successful fresh scan reconciles the installed files. It does not automatically repeat an uncertain update. Read the results and inspect the server files/backups before proceeding.

To recover an old JAR, keep the server stopped, preserve the unwanted replacement separately, and restore the matching backup at its original filename and enabled/disabled state. Do not leave two enabled copies of the same plugin in the active folder.

If publication and rollback both fail, NeoPanel blocks startup pending operator repair. This recovery gate is process-local: restarting the panel is not a repair. Restore the retained original before restarting the panel or starting the game server.

## Bounds and limitations

Discovery scans 20 files per page, with at most 1,000 records retained in the UI. Each file is limited to 200 MiB and each scan page or download batch to 512 MiB. Scans have a 60-second deadline; bulk updates have a 120-second deadline. Descriptors are capped at 128 KiB and archives at 10,000 entries with a 4 MiB central-directory limit. ZIP64 and prepended archives are conservatively rejected. Supported descriptors include Bukkit/Paper/Bungee YAML, Velocity/Fabric JSON, and Forge/NeoForge TOML. Files are never extracted or executed during inspection.
Downloads require a positive declared size and upstream SHA-512 or SHA-1 verification. Old JARs remain under `plugin-backups/<unique-batch>/` outside the active folder. Discovery is explicitly requested rather than repeated on each UI refresh.

Application file-operation leases coordinate NeoPanel's managed writes. They do not prevent an external process or another daemon from changing the same directory. Backups consume disk space; they are not a replacement for host disk/inode quotas or a complete server backup.
