# Changesets

A changeset represents the on disk data for a range of versions in an iavl tree.

## Directory Naming

### Original Changeset

An "original changeset" is a changeset written during normal execution before any compaction. It is stored in a directory named by the **start version** of the changeset, ex. `1`, `1053`, etc.In an original changeset, the write-ahead log (WAL) will always span from the start version to the end version. We can determine the **end version** by reading the WAL to the end.

### Compacted Changeset

A "compacted changeset" is a changeset that was rewritten from one or more original changesets with some data pruned during the process. It is stored as a directory that encodes the **start version**, **end version**, and **compacted version** as `<start>-<end>.<compacted>`, ex. `1-2501.3000`. The **compacted version** is simply the version at which we initiated the compaction operation. A compacted changeset may or may not have WAL entries spanning the full range of versions (likely some entries will have been pruned).

## Files

Changesets consist of the following files:
- `wal.log`  - the write-ahead log (WAL) which contains all the set, delete, and commit operations for each version
- `checkpoints.dat`: checkpoint metadata
- `branches.dat`: branch node data for checkpoints
- `leaves.dat`: leaf node data for checkpoints
- `kv.dat`: key-value data which is referenced by checkpoint nodes, but that isn't in the WAL
- `orphans.dat`: a log of nodes that were orphaned or removed from the tree, and the version at which they were removed