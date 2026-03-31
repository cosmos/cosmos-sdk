# Node IDs

Every persisted node has a **NodeID** — an 8-byte identifier that locates it on disk.

```
 NodeID (8 bytes)
 ┌─────────────────────────┬─────────────────────────┐
 │  checkpoint (4 bytes)   │  flagIndex (4 bytes)     │
 │                         │  ┌── bit 31: leaf flag   │
 │                         │  └── bits 0-30: index    │
 └─────────────────────────┴─────────────────────────┘
```

- **checkpoint**: which checkpoint this node was persisted in. This is a separate counter
  from the tree version — checkpoints happen less frequently than commits (see below).
- **bit 31**: 1 = leaf, 0 = branch. Tells you which file to look in (`leaves.dat` vs `branches.dat`).
- **index**: the node's position within that checkpoint. Leaf indices are 1-based in-order
  (leftmost leaf = 1). Branch indices are 1-based post-order (children before parent, matching
  the order they're written to disk).

Leaves and branches are indexed separately — they have their own index counters and live
in separate files.

## Checkpoint vs version

These are two different counters:
- **Version** increments on every commit. This is what callers use for historical queries.
- **Checkpoint** increments each time tree data is persisted to disk, which happens every
  N versions (default 100, controlled by `CheckpointInterval`).

NodeIDs reference the checkpoint counter. The mapping to versions is in `CheckpointInfo.Version`.

## Lookup

Given a NodeID: the checkpoint field finds the right changeset, the leaf/branch flag picks
the file, and the index locates the node within that checkpoint's data range. See
`ChangesetReader.ResolveLeafByID` / `ResolveBranchByID` for the implementation.

## Sentinel values

- **Zero value** (`NodeID{}`): "no information". `IsEmpty()` returns true.
- **Empty tree** (`NewEmptyTreeNodeID(checkpoint)`): "tree has zero keys at this checkpoint".
  Has a non-zero checkpoint but zero flagIndex. `IsEmptyTree()` returns true.