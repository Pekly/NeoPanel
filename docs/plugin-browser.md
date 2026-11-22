# Plugin and mod browser

Open **Server → Plugins** (or **Mods** for a modded server). Install and state changes still use the existing permissions and optional Plugins feature policy. Update the panel and workers together for the new project endpoints.

## Browse and inspect

- Click a result's card, title or icon to open its project page **inside NeoPanel**. Its Install button is a separate action.
- The project page presents Modrinth's description, gallery, changelogs and versions, plus compatibility metadata, creators, tags, license and project links when supplied by the author. Missing fields are shown as unavailable or omitted, not invented.
- Back returns to the browse results. **Load more** appends another page instead of replacing the current list. Changing the search or filters starts a new result set.
- The loader selector puts **Auto (All)** first and the current server software directly after it. Choosing a browse loader does not change the server software or the installation target.
- The Minecraft-version filter is a **search filter**, not an installation constraint.

Project text and URLs are untrusted provider content. NeoPanel does not embed the external Modrinth page or execute project HTML/scripts. External links open separately; descriptions use the panel's safe formatting subset rather than allowing arbitrary embedded HTML.

## Install latest or choose a version

**Install / Install latest** asks the backend for the newest published JAR release eligible for the server's actual software. It does **not** require the release's Minecraft-version list to contain the server version. A missing explicit Minecraft match no longer prevents installation.

**Install selected version** installs the chosen release instead. The backend still verifies that it belongs to the requested project and supports the server's software. The version list labels the author's release channel and advertised game versions; inspect these before choosing.

Project pages retain up to 500 recent releases, always including the newest eligible release for the server software. A notice identifies shortened histories; the Modrinth link opens the complete history. Browse results load 24 at a time and retain up to 1,000 results before asking you to refine the search.

Software rules remain in force: Paper can use Paper/Spigot/Bukkit releases; Purpur and Leaf use the Paper family; Waterfall uses BungeeCord; Folia requires an explicit Folia tag. A Fabric build is not interchangeable with Forge or a Bukkit plugin. Steel does not load Java plugins/mods.

**Newest does not guarantee runtime compatibility.** A plugin/mod author may require a different Minecraft version, Java version, dependency, or configuration. Read the compatibility metadata and release notes, back up important data, and choose a suitable release when needed. Dependencies are displayed as metadata; they are not silently installed.

Installation preserves the existing bounded download, confined destination and no-overwrite safeguards. Existing enabled/disabled files are not silently replaced. NeoPanel does not start or restart the game automatically; restart manually when ready and inspect its startup log.

## File tools

Trash is a pinned, folder-like entry at the top of the **root** Files listing. It is a navigation item, not a real server file: it cannot be bulk-selected, dragged, renamed or deleted. Existing Trash permissions and retention rules still apply.

World conversion is available from the Files **World tools** control rather than a prominent standalone toolbar action. Existing jobs remain accessible for monitoring/cancellation where allowed, even when new conversion is unavailable. The Linux-only execution, stopped-server preflight, loss acknowledgement, separate output and guarded replacement requirements are unchanged; see [world conversion](world-conversion.md).

For simultaneous Java/Bedrock ports, see [port allocations](port-allocations.md). Installing a bridge and configuring its UDP listener is separate from assigning the extra port.
