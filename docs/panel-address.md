# Changing your panel domain and HTTPS

Open **Admin → Panel address**. Enter a hostname such as `panel.example.com`; HTTPS is added automatically. The same verified hostname is used for the public panel URL, account-email links, and passkeys.

## Check, prepare HTTPS, apply

1. Point the new domain's DNS records at the panel or its HTTPS proxy. Keep the old domain working during the move.
2. Enter the new address. **Use this browser address** copies only your current origin, not its path or query. **Open address** opens it without forwarding your panel session token.
3. Click **Check domain & HTTPS**. The checklist reports:
   - **DNS:** all returned addresses must be usable public addresses. Every address is checked so a bad AAAA record cannot hide behind a working A record.
   - **HTTPS:** a trusted certificate covering this exact hostname. The issuer, expiry, and whether the served certificate is from Let's Encrypt are shown. Other trusted issuers work too; behind a CDN, this is the CDN's public-facing certificate.
   - **This panel:** the address must return a short-lived proof from this NeoPanel instance, not just any successful web page. Probes send no login tokens or cookies and don't follow redirects.
   - **Passkeys:** existing keys must be compatible with the proposed hostname. Blocking accounts are listed without exposing credential material.
4. Resolve any reported issues. On supported installations, **Set up HTTPS** can prepare Caddy and request a Let's Encrypt certificate as described below.
5. Once the checks pass, click **Apply verified address**. NeoPanel repeats the checks immediately before saving. Email links and the passkey hostname then use the new address without restarting the panel. The current browser tab is not redirected.

Checks expire after five minutes and belong to the administrator and exact address checked. Editing the address, signing out, or a failed apply requires another check. A concurrent address change by another administrator is not silently overwritten. **Reset changes** discards an unsaved edit.

**HTTP** is available only for explicit localhost development. Public addresses require HTTPS. Private/LAN-only domain probing is not supported by this public-domain wizard; it is not a general internal-network scanner.

The older **Admin → Appearance → Public URL** location opens the same verified form. Email settings and an administrator's passkey settings link directly to this workflow. SMTP delivery remains a separate setting from the URL embedded in account emails. Save email settings before sending a test: the test uses the saved configuration. Verify the sender with your email provider and follow its SPF/DKIM instructions; panel-domain setup does not configure email DNS.

## Automatic Let's Encrypt setup

Automatic setup is offered only for a recognized, installer-compatible **Linux panel with local Caddy**:

- NeoPanel must have the required privilege and use the expected loopback listener behind the standard active Caddy service.
- The existing `/etc/caddy/Caddyfile` must match the installer template, plus any aliases previously added by this feature. Custom configurations, imports, service overrides, API-managed drift, and unsupported installations are refused rather than rewritten.
- All DNS records for the new domain must match addresses of the recognized Caddy host. Standard HTTPS port 443 is required. Public ports 80/443 must reach Caddy; NeoPanel doesn't open firewall ports or edit DNS for you.

Enter a certificate contact email, confirm domain ownership and the Let's Encrypt agreement, and enter your administrator password. API keys cannot authorize this privileged operation.

NeoPanel adds a new site with an explicit Let's Encrypt issuer while preserving the old site bytes and upstream. It stages and validates the configuration, keeps a protected backup, and reloads Caddy—never restarts the panel or game servers. Caddy stores and renews its certificates.

**Reload success is not certificate success.** Issuance is asynchronous and subject to DNS, reachability, CA policy, and rate limits. The UI performs a bounded series of rechecks while it waits. The saved panel URL does not change until verification passes and you apply it. If issuance takes longer, use **Check domain & HTTPS** again; repeatedly submitting certificate setup is unnecessary.

If automatic setup is unavailable, the checklist explains why. Configure the hostname/certificate through your existing proxy or hosting provider and rerun the checks. You don't need to reinstall NeoPanel.

### Recovery after an uncertain reload

Protected `/etc/caddy/.neopanel-backup-*` files preserve the previous configuration. An unfinished transaction is recorded in `/etc/caddy/Caddyfile.neopanel-pending`; further automatic edits are blocked until an administrator reconciles it.

A timed-out systemd reload may still finish later. Do not assume it failed, delete the pending record, or blindly restore a backup. Check the reload job, running Caddy configuration, on-disk Caddyfile, and both hostnames first. The new configuration is additive and retains the old hostname. Resolve the service state and validate the intended file before clearing the recovery record.

For a manually managed Caddy configuration, preserve the existing upstream/port and other sites, validate the updated file, then reload only after validation succeeds:

```bash
sudo caddy validate --config /etc/caddy/Caddyfile
sudo systemctl reload caddy
```

No certificate requests or production proxy changes are made merely by opening the settings page or running domain checks.

## Moving passkeys

Passkeys belong to the hostname where they were registered; a certificate change alone cannot move them. Before changing hostnames, confirm password sign-in and have the listed accounts remove incompatible keys under **Account → Security**. Register replacements after applying the new address. Disabling passkeys doesn't remove them, and NeoPanel never silently deletes or rewrites users' credentials.

Changing only the scheme or port while retaining the hostname doesn't require deleting compatible keys, but the new origin still has to pass the checks. A registration racing the change is handled by the database's atomic hostname guard.

## Host-side recovery when a web change fails

First resolve the failed checks where possible. If the web panel still cannot save the address, connect to the **panel host**, rerun `sudo neopanel-installer`, and choose **12 — Change panel address / FQDN**. If that shortcut is unavailable, rerun your current `install.sh` and choose the same option. This does not reinstall NeoPanel. Older binaries must be updated before they support recovery.

Configure DNS and trusted HTTPS through your proxy and check the new address in a browser **before** confirming. Recovery is a manual, unverified address change; it does not provision certificates, edit proxy/config files, restart services, change game/node addresses, or bypass incompatible passkeys. Type `CHANGE` to save, or anything else to cancel.

Direct CLI on the host:

```bash
sudo neopanel -config /etc/neopanel/config.yaml panel-address --url https://panel.example.com
```

On Windows, run the installed `neopanel.exe -config C:\path\to\config.yaml panel-address --url https://panel.example.com` instead. Use the actual installed configuration, database, and service working directory (especially if Linux paths are relative).

For deliberate unattended recovery, the installer supports `--change-fqdn --panel-url https://panel.example.com --confirm-address`; the CLI uses `--confirm`. Generic `--yes` is not sufficient. `--config` and `--binary` select custom installations. Keep access at the original hostname until affected accounts have removed incompatible passkeys and the new address works.

## Scope and API compatibility

- The setting does not change game/SFTP addresses, worker API URLs, the panel listener, or `config.yaml`. It saves the existing database override for `server.public_url`.
- The value cannot be cleared. Invalid paths, credentials, query strings, fragments, public IP addresses, and untrusted certificates aren't accepted by the guided flow.
- `POST /api/admin/panel-url/check`, `/apply`, and `/https` implement the guided administrator workflow. Checks are bounded/rate-limited; apply consumes its verification ticket and rechecks before saving.
- The existing administrator-only `PUT /api/admin/panel-url` remains a **manual, unverified configuration API** for script compatibility. It still validates the origin and protects passkeys, but does not run network checks. The web form never falls back to it after a failed verification.
