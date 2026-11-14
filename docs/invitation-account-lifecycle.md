# Invitation-created account lifecycle

## Local account cleanup

An account is eligible only when `users.invitation_created` is true. This flag is
written in the same transaction as the new user INSERT during invitation
acceptance (including a durable remote reservation). Accepting an invitation as
an existing independent user, adding a user directly, and ordinary registration
or administrator account creation do **not** set it. It is internal metadata, not
a client-writable profile field.

Successful `DeleteServerSubuser` removes the membership and checks the affected
user in one SQLite write transaction. Successful database server deletion removes
all its memberships, schedules/tasks, ports and database records, revokes pending
local invitations, and checks affected invitation-created users, including its
owner. The runtime deletion handler invokes this only after its shutdown/operation
gate and filesystem deletion succeed. A runtime refusal or database transaction
failure does not remove the accounts or memberships.

Cleanup retains the account if **any** server ownership or membership remains,
regardless of the server's node or permission list. Built-in user ID 1, admins,
unknown/non-user roles, and any assigned panel role are retained. Unresolved
invitation grants and remote identities are also retained (see below). There is
no background orphan sweep: a later role change alone does not trigger deletion.

Eligible deletion removes password/TOTP/recovery information with the user row,
and explicitly deletes SSH keys, API keys, password reset tokens, passkeys and
WebAuthn handles. This works with SQLite foreign keys enabled **or disabled**.
Audit history and invitation provenance/receipts are retained. Existing JWT,
API-key, SFTP and console WebSocket revalidation reads current users/permissions;
deleted users cannot authenticate again. Existing connections are rejected on
their next authorization/revalidation, not by a new synchronous socket sweep.

The writer lock covers last-association checks and deletion. Membership guards
reject grants to missing users/servers. Deleted invitation-user IDs are tombstoned
so late identity synchronization and owner assignment cannot resurrect that ID.
A fresh independently registered account receives a new ID.

## Upgrades and legacy records

Both new user columns default to false. A one-time backfill marks only records
whose invitation `account_id` points to the same ID, exact username, matching
email, and a creation timestamp within that invitation's creation/expiry window
(UTC, second precision). The historical `account_id` was written only by the new
account INSERT, never when an existing user accepted. Mere membership, matching
email, or `status=accepted` without that creation record is **not** provenance.
Changed/missing/ambiguous legacy records remain unmarked and keep their login.
Backfilling is not an immediate deletion sweep.

Known remote invitation reservations and pre-upgrade users on panels with
configured remote nodes are conservatively pinned as exported. The `local` node-settings row does not count as a remote node.
Historical manual DB copies, nodes removed before this upgrade, or exports with no surviving records
cannot be reconstructed. Administrators of such installations must audit their
historical cluster before relying on automatic local cleanup; this feature does
not infer absent remote associations from absent historical data.

## Remote master/worker behavior and limits

NeoPanel's current auth sync is a credential copy, **not** an authoritative,
transactionally locked global membership/ownership ledger. A worker's empty local
membership list, an HTTP 404, or an unreachable node is not proof that a master
account is globally unused.

Therefore automatic **account deletion is local-only**: all accounts on a panel
with configured remote nodes (excluding `local`), and any identity ever exported/provisioned by the updated
paths, are retained. Export pins are durable before sending credentials, survive
failed delivery, and are not cleared when a node is removed. Workers preserve
master-owned identities even when the final local membership disappears. New
invitation provisioning carries creation provenance only for a genuinely new
identity and never overwrites an existing independent account's provenance.
An unreachable node cannot cause deletion. Retained accounts can still log in,
but removed server memberships no longer grant access.

Worker membership/server deletion durably tombstones invitation grant receipts.
Late acceptance retries cannot restore removed grants, including after server-ID
reuse. After a successful proxied deletion, the master fetches authenticated worker
receipts and reconciles **only matching invitation IDs on that node and server**
to revoked. It keeps their account IDs as provenance and keeps the accounts pinned.
No account is deleted merely because reconciliation succeeded.

If receipt delivery fails or the worker is older, the successful deletion response
is preserved and `X-NeoPanel-Invitation-Cleanup: pending` signals unresolved
reconciliation. Successful receipt processing returns `reconciled`. Direct worker
deletions and failed receipt deliveries can be reconciled explicitly by an admin:

```text
POST /api/nodes/{nodeID}/invitation-removals/reconcile?server_id={serverID}
```

The worker's admin-only receipt endpoint is
`GET /api/cluster/invitation-removals?server_id={serverID}`. Upgrade both ends.
Receipt reconciliation is idempotent and never recreates a grant. Responses larger
than 1 MiB fail closed and require operator reconciliation. There is no automatic
retry scheduler or cross-node account/credential deletion protocol in this change.

To remove a retained remote account, an administrator must independently verify
ownership, team memberships and privileged roles on **every** relevant node,
resolve pending grants, and coordinate account/credential deprovisioning with auth
sync disabled or updated appropriately. Do not clear export pins or delete receipt
tombstones just to force automatic deletion. Safe automatic global deprovisioning
requires a future distributed association/identity protocol; this implementation
does not claim to provide it.
