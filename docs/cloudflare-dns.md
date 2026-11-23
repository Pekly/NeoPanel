# Cloudflare DNS Automation

NeoPanel provides direct, automated DNS management with Cloudflare via the Cloudflare API v4. This eliminates manual copying of DNS records into domain dashboards and prevents connectivity issues caused by CDN proxying.

## Features

- **Automatic Zone Detection:** NeoPanel automatically queries Cloudflare zones matching your domain (e.g. `example.com` for `play.example.com`).
- **A / CNAME Record Synchronization:** Creates or updates the target address pointing to your node's public IP or hostname.
- **Strict DNS-Only Enforcement (`proxied: false`):** Automatically ensures Cloudflare proxy (orange cloud) is disabled, since HTTP proxying breaks Minecraft TCP connections.
- **SRV Record Configuration:** Configures the `_minecraft._tcp.<fqdn>` SRV record to map to your server's assigned port, allowing players to join without typing the port number.
- **1-Click Deletion:** Cleanly removes managed records from Cloudflare when a server or vanity domain is decommissioned.

## API Endpoints

- `GET /api/admin/dns/cloudflare`: Retrieve saved Cloudflare credentials (masked).
- `POST /api/admin/dns/cloudflare`: Configure global Cloudflare API Token and optional Zone ID.
- `POST /api/admin/dns/cloudflare/test`: Verify API token connectivity and list accessible zones.
- `GET /api/servers/{id}/subdomain/cloudflare`: Get server's Cloudflare synchronization receipt and live status.
- `POST /api/servers/{id}/subdomain/cloudflare/sync`: Synchronize server's address and SRV records to Cloudflare.
- `DELETE /api/servers/{id}/subdomain/cloudflare`: Delete server's DNS records from Cloudflare.
