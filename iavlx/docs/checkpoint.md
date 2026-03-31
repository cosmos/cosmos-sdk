# Checkpoints

A checkpoint persists the new IAVL tree nodes that weren't present in a previous checkpoint.
Unlike a full snapshot, a checkpoint only contains the nodes that changed — nodes from earlier
checkpoints are already on disk and don't need to be rewritten.

To reconstruct the full tree at a checkpoint's version, you need the checkpoint's nodes plus
all nodes from earlier checkpoints that are still part of the tree (i.e. haven't been orphaned).

## When checkpoints are written

Checkpoints are written in the background after a commit finalizes. Not every commit gets a
checkpoint — the `CheckpointInterval` option (default 100) controls how many versions pass
between checkpoints. Checkpoints are also always written when a changeset rolls over
(hits the `ChangesetRolloverSize` threshold).

Between checkpoints, the WAL provides durability. On startup, the tree is reconstructed by
loading the latest checkpoint and replaying WAL entries forward to the committed version.
More frequent checkpoints = faster startup (less WAL to replay) but more IO during commits.

## What a checkpoint contains

Each checkpoint writes data to four files within its changeset directory:

- **`leaves.dat`**: new leaf node records (fixed-size, 56 bytes each)
- **`branches.dat`**: new branch node records (fixed-size, 88 bytes each), written in post-order
- **`kv.dat`**: key/value blobs for nodes whose data isn't in the WAL
- **`checkpoints.dat`**: a `CheckpointInfo` entry (fixed-size, 64 bytes) recording the
  checkpoint number, tree version, root NodeID, node count/offset ranges, KV file size, and CRC32

## Checkpoint metadata (CheckpointInfo)

```
 CheckpointInfo (64 bytes)
 ┌────────────────────────────────────────────┐
 │  Leaves NodeSetInfo    (16 bytes)          │
 │  Branches NodeSetInfo  (16 bytes)          │
 │  Checkpoint number     (4 bytes)           │
 │  Version               (4 bytes)           │
 │  RootID                (8 bytes, NodeID)   │
 │  KVEndOffset           (8 bytes)           │
 │  CRC32                 (4 bytes)           │
 │  (padding)             (4 bytes)           │
 └────────────────────────────────────────────┘
```

Each `NodeSetInfo` (16 bytes) describes where a checkpoint's nodes live in the data file:

```
 NodeSetInfo (16 bytes)
 ┌────────────────────────────────────────────┐
 │  StartOffset   (4 bytes)  — 0-based offset into the data file (in number of nodes)
 │  Count         (4 bytes)  — how many nodes belong to this checkpoint
 │  StartIndex    (4 bytes)  — 1-based NodeID index of the first retained node
 │  EndIndex      (4 bytes)  — 1-based NodeID index of the last retained node
 └────────────────────────────────────────────┘
```

StartOffset + Count tell you WHERE in the file. StartIndex + EndIndex tell you WHICH
NodeID indices are present. After compaction, Count may be less than EndIndex - StartIndex + 1
because pruned nodes create gaps in the index range.

See `leaves.md` and `branches.md` for how these are used for node lookup.

## Checkpoints are not required for durability

The WAL + commit info file are the durability mechanism. Checkpoints are an optimization —
they reduce startup time by avoiding full WAL replay. If a checkpoint is corrupt or
incomplete (e.g. crash during background writing), it gets rolled back on startup
(see `VerifyAndFix` in `changeset.go`) and the WAL is replayed instead. No data is lost.

## Checkpoints and compaction

During compaction, checkpoint data is copied from the original changesets to the compacted
output, minus any orphaned nodes that are being pruned. The checkpoint's `NodeSetInfo`
ranges may shrink (fewer nodes retained), and the `RootID` may point to a node from an
earlier checkpoint if no tree changes happened at that version. See `compactor.go` for details.