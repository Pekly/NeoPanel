# Built-in world conversion (Chunker)

World conversion is a stopped-server tool, not a game launch. It never automatically stops, starts, restarts, replaces, or changes the software/version of a server. Conversion can lose entities, structures, custom dimensions and plugin/mod data. Keep independent backups and inspect exported worlds before using them.

Open it from the server's root **Files → World tools → Convert world**. The entry also provides access to existing-job status and cancellation when new conversion is unavailable; the controls inside still enforce permissions and host support.

## Supported hosts and resource policy

- **Linux only for execution:** unified cgroup v2, usable systemd scopes, Bubblewrap with `--size`, `/bin/sh`, `cp`, and `tar`. Missing or unusable enforcement fails closed, regardless of the game's insecure-fallback setting. Merely finding the tools does not establish enforcement: a fixed bootstrap checks the actual `memory.max` and `cpu.max` of its scope before executing the converter.
- **Windows and other hosts:** status, catalog and permission checks work, but execution explicitly refuses. Existing Windows Job ownership enforces memory but not CPU; this integration does not pretend otherwise. No Windows Job quota code was changed.
- Positive assigned CPU quota and at least 512 MiB assigned RAM are required. `CPULimitPercent=100` means one core, consistent with systemd. Memory is a **strict total cgroup ceiling equal to MemoryMaxMB**, not the game's Java heap-plus-headroom policy. JVM heap is half that budget. TasksMax is 128 and swap allowance is zero.
- The converter works entirely inside a size-bounded tmpfs with no network namespace sharing. Both its writable input copy and output consume the assigned memory ceiling. Consequently, worlds close to or larger than the RAM allocation may fail with OOM even when they fit the disk limit. This deliberately conservative implementation does not silently substitute an unbounded disk-backed output directory.
- The host mounts only the *private* source clone and runtime/tool files read-only. Neither the original server, backups, credentials nor the host output directory is a writable converter mount. Java logs go to a bounded log sink; a fixed tar exporter streams output through a pipe to bounded regular-file-only extraction in the daemon. Converter disk writes cannot fill host storage.
- Resource ceilings apply to the owned converter process tree, including its in-sandbox copy/export helpers. The daemon's bounded cloning, hashing, downloading, extraction and backup phases do not run in that cgroup. Do not interpret this as a CPU/RAM quota on all daemon work.

## Input, bounds and output

Choose one direct child world directory of the server, for example `world`, containing a regular `level.dat`. Server root, path aliases, nesting/traversal, symlinks, special files, ZIP/MCWORLD uploads and existing `converted-*` outputs cannot be selected. Upload/extract an unpacked world through normal Files operations first. Uploaded `*.chunker.json` settings are rejected rather than silently controlling Chunker's optional settings.

Bounds: two active jobs per daemon, one operation per server, 32 in-memory job records, 30-minute job deadline, 50,000 tree entries, 64 directory levels, 8 GiB per world tree (further limited to one quarter of an assigned disk budget), and 32 KiB of log tail. Output metadata is capped separately during tar extraction; no symlinks or hardlinks are accepted. The daemon creates real copies, never writable hardlinks to the original.

Managed Files, settings-file and SFTP writes are refused while a conversion or replacement holds its maintenance lease; reads remain available. This process-local coordination does not cover unsupported external host edits or multiple daemons writing the same server directory.

Preflight checks server bytes, retained staging/recovery bytes and available filesystem space against a reserve for clones, output and backup. These are conservative admission checks, **not a generic filesystem quota or reservation against unrelated host writes**. Shared tool/runtime storage belongs to the daemon, not a server's world allocation.

Completion requires exit success, certified process-tree cleanup, Chunker's `Conversion complete! Took ...` marker, a regular bounded output tree, and structurally parsed `level.dat` with the exact selected edition/version. A zero exit or displayed 100% alone is insufficient.

Successful output appears as `converted-<jobId>/` under the server's Files root, without merging or modifying the original. It is an unpacked world, not a ZIP. Keep at most three converted outputs per server; remove old exports explicitly in Files. Interrupted private staging and retained original directories also have admission limits and require administrator review instead of automatic deletion.

Jobs and replacement tokens are intentionally not persisted across daemon restarts. Existing exported files and safety copies remain on disk. After a restart, an old export is still readable in Files but cannot be replaced automatically using a lost token; run a fresh conversion. If cleanup cannot be certified, the manager operation/run lock and ownership handle remain held and new server operations fail closed until administrator recovery.

## Tool and Java provisioning

Chunker release **1.21.0**, official artifact:

`https://github.com/HiveGamesOSS/Chunker/releases/download/1.21.0/chunker-cli-1.21.0.jar`

- Size: `31883497` bytes.
- SHA-256: `63148e587f5ee5a2ca7ebae1079b235c82ae4d1eff8f7a8a8a3398332513dda6`.
- Catalog source commit: `50aaf6d82e8b68929279161b790d95fdddfe0a40`.

The JAR is automatically downloaded on first use to `<data_dir>/tools/chunker`, outside server storage, with a bounded HTTPS download and allowlisted redirects. Size and hash must both match before atomic publication; existing JARs are reverified before use. This is an integrity pin from official release metadata, not a claim of an independently authenticated binary signature. Upgrades require an explicit reviewed code/catalog update, not `/latest` at job execution.

Java 17/21/25 is discovered using existing managed Java resolution, never a tenant-selected executable or custom JVM arguments. If absent on Linux amd64/arm64, the converter provisions official Eclipse Temurin Java 21 into private `tools/chunker/java21`. This reuses the existing Temurin HTTP/checksum downloader with bounded metadata, a 512 MiB package ceiling, 2 GiB expanded ceiling, 20,000 entries, cancellable confined extraction and atomic publication. The older general Linux installer is deliberately not called because its extraction/download path is not bounded/cancellable. Runtime notices are retained from its archive. No Electron, Node, arbitrary CLI flags or game custom Java agents are used.

The exposed catalog is a reviewed **subset** of the pinned release (Java 1.19.4, 1.20–1.21 patch releases, 26.x, and Bedrock 1.20.80/1.21). Only returned IDs are accepted. There is no synthesized target version, `INPUT`, `PREVIEW`, or `SETTINGS` target. Chunker tool version 1.21.0 is not the target Minecraft version.

## Explicit replacement and recovery

Replacement is a separate asynchronous action requiring `files.view + files.edit + backup.manage`, a successful job, its one-time token, `confirm: "REPLACE"`, `backup: true`, and loss acknowledgement. The current server must still be stopped. The module must still be enabled.

Automatic replacement is intentionally restricted to the configured primary world of **vanilla Java with an exactly matching server version**. Bedrock targets, Paper/Bukkit split dimensions, proxies, mods, custom software, unknown/`latest` versions and different target versions remain export-only. No Java server is replaced by a Bedrock world.

Automatic replacement also requires unambiguous `key=value` Java properties. Files using escaped values/keys, continuations, or alternative separators remain export-only rather than risking replacement of the wrong world.

The service reacquires the operation/run lease, validates current layout, hashes the original and output against their conversion fingerprints, clones the candidate and original, and completes a gzip/tar backup before publication. It uses Linux `RENAME_EXCHANGE` to atomically swap the candidate with the original: failure leaves both unchanged, and a daemon crash cannot leave the live world absent between two renames. Filesystems without atomic exchange refuse replacement rather than use a risky fallback.

The previous original is retained at `<servers_dir>/.chunker-original-<serverId>-<random>/world`, reported as `safety_copy` relative to the **servers directory, not the server Files root**. `RECOVERY.txt` describes the operation; `snapshot/` contains the pre-replacement copy. These are private administrator recovery data and are never automatically pruned. At most three retained original trees are admitted.

The backup is **world-only**, stored under `<backup_dir>/<serverId>/world-conversions/world-before-chunker-<jobId>.tar.gz`; the returned `backup` field is relative to that server's backup directory. It is intentionally excluded from the ordinary whole-server backup catalog, because using a world-only archive with whole-server Restore would remove unrelated server files. Administrator recovery restores its contents into the stopped server's world directory, or exchanges the retained original back after inspection. Do not offer the ordinary whole-server Restore action for this archive. Existing separate converted output remains available after replacement.

## API integration contract

Create exactly one `newWorldConversionHandler(cfg, db, manager)` per daemon router:

| Route | Handler | Permission |
|---|---|---|
| `GET /api/servers/{id}/world-conversion` | `Status` | `files.view` |
| `POST /api/servers/{id}/world-conversion` | `Start` | `files.view+files.edit` |
| `DELETE /api/servers/{id}/world-conversion/{jobId}` | `Cancel` | `files.view+files.edit` |
| `POST /api/servers/{id}/world-conversion/{jobId}/replace` | `Replace` | `files.view+files.edit+backup.manage` |

Start body: `{"input":"world","format":"JAVA_1_20_5","acknowledge_loss":true}`.

Replace body: `{"token":"<replace_token>","confirm":"REPLACE","backup":true,"acknowledge_loss":true}`.

Mutations return 202 plus the job. Poll Status for `jobs`; each includes `id`, `state`, bounded `log`, advisory `progress`, `output`, `can_replace`, `replace_token`, `backup`, `safety_copy`, and timestamps. States: `preparing`, `converting`, `validating`, `succeeded`, `replacing`, `replaced`, `failed`, `cancelled`, `blocked`. Status includes `enabled`, `supported`, `unavailable_reason`, `formats`, `warning`, tool version and bounds. `supported` is a prerequisite check, not proof that a particular launch will pass kernel enforcement checks.

Do not blanket-gate GET/DELETE with the module: disabling `db.FeatureEnabled(id,"world_conversion")` prohibits new Start/Replace and publication, while Status, Cancel and ordinary Files reads remain accessible. Permissions are rechecked in the handlers. Use hosting-node routing for remote servers.

No manager/shutdown integration edits are required: the new stopped lease registers under `launchMu` in existing `importWG`, observes `shutdownCh`, and fences launch/publication. It holds the existing operation lock and live-directory cross-process run lock through all cleanup; it never creates a game ProcessInstance.

## License

Chunker is maintained by Hive Games and licensed under MIT, copyright 2024 Hive Games. Its complete license is bundled in `internal/chunker/LICENSE.chunker`, embedded in NeoPanel, and installed beside the JAR. Preserve dependency and runtime notices when redistributing artifacts. Chunker is not a Microsoft warranty of conversion fidelity.
