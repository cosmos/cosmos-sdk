# iavlx

## Overview

iavlx is a high-performance ACID merkle-tree database based on immutable AVL+ trees and is hash compatible with
https://github.com/cosmos/iavl.

It improves upon the earlier iavl design with:
- **no external database dependency** — manages its own on-disk format with append-only WALs,
  periodic checkpoints, and background compaction (no LevelDB/RocksDB/PebbleDB)
- **custom disk format** which:
  - is append-only, requiring minimal disk IO for writing and compaction
  - allows for fast node-to-node read traversals — O(1) within the same changeset via direct
    file offsets, O(log log n) average cross-changeset via interpolation search
- **optimal concurrency** which:
  - utilizes as many CPU cores as possible to split up work within a single tree and across trees
  - reduces commit latency as much as possible, supporting an optional optimistic commit or rollback path to take advantage of ABCI optimistic execution
  - allows read concurrency to scale across threads
  - allows for pruning/compaction to proceed with zero blockage to active commits and minimal resource overhead
- **WAL-based durability and atomicity** — commits are durable once the WAL is fsynced and
  the commit info file is atomically renamed; crash recovery replays from the latest checkpoint.
  Unlike iavl/v1 which relied on the external database for durability and had no atomic
  multi-tree commit, iavlx commits all trees atomically via a single commit info file

## Usage

## Performance

## Production Readiness

## Future Optimizations
