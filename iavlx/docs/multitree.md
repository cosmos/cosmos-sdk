# Multi-Tree

A Cosmos SDK app's state is stored as a collection of independent IAVL trees — one per
module store (bank, staking, auth, etc.). iavlx calls this collection a **multi-tree**.

## Directory layout

```
<data-dir>/
├── commit_info/        — one file per committed version (see commit-info.md)
│   ├── 1
│   ├── 2
│   └── ...
└── stores/
    ├── auth.iavl/      — each store is its own IAVL tree directory
    │   ├── 1/          — changeset directories (see changeset.md)
    │   │   ├── wal.log
    │   │   ├── checkpoints.dat
    │   │   ├── branches.dat
    │   │   ├── leaves.dat
    │   │   ├── kv.dat
    │   │   └── orphans.dat
    │   └── 1053/
    ├── bank.iavl/
    └── staking.iavl/
```

## How commits work across trees

All trees commit in parallel — each tree runs its own WAL write, tree mutation, and hash
computation independently. The `CommitFinalizer` coordinates: it waits for all per-tree
hashes, computes the combined multi-tree hash, then writes the commit info file which
makes the commit durable across all trees atomically. See `commit-lifecycle.md`.

## Compaction

Each tree compacts independently. Compaction is triggered after each commit if conditions
are met (see `compactIfNeeded` in `commit_multi_tree.go`). Only one compaction runs at a
time across the entire multi-tree.