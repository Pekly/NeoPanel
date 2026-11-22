# New-server defaults and Java selection

New Java servers default to **Latest available** and **Auto — match server JAR** (`java_path: "auto"`). Steel remains native and does not use Java.

## MCJars builds

Java distributions are resolved through **https://mcjars.app/api**, not a pinned Minecraft version. The selected build's Java requirement is used first; when it is absent, the matching Minecraft version record may supply the requirement. Missing evidence is not replaced with Java 25 or a version-number heuristic.

- `latest` resolves at download time for the selected software. It does not update an installed JAR automatically.
- Minecraft version, loader/proxy project version, and build identity remain distinct. A proxy release is not displayed as a Minecraft release.
- Exact explicit versions stay explicit. Unavailable builds fail rather than silently downloading a different release.
- Installation accepts a validated raw JAR or a narrowly supported prepared Forge/NeoForge ZIP containing the launcher and libraries. Arbitrary provider scripts/instructions are never executed. Unsupported layouts fail safely.
- Software changes require a stopped server. Reinstall stages and validates downloads before requested cleanup. File publication and database changes are not crash-atomic; keep backups before reinstalling.

## Auto and manual runtimes

Auto resolves from metadata bound to the installed JAR's SHA-256. Before each launch, NeoPanel re-inspects local bytes; launching does not depend on a provider request. The verified exact Java major must be installed on that hosting machine. It never silently substitutes the highest installed runtime.

Supported managed families are **8, 11, 16, 17, 21, and 25**. Historical builds declaring Java 7 use the supported legacy Java 8 family, not an obsolete Java 7 installation; very old software may still need manual compatibility work. Unknown or future requirements need an explicit compatible runtime selection.

Explicit major selections and custom absolute paths are preserved. Existing blank, `java`, or `default` choices continue to mean the host's system default—not Auto. Existing Java 25 settings are not silently migrated because they may be deliberate. Select Auto in Settings to opt in. Reinstall/software changes preserve runtime choices; moving from native Steel to Java selects Auto.

## Custom JAR identity

The active-server identity check reads a bounded local JAR without executing or extracting it. It sends only its SHA-256 digest to MCJars for exact identification, never the JAR contents. Recognized embedded manifests/version metadata provide a fallback. A Main-Class bytecode level is shown only as a **minimum**, not a guaranteed supported runtime.

The header, cards, and settings show verified detected software/version separately from configured software. Unknown/missing identity uses the bundled Minecraft icon and an unknown label, not a guessed or stale configured version. A detected engine never rewrites the configured launch engine. Review that setting separately when uploading a different kind of server.

Display caching is bounded and invalidated by filename/file identity, size, modification time, and expiry. These filesystem hints are not integrity proof; Auto launch always hashes again. Unsupported/deceptive metadata, unusual bundle layouts, or oversized JARs may remain unknown. Use **Identify again** in Settings to bypass the display cache and retry, or choose Java manually when an exact requirement cannot be established. Provider outages leave available embedded information visible.

## Installation and repair

The Linux panel/worker installer provisions all supported families using the newly built backend before its existing service-start step. Updates and `--java` repair propagate provisioning failures. Runtime archives use bounded official downloads, checksums, confined staging, and exact-major executable verification.

```text
neopanel java install-all
neopanel java list
```

These local commands do not start a panel, worker, game server, or service. `install-all` downloads missing families and exits nonzero if any required family is unavailable; `list` reports installed/missing families. Run as the same hosting account and use the same `NEOPANEL_JAVA_DIR` if configured. Platform/upstream availability matters, especially old Java families on Windows ARM64. The panel does not claim installation success when a family cannot be provisioned.

For manual Windows setup, run the command before starting the worker, or use `scripts/install-java.ps1 -BinaryPath <existing-backend.exe>`. The helper does not download the backend, elevate, or register/start services. See [Windows hosting](windows.md).

Update panel and workers together so Auto, artifact metadata, and download handling match. No existing game server is upgraded or restarted merely by changing these defaults.
