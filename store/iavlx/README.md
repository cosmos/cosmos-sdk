# iavl

## Code Organization

### Node Types, Memory & Disk Layouts

Much of this code was influenced by memiavl and sometimes even copied directly from it.
The `NodeID` design is mainly from iavl/v2.
The `NodePointer` design introduces the possibility of doing node eviction similar to iavl/v2,
but with non-blocking thread safety using `atomic.Pointer` so that eviction can happen in the background without
blocking reads or writes.

* `node.go`: the `Node` interface which all 3 node types implement (`MemNode`, `BranchPersisted`, `LeafPersisted`)
* `mem_node.go`: in-memory node structure, new nodes always use the `MemNode` type
* `node_pointer.go`: all child references are wrapped in `NodePointer` which can point to either an in-memory node or an
  on-disk node, or both (if the node has been written and node evicted)
* `node_id.go`: defines `NodeID` (version + index + leaf) and `NodeRef` (either a `NodeID` or a node offset in the
  changeset file)
* `branch_layout.go`: defines the on-disk layout for branch nodes
* `leaf_layout.go`: defines the on-disk layout for leaf nodes
* `branch_persisted.go`: a wrapper around `BranchLayout` which implements the `Node` interface and also tracks a store
  reference
* `leaf_persisted.go`: a wrapper around `LeafLayout` which implements the `Node` interface and also tracks a store
  reference

### Tree Management & Updating

For managing tree state, we define two core types `Tree` and `CommitTree`.
We directly read from and apply updates to `Tree`s but these updates only affect the persistent state of the tree if
they are applied and committed to a `CommitTree`.

* `tree.go`: a `Tree` struct which implements the Cosmos SDK `KVStore` interface and implements the key methods (get,
  set,
  delete, commit, etc). `Tree`s can be mutated, and changes can either be committed or discarded. This is essentially an
  in-memory reference to a tree at a specific version that could be used read-only or mutated ad hoc without affecting
  the underlying persistent tree (say for instance in `CheckTx`).
* `commit_tree.go`: defines the `CommitTree` structure which manages the persistent tree state. Using `CommitTree` you
  can
  create new mutable `Tree` instance using `Branch` and decide to `Apply` its changes to the persistent tree or discard
  them. Calling `Commit` flushes changes to the underlying `TreeStore` which does all of the on disk state management
  and cleanup. In `CommitTree` we also have an asynchronous WAL writing process (optional) and maintain a background
  eviction process.
* `update.go`: types for batching changes which can later be commited or discarded
* `node_update.go` and : the code for setting and deleting nodes and doing tree rebalancing, adapted from memiavl and
  iavl/v1
* `node_hash.go`: code for computing node hashes, adapted from memiavl and iavl/v1
* `iterator.go`: implements the Cosmos SDK `Iterator` interface, adapted from memiavl and iavl/v1

### Disk State Management

### Central Coordination

These files are the central core of managing on-disk state across multiple changesets which may be in the process of
being written or compacted. **This is the most complex part of the codebase.**

* `tree_store.go`: code for dispatching read operations to the correct changeset, writing commits to new changesets,
  and coordinating background compaction and cleanup of old changesets
* `cleanup.go`: the actual background cleanup and compaction thread

#### Changeset Reading, Writing and Compaction

* `changeset_files.go`: `ChangesetFiles` represents the five files which make up a changeset:
    * `kv.log`: all of the key/value pairs in the changeset, and optionally the write-ahead log for replay (this is
      configurable)
    * `leaves.dat`: an array of `LeafLayout` structs
    * `branches.dat`: an array of `BranchLayout` structs
    * `verions.dat`: an array of `VersionInfo` structs, one for each version in the changeset
    * `info.dat`: a single `ChangesetInfo` struct which tracks metadata about the changeset including the range of
      versions
      it contains and the number of orphaned nodes
* `changeset.go`: the `Changeset` struct wraps mmap's of the five changeset files and provides
  methods for reading nodes from disk and marking them as orphaned. It includes some complex code for safely disposing
  of `Changeset` instances because we need to either 1) reopen the memmap to change its size, or 2) close the
  `Changeset` because it has been compacted and will be deleted. This is managed using pinning, a reference count, and
  atomic booleans to track eviction (the desire to dispose and delete) and disposal (the actual disposal).
* `changeset_writer.go`: code for iteratively writing changesets to disk node by node in post-order traversal order.
  Node references can either be by
  `NodeID` or offsets (offsets have been disabled due to some unresolved bugs)
* `compactor.go`: code for rewriting one or more changesets into a new compacted changeset, skipping
  orphaned nodes and updating offsets as needed (this offset rewrite code is currently buggy and disabled)

#### Helpers

* `version_info.go`: defines the on-disk layout for version info records, which track the root node and other metadata
  for
  each version
* `changeset_info.go`: defines the on-disk layout for the changeset info record, which tracks metadata
  about the entire changeset including version range and number of orphaned nodes
* `kvlog.go`: code for reading key/value pairs from the `kv.log` file
* `kvlog_writer.go`: code for writing key/value pairs to the `kv.log` file, which can be structured as a write-ahead
  operation log for replay and crash recovery (reply and recovery aren't implemented yet)
* `mmap.go`: the `MmapFile` mem-map wrapper
* `writer.go`: `FileWriter` and `StructWriter` wrappers for writing raw bytes and structs to files
* `reader.go`: `StructMap` and `NodeMap` wrappers for representing memory-mapped arrays of structs and nodes

### Multi-tree Management

* `multi_tree.go`: wraps multiple `Tree`s into a `MultiTree` which provides a mutable way to write a tree without
  committing the changes to the persistent tree immediately (can be discarded)
* `commit_multi_tree.go`: wraps multiple `CommitTree`s into a `CommitMultiTree` which provides a way to create mutable
  `MultiTree`s and commit their changes to the underlying persistent trees (or discard them). This can eventually
  implement `RootMultiStore` and replace the SDK's store package. `CommitMultiTree` makes the optimization of running
  `Commit` in parallel across all `CommitTree`s which could improve performance.

### Options

Options are mantained by the `Options` struct in `options.go`. Many options have a getter which uses a default value if
the option is not set.

The main options we're controlling now are:

* `WriteWAL`: whether we write all updates to the kv-log as a replayable write-ahead log (WAL). If this is enabled we
  will fsync the WAL either asynchronously or synchronously (based on the `WalSyncBuffer` option). Enabling WAL could
  actually improve performance because we asynchronously write key/value data in advance of `CommitTree.Commit` being
  called.
* `EvictDepth`: the depth of the tree beyond which we will evict nodes from memory as soon as they are on disk. This is
  the main lever for controlling memory usage. Using more memory could improve performance.
* `RetainVersions`: the number of recent versions to retain when we are compacting. Eventually we also want to enable
  some sort of snapshot-based compaction (retaining full trees every N versions).
* `MinCompactionSeconds`: the minimum number of seconds to wait before starting a new compaction run (note that this
  currently includes the time it takes to compact).
* `CompactWAL`: whether to compact the WAL when we are compacting changesets. In the future, we can distinguish between
  compacting the WAL before our first checkpoint and retaining it after the first checkpoint.
* `ChangesetMaxTarget`: the size of a changeset after which we will roll over to a new changeset for the next version.
* `CompactionMaxTarget`: the target size of a compacted changeset. When adding a new changeset into our compaction will
  stay below this number, we will join multiple changesets into a single compacted changeset.
* `CompactionOrphanRatio`: the ratio of orphaned nodes in a changeset beyond which we will trigger it for early
  compaction (used together with `CompactionOrphanAge`)
* `CompactionOrphanAge`: the average age of orphaned nodes in a changeset beyond which we will trigger it for early
  compaction (used together with `CompactionOrphanRatio`)
* `CompactAfterVersions`: the number of versions after which we will trigger a compaction when any orphans are present,
  measured in versions since the last compaction.
* `ReaderUpdateInterval`: when writing multiple versions to a changeset, the number of versions after which we will open
  the changeset for reading even if it has not been completed, so that readers can access the latest versions sooner and
  flush memory. Set this to a shorter interval if we want to constrain memory usage more tightly and longer if we want
  to reduce the number of times memmaps are re-opened for reading.

### Utilities

* `dot_graph.go`: code for exporting trees to Graphviz dot graph format for visualization
* `verify.go`: code for verifying tree integrity

### Tests

* `tree_test.go`: the only tests we have so far. These do, however, use property-based testing so we are generating
  random operation sets, applying them to both iavlx and iavl/v1 trees. At each step, we confirm that behavior is
  identical, including verification of hashes and verifying that invariants are maintained.