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

iavlx implements the `store/v2/types.CommitMultiStore` interface, so it can be used as a
drop-in replacement in any Cosmos SDK app.

### go.mod setup

iavlx is its own Go module (`github.com/cosmos/cosmos-sdk/iavlx`) on the `aaronc/iavlx2` branch.
To depend on it, add a require and replace pointing to the branch:

```
require github.com/cosmos/cosmos-sdk/iavlx v0.0.0

replace github.com/cosmos/cosmos-sdk/iavlx => github.com/cosmos/cosmos-sdk/iavlx aaronc/iavlx2
```

### Basic integration (app.go)

```go
import (
    "github.com/cosmos/cosmos-sdk/iavlx"
    "github.com/cosmos/cosmos-sdk/server/flags"
)

// In your app constructor, before creating BaseApp:
// appOpts is the server.AppOptions passed to your app's constructor.
iavlxDir := filepath.Join(appOpts.Get(flags.FlagHome).(string), "data", "iavlx")
cms, err := iavlx.LoadCommitMultiTree(iavlxDir, iavlx.Options{}, logger)
if err != nil {
    panic(err)
}
// Prepend SetCMS so that options like pruning get applied to iavlx, not the default store.
baseAppOptions = append(
    []func(*baseapp.BaseApp){func(app *baseapp.BaseApp) { app.SetCMS(cms) }},
    baseAppOptions...,
)

bApp := baseapp.NewBaseApp(appName, logger, db, txConfig.TxDecoder(), baseAppOptions...)
```

See `simapp/app.go` for a working example.

### Optimal integration with CommitFinalizer

The above integration uses the `CommitMultiStore` compatibility layer, which works but doesn't
take full advantage of iavlx's optimistic commit pipeline. For optimal performance, `BaseApp`
should integrate `CommitFinalizer` directly — starting the commit during `FinalizeBlock` and
only waiting for the hash, while the rest of the commit (WAL fsync, commit info write, per-tree
finalization) proceeds in the background.

This was prototyped in https://github.com/cosmos/cosmos-sdk/pull/26056 but reverted as it was
removed from scope. That PR is the reference for future re-integration.

### `iavlx` CLI tool

The `iavlx` CLI (`cmd/iavlx/`) provides offline inspection and management:

- `iavlx view [dir]` — interactive TUI for browsing tree data (changesets, checkpoints,
  WAL entries, nodes, orphans, commit info). Opens files read-only via mmap; node does not
  need to be running. **Press `?` to open the built-in documentation** — it has detailed
  explainers on the on-disk format, WAL structure, node IDs, commit lifecycle, and more.
  **This is probably the best way to get oriented with the codebase internals.**
- `iavlx import --from [dir] --to [dir]` — one-time migration from iavl/v1 LevelDB format.
- `iavlx rollback [dir] --version [version]` — roll back to a specific version (node must be
  stopped). Original files are moved to a backup directory.

## Performance

## Production Readiness

I consider this code base mostly done. It was tested in multi-day devnets with pruning enabled and ran without error.
Property-based testing in this code-base checks that each individual tree and the whole multi-tree match iavl/v1
in observable hashing and data.

The remaining TODOs are mostly operational support methods, specifically:
- **Snapshot / state sync**: the per-tree primitives exist (`TreeReader.Export` for serialization,
  `Importer` for deserialization) but `Snapshot` and `Restore` on `CommitMultiStore` are not
  wired up. What remains is the multi-store orchestration and snapshot framing format.
  There is an iavl/v1 import utility in the `iavlx` CLI for one-time offline migration.
- **Store upgrades**: `LoadLatestVersionAndUpgrade`, `LoadVersionAndUpgrade`, and `LoadVersion`
  are not implemented — currently only `LoadLatestVersion` is supported.
- **SetInitialVersion**: not implemented (used for genesis export/import at non-zero heights).
- **Listeners**: `ListeningEnabled`, `AddListeners`, and `PopStateCache` are not implemented
  (used for state streaming / indexing).

## Future Optimizations

- **Smarter compaction skipping** — currently every sealed changeset is compacted. A low-hanging
  optimization is pre-scanning the orphan file to count prunable nodes before committing to
  the full compaction IO. `OrphanRewriter.Preprocess` already computes the exact prune count
  (the `deleteMap`) — this could be called as a pre-check and changesets with few prunable
  orphans could be skipped entirely. See the TODO in `compactor_proc.go`.

- **Memory-aware eviction** — the scaffolding for tracking in-memory tree size is partially in
  place but not wired up. `TreeStore.rootMemUsage` (atomic counter) exists but is never
  incremented or read. `ChangesetWriter.memUsage` accumulates the estimated memory of nodes
  written per checkpoint but nothing reads it. The orphan processor has a TODO to decrement
  `rootMemUsage` when orphans are evicted. The intended design: track memory added during
  commits, track memory freed by eviction and orphan clearing, and use this to make adaptive
  eviction decisions (e.g. evict more aggressively when memory exceeds a budget, or choose
  eviction depth dynamically instead of using a fixed threshold). Currently the `BasicEvictor`
  uses a fixed depth and is completely unaware of memory pressure.
  Implementation note: eviction counts must use atomic swap (not load/store) to avoid races
  between the evictor goroutine and orphan processor both decrementing the counter.
  More broadly, memory-aware eviction will require careful tuning and benchmarking — monitoring
  memory pressure accurately is non-trivial (Go's runtime stats, OS RSS, mmap cache effects all
  interact), and overly aggressive eviction can hurt read performance more than it helps memory.

- **Hash algorithm** — we currently use Go's `crypto/sha256` (required for iavl/v1 hash
  compatibility). There's a TODO in `node_hash.go` to benchmark `minio/sha256-simd` which uses
  hardware SHA extensions (SHA-NI on x86, crypto extensions on ARM) — early tests showed no
  improvement but this is hardware-dependent and worth revisiting on newer CPUs. Longer term,
  if hash compatibility with iavl/v1 is no longer required, a faster hash like BLAKE3 could
  significantly reduce hashing time (BLAKE3 is ~3-5x faster than SHA-256 on modern hardware).

- **Hash concurrency tuning** — the `AsyncHashScheduler` currently uses a fixed height >= 4
  threshold for deciding when to parallelize subtree hashing, and caps concurrency at NumCPU
  **per tree**. Since all trees commit in parallel, a multi-tree with many stores can massively
  oversubscribe the CPU (e.g. 20 stores × NumCPU goroutines each). A shared work pool or
  work-stealing scheduler across all trees would give better CPU utilization — large trees
  that need more hashing would naturally consume more of the pool while small trees finish
  quickly. The height threshold and concurrency cap could also be adaptive based on tree shape
  and whether hashing or tree mutation is the current bottleneck.

- **Leaf hash pre-computation batching** — the current algorithm for parallel leaf hashing in
  `prepareCommit` divides leaves into equal-sized buckets across workers. This assumes uniform
  hash cost per leaf, but leaves with very large keys or values take longer to hash. A work-
  stealing approach or dynamic batch sizing could reduce tail latency when leaf sizes vary.

- **Iterator zero-copy** — `Iterator.Next()` currently `SafeCopy`'s every key and value returned,
  even when the caller doesn't mutate them. There's a TODO in `iterator.go` to keep a stack of
  Pins and return unsafe references, only copying when the caller explicitly requests it. For
  large range scans this could eliminate millions of allocations. Related: the iterator uses
  `defer pin.Unpin()` inside the `for` loop, which delays all unpins until the function returns
  instead of releasing each pin immediately — fixing this would also improve memory release
  timing during large iterations.

- **NodeUpdates slice reuse** — `prepareCommit` allocates a new `[]NodeUpdate` slice every commit
  (TODO in `commit_tree.go`). A `sync.Pool` or reusable buffer could reduce GC pressure on
  high-TPS chains.