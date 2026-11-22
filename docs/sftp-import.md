# Import server files over SFTP

Use **Server → Files → Import SFTP** to copy files directly from another host. The transfer runs on the server's hosting node, so it does not pass through your browser. It works the same way for free and premium servers.

## Before you start

- Create the destination server and stop it. Its file directory must be **empty**, including generated `eula.txt` and `server.properties`. Back up and move or remove existing files through Files first. The importer never clears a server for you.
- Stop the source server if you need a consistent world. SFTP is a file transfer, not a filesystem snapshot; changes during a copy can produce an inconsistent world even when files have the same size.
- Obtain the source host, SFTP port, username and either a password or private SSH key. Encrypted keys also need their passphrase. FTP and interactive SSH/shell commands are not supported.
- Obtain the host's **SHA256 SSH host-key fingerprint** from the provider or another trusted channel. Fetching a fingerprint in the dialog only shows what that connection offered; compare it before confirming trust. Credentials are sent only after the key matches.

Owners and admins can import. Team members need all three permissions: **View Files**, **Create & Edit Files**, and **SFTP Access**. Starting the server later still requires its usual power permission.

## Choose what to copy

1. Open Import SFTP and enter the connection details.
2. Enter or fetch the host-key fingerprint, verify it independently, then confirm it.
3. Test the connection and preview the remote folder.
4. Copy everything in that folder, or select relative paths. Selected directories include their contents.
5. Add excluded paths if needed. Exclusions are relative paths/subtrees, not wildcard patterns. You can leave out logs, backups or caches and optionally skip hidden entries.
6. Set a transfer size limit, optional bandwidth cap and timeout. Timestamp preservation is optional.
7. Start the import. The dialog shows scanning/transfer progress and supports cancellation. Closing the dialog does not cancel an accepted job; reopen it on the same server to check status.

Files land at the server root, not the folder currently open in Files. Remote ownership and special permission bits are not copied. Symbolic links and special files are rejected rather than followed. Filenames must also be safe on Windows; case-colliding names are rejected.

## Completion and recovery

NeoPanel downloads to a private staging directory. The live destination is checked again before publication. Failure or cancellation before publication leaves it unchanged, and staged files are removed. If publication and rollback both fail, the server is blocked pending administrator recovery; the error identifies the retained data.

A successful import leaves the server **stopped**. Review the startup JAR/executable, Java/runtime version and assigned game/query ports before starting. The importer does not migrate panel accounts, allocations, schedules or startup settings. It never deletes or changes source files.

The previous empty destination is retained as a recovery copy, so late writes through already-open file handles are not silently discarded. Imports and restores share a limit of three retained recovery directories per server. An administrator must inspect and remove or archive older copies before another import/restore at that limit.

Progress is in memory, not a durable job queue. Completed summaries expire after an hour, with at most 128 summaries retained per node. A node shutdown cancels active imports; jobs do not resume automatically after a restart. If a start request times out, refresh status before retrying rather than assume it failed.

## Limits

- Two active imports per node, one per initiating account, and one per server.
- Two concurrent connection checks, with a separate connection-attempt rate limit.
- Default transfer budget: 10 GiB, or the server's smaller configured disk allowance. A chosen budget cannot exceed the server allowance or the 100 GiB hard cap.
- At most 100,000 entries and 64 directory levels. Directory metadata and SFTP packets are bounded independently of file contents.
- Timeout: 1–240 minutes, default 60. Idle connections and SSH handshakes have shorter deadlines.
- Credentials are used in memory only and are not stored in job summaries, audit messages or the database. The dialog clears them on close or after an accepted start. Use a temporary, read-only source account and revoke it when finished.

These are application limits, not OS disk quotas. Allow room for staging/recovery copies and enforce disk/inode quotas on the host. Concurrent source changes can only be detected on a best-effort basis.

## Private-network sources

By default, outbound imports accept public unicast addresses only. Private, loopback, link-local and special-use destinations are blocked to prevent users reaching node-local services. DNS results are checked and the chosen IP is pinned for the connection.

An operator can permit a specific private SFTP endpoint in the **hosting node's** YAML configuration:

```yaml
security:
  sftp_import_allowed_hosts:
    - "migration.internal.example:2022"
```

Use exact `host:port` entries, not wildcards or arbitrary networks. Restart the updated node after changing the configuration. Link-local, metadata and other prohibited special destinations remain blocked even with an exception. Keep network/egress firewall rules in place; an allowlist entry gives permitted server users access to that endpoint, not to just one particular source account.

Update the panel and **every worker** to use the importer. A 404/405 from the worker means its API may be too old or a proxy is blocking the route. Private-network exceptions must be configured on the machine making the SFTP connection, not merely on the main panel.
