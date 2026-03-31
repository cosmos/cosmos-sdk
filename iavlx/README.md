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

### Basic integration (app.go)

```go
import "github.com/cosmos/cosmos-sdk/iavlx"

// In your app constructor, before creating BaseApp:
iavlxDir := filepath.Join(homePath, "data", "iavlx")
cms, err := iavlx.LoadCommitMultiTree(iavlxDir, iavlx.Options{}, logger)
if err != nil {
    panic(err)
}
// Prepend SetCMS so that options like pruning get applied to iavlx, not the default store.
baseAppOptions = append(
    []func(*baseapp.BaseApp){func(app *baseapp.BaseApp) { app.SetCMS(cms) }},
    baseAppOptions...,
)
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

- a low hanging fruit optimization is checking orphan metrics before compacting CHANGESETs