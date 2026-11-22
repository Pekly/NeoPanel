# Installed server artifact UI

## Integration contract

- `server.artifact` is optional on inventory/status records. It is detected metadata, not configured settings. `GET srvApiPath(server, '/artifact')` returns the direct Info object and requires `files.view` (including workers).
- `NPServerArtifact` is loaded before `app.js`. It never writes detected software to `S.core` or `server.type`. Initial and updated headers, dashboard modes and admin inventories use the same display helper. The Software chooser labels its configured installation explicitly.
- Required fields used: `status`; optional `software`, `minecraft_version`, `project_version`, `build_id`, `build_name`, `required_java`, `java_source`, `minimum_java`, `filename`, `size`, `modified`. Provider messages/HTML are not rendered. Project releases are distinct from Minecraft versions. Native Steel retains its existing configured native display.
- One optional request at a time, 15-second timeout, 24 bounded cache entries. Account, token, node, server, configured filename/type/version form its immutable request key. No polling or automatic retry after failure. Supplied metadata avoids the extra request; fresh supplied summaries take precedence over lookup cache.
- Requests only run on an authorized active server page, never inventory pages or without `files.view`. `optionalMetadata: true` prevents unsupported/provider errors from changing the workspace connectivity banner. No special backend provider exception is required by this UI; HTTP 404/unsupported workers show a local retry/manual-selection message.
- `files(server, directory, rows)` observes the authoritative configured filename and size/mtime signature after successful directory loads. `requestResult` invalidates successful file/settings/install mutations; unrelated name/Java settings saves retain metadata. `installed` invalidates the old installation, then accepts response metadata. Retry deliberately invalidates. `reset` aborts and clears caches at logout. No raw file/hash upload or browser hash lookup is introduced.
- Settings PUT must return its authoritative server record (current API already does). Returned filename/version/java_path and artifact win over submitted settings. Software/reinstall responses may optionally include `artifact`; otherwise one bounded lookup identifies the replacement.

## Java behavior

New Java servers send `java_path: 'auto'`. Choices are Auto (detected build), System default, managed 8/11/16/17/21/25 and existing custom paths in Settings. Blank/java/default remain system-default aliases, never inferred latest Java; untouched aliases are preserved on autosave. Node Java locks and node-specific installation status remain in force.

Only identified metadata `required_java` is an exact recommendation. `minimum_java` is explicitly a bytecode minimum, not exact. Unknown/missing/error states ask for identification or a manual selection. Catalog versions do not imply an exact requirement before download.

Software/reinstall dialogs preserve the current runtime and point to Settings for an intentional Auto choice. They never send a heuristic follow-up Java PUT. This avoids silently replacing an existing explicit runtime/custom path or bypassing the host's Java lock.

## Design and checks

Uses the established graphite/light semantic tokens, Inter/mono typography, 8px radius and 4/8px spacing. Inline identity/status/retry UI wraps long labels; no clipping workaround or animation dependency. Retry is a semantic button with visible keyboard focus, busy/disabled and recoverable error states. Metadata labels are escaped and image failures fall back once to `/img/default-server-icon.png`; user-uploaded server icons are unchanged.

Run `node --test web/tests/server-artifact.test.cjs web/tests/create-defaults.test.cjs web/tests/software-selection.test.cjs web/tests/steel.test.cjs web/tests/proxy-settings.test.cjs`.

For actual Chrome checks set `NP_LAYOUT_PLAYWRIGHT` to a Playwright package directory and `NP_LAYOUT_CHROME` to Chrome's executable. The tests use mocked API fixtures and blocked network, not a live service. Artifact tests check 1440/768/390/320px in light/dark modes, keyboard retry, loading/error/success, one-shot image fallback, long-label overflow and reduced motion. Create tests cover 1440/1024/768/390/320px and engine/catalog/manual-choice lifecycle. Optional screenshot output: `$PI_SCRATCH_DIR/artifact-mobile.png`.
