# JVM memory and node policy

`memory_max` / `Server.MemoryMaxMB` is a Java **heap** budget (MiB), not an
RSS/process budget. Managed `-Xms` / `-Xmx` options are emitted after custom VM
options, so explicit custom heap-size aliases cannot supersede panel settings.
Custom arguments must be self-contained VM options. Launcher operands, `@`
argument files, `-XX:Flags=` and `-XX:VMOptionsFile=` are rejected at startup:
a tenant could otherwise edit an indirectly loaded file or terminate option
parsing before managed heap flags. Existing configurations using those forms
must be converted to inline VM options. Whitespace splitting remains unchanged;
this field is not a shell command.

When a configured node has `AllowUserRAM=false`, non-admins (including owners)
cannot change **any** `java_args`, not just recognized heap flags. Unchanged
values may be echoed by settings forms. Admins can change flags; on permissive
nodes, owners/subusers with `settings.edit` retain that ability. Administrators
must change the panel memory fields, rather than custom flags, to change the
managed heap budget. For direct worker/local requests, an empty legacy node ID
resolves to `local`; restrictions apply only where a local node record exists.

For proxied non-admin requests, the master forwards RAM, port, and Java-runtime
policy from its trusted destination-node record. The worker uses that policy
even when its server has an empty/local node ID and no local policy row. Only
forwarding middleware, after authenticating the admin/node credential and original
user identity, can attach this request-only policy; it is not a persistable JWT
claim. Client headers cannot supply or override it. Admins retain their bypass.
Direct worker requests and older masters without the policy header retain local
policy lookup behavior. **Deploy this change to both master and workers** before
relying on remote restrictions; older workers cannot enforce forwarded policy.
Keep worker access limited to the intended trusted deployment, or configure its
local node policy too if tenants may connect directly.

Both systemd-only and bubblewrap-plus-systemd launch paths set:

- Java: `MemoryMax = heap + max(512 MiB, floor(heap / 2))`.
- Native software (Steel): `MemoryMax = memory_max` (unchanged).
- Non-positive/unconfigured memory budgets: no memory limit (unchanged).

For example, a 2 GiB Java heap has a 3 GiB cgroup memory ceiling; a 512 MiB
heap has a 1 GiB ceiling. The additional allowance is for metaspace, JIT/code
cache, thread stacks, direct buffers, JVM native allocations, and cgroup-accounted
file cache. It deliberately does **not** set the whole-process limit equal to
heap size. The allowance is finite, not a guarantee for every modpack; operators
must monitor cgroup OOM events and size allocations for workload overhead. Node
capacity planning should reserve the **whole ceiling**, not just the heap.
Changes take effect on the next start/restart, not on existing processes.

This is defense in depth, not a JVM-code security boundary. Plugins, agents,
native libraries, and custom runtimes can allocate outside the heap. Only a
working systemd/cgroup memory controller enforces the aggregate memory ceiling.
`process`/`none`, bubblewrap without systemd, and automatic unrestricted fallback
have **no cgroup ceiling**. Select and validate systemd-backed isolation for
untrusted tenants. `MemoryMax` is the cgroup memory limit; swap remains governed
by the host's existing swap policy (no `MemorySwapMax` change in this patch).

# Port and destructive file policies

Port add/delete/set-primary handlers enforce the same configured-node
`AllowUserPorts` restriction as server settings updates, before database,
server.properties, or firewall mutation. Admins bypass this node restriction;
normal route membership and `settings.edit` checks still apply to tenants.

HTTP rename, move, unzip, and extract require **both** `files.edit` and
`files.delete`, consistently with destructive SFTP writes. Owners/admins retain
full access. Rejected requests never enter the file handlers.

## Known firewall limitation

Existing UFW rules have no panel/server ownership marker. Authorized allocation
deletion still calls the legacy generic `ufw delete allow PORT`, `PORT/tcp`, and
`PORT/udp` cleanup, which can remove administrator-created or shared rules.
Changing this safely requires rule provenance plus reconciliation of existing
allocations/rules; guessing ownership or automatically migrating unmarked rules
is intentionally outside this bounded fix. Denied mutations do not call UFW.
