# Orphans

When a key is updated or deleted in the IAVL tree, the old nodes along the path from root
to that key are replaced by new nodes (due to immutability — each version gets its own nodes).
The replaced nodes are called **orphans**. They're no longer part of the current tree, but may
still be needed to reconstruct historical versions for queries.

## Orphan lifecycle

1. **Created during commit.** When `SetRecursive` or `RemoveRecursive` replaces a node, the old
   node is recorded in the `MutationContext`'s orphan list. (See `mutation_context.go`.)

2. **Written to disk.** After commit finalization, the `OrphanProcessor` writes each orphan as a
   fixed-size `OrphanEntry` (12 bytes) to the `orphans.dat` file in the changeset where the
   orphaned node was originally checkpointed. This runs in the background and doesn't block commits.

3. **Evicted from memory.** The `OrphanProcessor` also clears the orphan's in-memory `MemNode`
   pointer, freeing heap memory — if the node is ever needed again for a historical query, it
   will be loaded from disk.

4. **Pruned during compaction.** The `OrphanRewriter` reads orphan entries and decides which
   orphaned nodes to delete vs retain based on the `RetainCriteria`. Nodes orphaned before
   the retain version are prunable — they're too old for any queryable historical version.
   Retained orphan entries are copied to the compacted changeset's orphan file for future
   compaction cycles.

## OrphanEntry format

```
 OrphanEntry (12 bytes)
 ┌────────────────────────────────────────────┐
 │  OrphanedVersion  (4 bytes)  — the version at which this node was removed from the tree
 │  NodeID           (8 bytes)  — identifies the orphaned node (checkpoint + leaf/branch + index)
 └────────────────────────────────────────────┘
```

The `OrphanedVersion` tells compaction when the node stopped being needed. The `NodeID` tells
it which checkpoint's data files contain the node, so it can be removed from `leaves.dat` or
`branches.dat` during compaction.

## Where orphans are stored

Orphan entries are written to the `orphans.dat` file of the changeset where the orphaned node
was **originally created** (i.e. the changeset identified by the node's checkpoint number),
not the changeset where the orphan-creating commit happened. This means the orphan file is
co-located with the node data it refers to, which simplifies compaction — the compactor
processes one changeset at a time and has both the orphan entries and the node data together.

## Concurrency: the orphan processor lock

The `OrphanProcessor` holds a lock (`mtx`) while writing orphan entries. Compaction also
acquires this lock during two critical moments:

- **Preprocessing**: reading existing orphan entries to build the delete map.
- **Switchover**: marking old changesets as compacted (so new orphan writes go to the
  compacted changeset instead of the old one being deleted).

The commit path never blocks on this lock — it only takes the lighter `queueMu` lock
to enqueue work. See `orphan_proc.go` for the two-lock design.