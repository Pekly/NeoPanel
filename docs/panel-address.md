# Changing your panel address

Open **Admin → Panel address**. Enter a hostname such as `panel.example.com` and NeoPanel adds `https://` for you. You can also enter a full URL with a custom port.

The preview shows the address that will be saved. **Use this browser address** copies the origin you're currently visiting, without the page path or query. **Open address** opens it in another tab so you can check it before saving. It does not send your current session token to that address.

Click **Save panel address** when you're ready. Invitation links, password-reset links, and passkey settings use the new value immediately—no panel restart is needed. You stay on the current page rather than being redirected. **Reset changes** discards an unsaved edit.

The old **Admin → Appearance → Public URL** form still works. Both pages edit the same setting.

## If you're moving to a new domain

Saving the panel setting doesn't change DNS or your web server. Prepare the new address first:

1. Point its DNS record at the machine or proxy serving NeoPanel.
2. Add the hostname to your reverse proxy and enable HTTPS for it. Keep the old hostname working until you've tested the new one.
3. Open the new address and check that the panel loads and you can sign in.
4. Save the new address in **Admin → Panel address**.
5. Update any bookmarks or external integrations that use the old URL.

For a standard Caddy installation, update the site hostname in `/etc/caddy/Caddyfile`. Preserve the existing reverse-proxy target, port, and any other sites. Validate before reloading:

```bash
sudo caddy validate --config /etc/caddy/Caddyfile
sudo systemctl reload caddy
```

Only reload after validation succeeds. If you use another proxy or a hosted proxy service, change its hostname and certificate settings there instead. You don't need to rerun the NeoPanel installer.

## Passkeys

Passkeys belong to the hostname where they were registered. Before switching hostnames, make sure you can sign in with a password, remove existing passkeys from **every account**, and register new ones after the move. Disabling passkeys alone does not remove them.

NeoPanel refuses a hostname change while incompatible passkeys exist. This safeguard also applies when you enter a bare hostname through the API.

## What this setting changes

- Changes the panel's public URL, not Minecraft game addresses, SFTP hosts, worker API URLs, or the HTTP listening address/port.
- Saves a database override for `server.public_url`; it does not rewrite your YAML file.
- Requires HTTPS, except for explicit HTTP localhost development URLs.
- Accepts a hostname and optional port, not a page path, credentials, query string, or fragment. Use a DNS hostname rather than a public IP address.
- Cannot be cleared. If a save fails, the existing address stays active and your edit is kept so you can correct it.
