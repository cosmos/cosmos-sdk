# Leaf Nodes

Leaf nodes hold the actual key-value pairs in the IAVL tree. Every key in the tree
corresponds to exactly one leaf node. Branch (inner) nodes connect leaves together
and provide the tree structure for balanced lookups.

When a checkpoint is written, leaf nodes are serialized as fixed-size 56-byte records
and appended to `leaves.dat` in the changeset directory. Because they're fixed-size,
any leaf can be looked up by its file offset in O(1) without an index.

## On-disk layout (LeafLayout, 56 bytes)

```
 Offset  Size  Field
 ──────  ────  ──────────────────────────────────────────────────
   0       8   NodeID
                 ├── checkpoint  (4 bytes) — which checkpoint wrote this node
                 └── flagIndex   (4 bytes) — leaf flag (bit 31=1) + 1-based in-order index
   8       4   Version          — the tree version that created this leaf
  12       5   KeyOffset        — Uint40, byte offset of the key blob
  17       5   ValueOffset      — Uint40, byte offset of the value blob
  22       1   flags            — bit flags:
                 ├── bit 0: key is in kv.dat (1) vs wal.log (0)
                 └── bit 1: value is in kv.dat (1) vs wal.log (0)
  23      32   Hash             — SHA-256 hash of this leaf node
  55       1   (padding)
```

Total: 56 bytes per leaf node.

## Key and value storage

A leaf's key and value data are NOT stored inline in the leaf record — the record only
holds offsets pointing to where the actual bytes live. The `flags` field tells you which
file to look in:

- **wal.log**: if the blob was written during the WAL phase of a commit, its offset points
  into the WAL file. This is the common case for recently committed data.
- **kv.dat**: if the blob wasn't in the WAL (e.g. after compaction rewrote the WAL and
  dropped old entries), it gets copied to kv.dat and the offset points there instead.

This two-file scheme avoids duplicating data — keys that are still in the WAL don't need
to be copied to kv.dat. During compaction, blobs migrate from WAL to kv.dat as old WAL
entries are pruned (see `remapBlob` in compactor.go).

## Leaf indexing

Within a checkpoint, leaves are indexed by their 1-based in-order index (the `Index()` from
their NodeID). When leaves are contiguous (no compaction gaps), a leaf can be found by
direct arithmetic: `file_offset = checkpoint.Leaves.StartOffset + (index - StartIndex)`.
When compaction has created gaps (some leaves pruned), interpolation search is used
(see `NodeMmap.FindByID`).

Parent branch nodes can also reference leaves by their 1-based file position (`fileIdx`),
giving O(1) lookup when parent and child are in the same changeset.
