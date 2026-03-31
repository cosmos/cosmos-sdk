# Commit Info

The `commit_info/` directory stores one file per committed version for a mult-tree, named by version number
(e.g. `1`, `2`, `1053`). Each file records which stores participated in that commit and
their root hashes. The presence of a commit info file IS the durability signal for the multi-tree — 
if the file exists, that version was committed. If it doesn't, it wasn't.

## File format

The file is written in two phases (see `commit_finalizer.go`):

**Header** (fsynced before the atomic rename):
```
  [version: 4 bytes LE uint32]
  [timestamp: 8 bytes LE int64, unix nanos]
  [store count: 4 bytes LE uint32]
  for each store:
    [name length: varint] [name: bytes]
```

**Footer** (appended after rename, NOT fsynced):
```
  for each store:
    [hash length: varint] [hash: bytes]
```

The header is the durable part. Hashes in the footer are best-effort — if the process
crashes before the footer is written, startup loads the commit info without hashes and
recomputes them from the trees. This two-phase design allows the expensive hash fsync
to be skipped (the WALs already guarantee recoverability).

## Durability protocol

The commit info file is the last thing written during a commit, AFTER all per-tree WALs
have been fsynced. The write sequence is:

1. Write header to `.pending.<version>`, fsync
2. Wait for all per-tree WAL writes to complete
3. Rename `.pending.<version>` → `<version>` (atomic)
4. Fsync parent directory

After step 4, the commit is crash-recoverable. See `writeCommitInfoHeader` in
`commit_finalizer.go` for details on why fsync-before-rename is necessary.

## Cleanup

On startup, any `.pending.*` files are deleted — they represent interrupted commits
that never became durable.

During rollback (`iavlx rollback`), commit info files beyond the target version are
moved to the backup directory.