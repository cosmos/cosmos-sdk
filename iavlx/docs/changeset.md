# Changesets

A changeset is a directory containing the on-disk data for a range of tree versions.
Each IAVL tree has one or more changeset directories.

## Directory naming

**Original changeset** — written during normal execution, named by start version:
`1`, `1053`, etc. The WAL spans from start version to end version (determined by reading
the WAL to the end).

**Compacted changeset** — produced by merging one or more changesets with orphan pruning.
Named `<start>-<end>.<compacted>`, e.g. `1-2501.3000`. The end version and compacted-at
version are encoded in the name. The WAL may not span the full range (pruned entries).

**Temp directories** — during compaction, the output is written to a `-tmp` suffixed directory,
then renamed atomically when complete. On startup, any `-tmp` directories are deleted
(interrupted compaction).

## Files

```
  wal.log          — write-ahead log (set/delete/commit entries per version)
  checkpoints.dat  — checkpoint metadata (64 bytes each, see checkpoint.md)
  branches.dat     — branch node records (88 bytes each, see branches.md)
  leaves.dat       — leaf node records (56 bytes each, see leaves.md)
  kv.dat           — key/value blob data referenced by nodes not in the WAL
  orphans.dat      — log of orphaned nodes and when they were orphaned (see orphans.md)
```

## Lifecycle

1. **Active** — the current changeset being written to. Commits append WAL entries,
   the checkpointer writes node data and checkpoint metadata.
2. **Sealed** — hit the rollover size threshold (`ChangesetRolloverSize`). A new changeset
   is created; this one won't be written to again.
3. **Compacted** — the compactor merged this changeset into a new one. Readers are redirected
   to the compacted output. Queued for deletion once all readers unpin.
4. **Deleted** — all readers done, cleanup proc removes the directory from disk.

Compacted changesets can themselves be re-compacted in a future compaction run.

## When changesets roll over

A new changeset is created when the current one's WAL reaches `ChangesetRolloverSize`
(default 2GB). A checkpoint is always written at rollover time regardless of
`CheckpointInterval`, since the new changeset needs a starting point.