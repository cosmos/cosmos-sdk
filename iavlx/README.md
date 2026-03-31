# iavlx

## Overview

iavlx is a high-performance ACID merkle-tree database based on immutable AVL+ trees and is hash compatible with
https://github.com/cosmos/iavl.

It improves upon the earlier iavl design with:
- a custom disk format which:
  - is append-only, requiring minimal disk IO for writing and compaction
  - allows for fast node-to-node read traversals - O(1) in most cases, degrading to O(log n) worst-case
- optimal concurrency ordering which:
  - utilizes as many CPU cores as possible to split up work within a single tree and across tree
  - reduces commit latency as much as possible, supporting an optional optimistic commit or rollback path to take advantage of ABCI optimistic execution
  - allows read concurrency to scale across threads
  - allows for pruning/compaction to proceed with zero blockage to active commits and minimal resource overhead

## Usage

## Performance

## Production Readiness

## Future Optimizations
