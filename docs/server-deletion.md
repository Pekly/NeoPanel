# Server deletion and leftover processes

## Deletion contract

Update/redeploy both the panel and its workers together. The UI only reports completion after the hosting node returns `deleted: true` with the requested `server_id`. Older workers without this acknowledgement require an inventory refresh and upgrade; an ambiguous response is not presented as success.

The hosting node stops the tracked process even if its displayed status is stale. Linux process-group termination is verified before files are removed. The launcher remains unreaped until group cleanup finishes, preventing its PID from being reused during group signalling. Systemd scope shutdown must be confirmed; connection/bus errors are not treated as an absent scope.

Linux nodes also inspect live process working directories and executable paths before destructive maintenance. A surviving process referencing the server directory (including a path marked ` (deleted)`) blocks deletion and reports its PID. Inspection failures block cleanup rather than silently treating the server as stopped. This requires access to the node's process table. Windows development hosts use tracked process completion; they do not have the Linux orphan-inspection guard.

Failed process cleanup retains the instance and server metadata. Once the survivor is stopped or a transient systemd error is resolved, retry Stop/Delete to revalidate cleanup. Failed file/database cleanup retains the instance for retry, though files already removed by a partially completed filesystem cleanup cannot be rolled back.

Changing `node_id` in settings is not migration and is rejected. The hosting node is read-only in admin settings. Master-side records assigned to a remote node cannot be silently removed using the master's local process manager. A real move requires stopping and transferring the server.

## An orphan from an older deletion

Existing processes are not automatically killed based on a port number. On the **affected Linux node**, identify the listener first:

```sh
sudo ss -ltnp '( sport = :25565 )'
```

Using the PID reported there, inspect its working directory, command, start time and cgroup (replace `PID` below):

```sh
sudo ps -p PID -o pid,ppid,lstart,args
sudo readlink /proc/PID/cwd
sudo readlink /proc/PID/exe
sudo cat /proc/PID/cgroup
```

Confirm that these identify the deleted server, not another Minecraft server or proxy. If it belongs to a `neopanel-....scope`, stop **that exact scope** with `sudo systemctl stop EXACT_SCOPE_NAME`. If no scope exists, terminate only the verified server PID; do not use broad commands such as `pkill java`, `killall steel`, or blindly kill everything on a port. Recheck the PID identity immediately before signalling it because numeric PIDs can be reused.

Run `ss` again to verify the port is released. Then retry the new server. Disconnecting/removing a cluster node registration is separate from deleting its servers: disconnection intentionally leaves its servers running on that worker.

## Tests

- `go test ./...` covers deletion guards, cleanup retries, explicit acknowledgement, remote-node failures, and port release by a real test listener.
- On Linux, `go test ./internal/runner -run 'TestLinux(Delete|Cleanup|Owned)' -count=1` additionally exercises an untracked live listener, unreaped launcher identity, and child-process termination.
- `node --test web/tests/*.test.cjs` covers frontend deletion confirmation and routing.
