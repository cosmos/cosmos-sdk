# Write-Ahead Log (WAL)

The WAL is the source of truth for durability. Every set and delete operation is appended
to `wal.log` before the tree is mutated, and the WAL is fsynced before the commit is
considered durable. On crash recovery, the tree is reconstructed by replaying WAL entries
forward from the latest checkpoint.

## Entry types

The WAL is a sequence of three entry types:

- **Set** — records a key-value pair being written
- **Delete** — records a key being removed
- **Commit** — marks the end of a version's updates (the commit boundary)

A version's entries are always: zero or more Set/Delete entries, followed by exactly one Commit.

## WAL as a data store

The WAL also serves as blob storage — leaf and branch nodes can reference key/value data
in the WAL by byte offset instead of duplicating it in `kv.dat`. The `WALWriter` maintains
a key cache that maps keys to their WAL offsets so checkpoint writing can look them up.

During compaction, old WAL entries may be pruned (versions before the retain point are dropped).
When a WAL entry is pruned but a node still references its data, the blob is copied to `kv.dat`
and the node's offset is remapped. See `remapBlob` in `compactor.go`.

## Rollback

Because the WAL is append-only, rollback is just a file truncation back to the previous
commit boundary offset. The `WALWriter` tracks this offset so rollback is O(1).

During crash recovery, if the WAL contains entries for one version beyond what the commit info
file records (a crash mid-commit), those entries are automatically truncated. More than one
version beyond is treated as corruption. See `wal_replay.go`.
