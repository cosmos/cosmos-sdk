# Commit Lifecycle

IAVLX makes every attempt possible to reduce commit latency.
There are three things that need to happen when the tree is updated and committed:
1. update the tree
2. compute the root hash
3. write changes to disk

Significant portions of these operations can actually be decoupled so that work can proceed in parallel.

```
                    B-tree cache (sorted updates)
                                 │
              ┌──────────────────┼─── ───────────────┐
              │                  │                   │
              ▼                  ▼                   ▼
       ┌──────────────┐   ┌──────────────┐   ┌───────────────┐
       │ Leaf Hashing │   │Tree Mutation │   │  WAL Writing  │
       │  (parallel   │   │ (sequential) │   │ (append-only) │
       │   batches)   │   │              │   │               │
       └──────┬───────┘   └──────┬───────┘   └───────┬───────┘
              │                  │                   │
              └────────┬─────────┘                   │
                       │ both complete               │
                       ▼                             │
                ┌──────────────┐                     │
                │ Root Hashing │                     │
                │  (parallel   │                     │
                │   subtrees)  │                     │
                └──────┬───────┘                     │
                       │                             │
                       └─────────────┬───────────────┘
                                     │ both complete
                                     ▼
                              ┌──────────────┐
                              │   Caller     │
                              │  confirms or │
                              │  rolls back  │
                              └──────┬───────┘
                                     │ confirmed
                          ┌──────────┴──────────┐
                          │                     │
                          ▼                     ▼
                ┌────────────────┐   ┌────────────────┐
                │ Update version │   │Write CommitInfo│
                │& root pointer  │   │  (durable,     │
                │  (in-memory)   │   │   multi-tree)  │
                └───────┬────────┘   └───────┬────────┘
                        │                    │
                        └─────────┬──────────┘
                                  │
                                  ▼
                        ── return to caller ──
                                  │
                                  ▼ (background)
                         ┌─────────────────┐
                         │  Checkpoint     │
                         │  (non-blocking, │
                         │   periodic)     │
                         └─────────────────┘
```

## The sorted update list

All updates from block execution settle into a sorted B-tree cache layer before commit.
Because AVL balancing makes tree structure dependent on insertion order, updates must be applied
in sorted key order for determinism — the B-tree provides this for free.

## Three concurrent operations

From this sorted list, three things kick off simultaneously:

- **Leaf hashing** — each new leaf can be hashed independently, so we hash them in parallel
  batches before the tree is even mutated.
- **WAL writing** — the updates are appended to the write-ahead log on disk. Because it's
  append-only, it's fast and can be rolled back by simple truncation.
- **Tree mutation** — updates are applied to the in-memory tree sequentially (AVL rebalancing
  requires this). For large trees, this can be the slowest step due to random disk reads
  for nodes not in memory.

Each IAVL tree in a multi-tree runs these three operations independently and in parallel.

## Root hashing

Once tree mutation and leaf hashing are both done, root hashing begins. Unlike tree mutation,
hashing can run in parallel — left and right subtree hashes are independent. We don't need
to wait for the WAL to finish for this step.

## Confirm or rollback

After root hashing and WAL writing are both complete, we wait for the caller to confirm or
roll back. Up to this point, nothing is committed — rolling back just truncates the WAL.

## Finalization

On confirmation:
- The **commit info file** is written, fsynced, and atomically renamed — this is the durability
  boundary. All per-tree WALs must be fsynced before the rename (see `commit_finalizer.go`).
- The **in-memory root pointer** is swapped so readers see the new version.
- A **checkpoint** may be kicked off in the background.

At this point, we return to the caller. Only the WAL and commit info are guaranteed on disk.
This is sufficient for durability — on restart, the tree is reconstructed from checkpoint + WAL
replay. Checkpoints are an optimization for faster startup, written periodically in the background
at configurable intervals (default every 100 versions). If checkpoint writing fails, startup
detects the error, cleans up, and replays more WAL instead. See `checkpoint.md` for details.