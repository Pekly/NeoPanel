# Built-in SFTP resource safeguards

## Connections

Admission occurs before launching a connection worker. The listener allows at
most 64 concurrent incomplete SSH handshakes, 8 per remote source IP, and 256
total connections (including authenticated connections). Source ports do not
create separate allowances. Each connection allows at most 8 session channels.
Excess connections/channels are rejected; established sessions are not evicted.

A 30-second absolute socket deadline covers SSH negotiation and authentication,
including silent/slow peers. Successful handshakes clear that deadline so normal
long-lived transfers keep working. Socket deadlines do not interrupt synchronous
database or password-verification work already running inside an auth callback;
its concurrency is nevertheless bounded by admission. Shutdown closes tracked
sockets and waits for connection/session workers. These are local-listener
limits, not distributed limits across multiple panel processes.

## Logical file size

Every WRITE and size-changing SETSTAT/FSETSTAT uses the smaller of 8 GiB and the
server's current **positive** `DiskLimitMB` assignment (MiB units). Unassigned or
nonpositive assignments still have the 8 GiB per-file ceiling. Offsets, lengths,
and unsigned protocol sizes are checked before arithmetic/conversion, and holes
count toward logical size. Existing handles re-read the assignment on mutation.

OPEN creates empty files or truncates to zero; it does not apply a client-supplied
size attribute. Oversized existing files may be read, deleted, or explicitly
shrunk, but cannot be opened for ordinary writes or modified through old write
handles until shrunk. Permission/feature checks, mutation leases, and root/inode
confinement still apply.

## Not a total disk quota

Size checks and SFTP mutations are serialized without per-packet directory walks
or unbounded per-file accounting maps. Delete, shrink, and replacement uploads do
not consume a lifetime transfer allowance.

The ceiling applies to each file, not aggregate occupancy across multiple files.
Existing data, API/restore/game-process writes, other listeners, filesystem
metadata, and external mutations are not accounted for. These limits mitigate
individual sparse-file abuse but do not guarantee against disk exhaustion.
Use filesystem/OS disk and inode quotas for aggregate enforcement.
