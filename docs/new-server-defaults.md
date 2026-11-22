# New-server defaults

New Java servers now default to **Latest available** and **Java 25**. This applies to the creation form and to API creation requests that omit the version or Java runtime. Steel remains a native server and does not use Java.

## What “latest” means

`latest` is resolved when server files are downloaded, using the selected software's upstream or supported download source. It is not a version number baked into the panel, and it is not the highest entry in the panel's mixed version catalog.

- Paper, Fabric, Velocity and other software can have different newest supported releases/builds.
- Waterfall's latest available build is its newest archived release, not the newest Minecraft release.
- If the current release cannot be resolved or downloaded, the download fails instead of silently using a pinned older artifact.
- The version field can retain the `latest` selection; it is not a promise that an already-installed JAR updates itself later.

## Manual choices and existing servers

You can still select or type a specific version and choose another Java runtime. Changing to a different software type resets the version to Latest available; a catalog refresh does not replace an explicit choice. Java stays at 25 unless you select another runtime.

Older server versions and some plugins/mods may require Java 8, 17 or 21. Choose the runtime required by that installation. If Java 25 is missing, the form keeps it selected and shows its availability; it does not silently downgrade to another installed Java version.

Existing servers are not upgraded, restarted or assigned a different runtime by this change. Reinstall and existing-server settings continue to preserve their current selections unless you explicitly change them.

Update the panel and workers together so creation defaults and download handling match.
