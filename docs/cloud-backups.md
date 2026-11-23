# Cloud & Offsite Backups

NeoPanel supports offsite backup storage using standard S3-compatible cloud providers:
- **Cloudflare R2**
- **Amazon Web Services (AWS S3)**
- **Backblaze B2**
- **MinIO**
- **Wasabi & DigitalOcean Spaces**

## Authentication

Uses AWS Signature Version 4 (SigV4) implemented in pure Go without external SDK dependencies.

## Features

- **Offsite Redundancy:** Store compressed server archives (`.tar.gz`) offsite to protect against local hardware failure or disk exhaustion.
- **Automated Cloud Uploads:** Automatically uploads newly created manual or scheduled backups to S3 storage when enabled.
- **Retention Policies:** Configurable retention periods (e.g. 30 days) to prune old archives from cloud storage.
- **1-Click Cloud Restore:** Downloads the remote backup directly into the server node and restores files seamlessly.

## API Endpoints

- `GET /api/admin/backups/cloud`: Retrieve Cloud Backup settings (masked credentials).
- `POST /api/admin/backups/cloud`: Update Cloud Backup settings.
- `POST /api/admin/backups/cloud/test`: Test connection and permissions against the configured bucket.
- `GET /api/servers/{id}/backups/cloud`: List cloud-stored backups for a server.
- `POST /api/servers/{id}/backups/{name}/cloud/upload`: Upload a local backup to cloud storage.
- `POST /api/servers/{id}/backups/{name}/cloud/restore`: Restore a server from cloud storage.
- `DELETE /api/servers/{id}/backups/cloud/{name}`: Delete a backup from cloud storage.
