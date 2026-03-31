# Branch Nodes

Branch (inner) nodes form the structure of the IAVL tree. Each branch has a left and right
child (which may be leaves or other branches) and stores the metadata needed for balanced
lookups: the key that separates the left and right subtrees, the subtree height for AVL
balancing, and the subtree size for index-based access.

Unlike leaf nodes, branches don't store values — only the key used as the split point.
The key is stored externally (in `wal.log` or `kv.dat`) and referenced by offset, but
the first 8 bytes are also inlined in the record for fast prefix comparisons that can
skip the external read in many cases.

When a checkpoint is written, branch nodes are serialized as fixed-size 88-byte records
and appended to `branches.dat` in post-order (children before parents). Post-order is
critical because it means a child's file offset is known before its parent is written,
so the parent can store the child's offset directly for O(1) lookups.

## On-disk layout (BranchLayout, 88 bytes)

```
 Offset  Size  Field
 ──────  ────  ──────────────────────────────────────────────────
   0       8   NodeID
                 ├── checkpoint  (4 bytes) — which checkpoint wrote this node
                 └── flagIndex   (4 bytes) — branch flag (bit 31=0) + 1-based post-order index
   8       4   Version          — the tree version that created this branch
  12       8   Left             — NodeID of the left child
  20       8   Right            — NodeID of the right child
  28       4   LeftOffset       — 1-based file offset of left child (0 if in a different changeset)
  32       4   RightOffset      — 1-based file offset of right child (0 if in a different changeset)
  36       5   KeyOffset        — Uint40, byte offset of the key blob
  41       1   Height           — AVL tree height of this subtree
  42       1   flags            — bit fields:
                 ├── bit 7: key is in kv.dat (1) vs wal.log (0)
                 └── bits 0-4: inline key prefix length (0-31)
  43       5   Size             — Uint40, number of leaf nodes in this subtree
  48       8   InlineKeyPrefix  — first 8 bytes of the key (for fast prefix comparisons)
  56      32   Hash             — SHA-256 hash of this branch node
```

Total: 88 bytes per branch node (no padding).

## Child references: NodeID + file offset

Each child is referenced by two things:

- **NodeID** (Left/Right): the stable identifier. Always set. Used to find the child
  when it's in a different changeset — look up the checkpoint metadata for that NodeID's
  checkpoint number, then search within the checkpoint's node set.

- **File offset** (LeftOffset/RightOffset): the 1-based position in `leaves.dat` or
  `branches.dat` within THIS changeset. Only set when the child is in the same changeset
  (its checkpoint number >= this changeset's start checkpoint). When set, this gives O(1)
  child lookup — no checkpoint metadata search needed.

  When the child is in a different changeset (e.g. an older node that wasn't rewritten),
  the offset is 0, and resolution falls back to the NodeID path.

This dual-reference scheme costs 8 extra bytes per branch (the two offset fields) but
avoids checkpoint metadata lookups on the hot path. See the design note in `branch_layout.go`
for discussion of a more compact alternative.

## Inline key prefix

The `InlineKeyPrefix` field stores the first 8 bytes of the branch's key directly in
the record. The `flags` field encodes the actual key length (up to 31) in its lower 5 bits.

This allows many key comparisons to be resolved without reading the full key from
`wal.log` or `kv.dat`. For keys shorter than 8 bytes, the entire key is inline and no
external read is ever needed. For longer keys, the inline prefix can eliminate most
comparisons (only keys sharing the same 8-byte prefix require the full read).

See `key_prefix_cmp.go` for the comparison logic.

## Branch indexing

Within a checkpoint, branches are indexed by their 1-based post-order traversal index.
The same lookup strategies as leaf nodes apply: direct arithmetic when contiguous,
interpolation search when compaction has created gaps (see `NodeMmap.FindByID`).