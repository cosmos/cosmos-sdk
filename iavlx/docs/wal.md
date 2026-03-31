# Write-Ahead Log (WAL)

The WAL (`wal.log`) is the durability mechanism and doubles as primary key-value storage for nodes.
All set/delete operations are appended and fsynced before a commit is considered durable.
On crash recovery, the tree is reconstructed by replaying WAL entries from the latest checkpoint.

## Binary format

Each entry starts with a 1-byte type tag. Flags are OR'd into the high bits.

```
 Type byte layout:
   bits 0-5: entry type
   bit 6:    WALFlagCheckpoint (on commit entries: a checkpoint was scheduled)
   bit 7:    WALFlagCachedKey  (on set/delete entries: key is a 5-byte offset, not inline)
```

### Entry types

**Start (0x00)** — first entry in a WAL file, written once:
```
  [type: 1] [version: varint]
```

**Set (0x01)** — key-value pair written:
```
  [type: 1] [key: len-prefixed bytes] [value: len-prefixed bytes]
```

**Delete (0x02)** — key removed:
```
  [type: 1] [key: len-prefixed bytes]
```

**Commit (0x03)** — version boundary:
```
  [type: 1] [version: varint]
```

A version's entries are always: zero or more Set/Delete, then exactly one Commit.

### Key compression

When the same key is written multiple times (common for frequently updated keys), the WAL
deduplicates it. The `WALWriter` maintains an in-memory key cache mapping key bytes to their
first WAL offset. On subsequent writes of the same key, if the key is >= 5 bytes, the writer
sets the `WALFlagCachedKey` flag and writes a 5-byte (Uint40) offset instead of the full key:

```
  Set with cached key:
  [type: 0x81] [key offset: 5 bytes LE] [value: len-prefixed bytes]
```

Keys shorter than 5 bytes are always written inline (the offset would be larger than the key).

## WAL as blob storage

Checkpoint nodes reference key/value data by WAL byte offset instead of duplicating it in
`kv.dat`. During compaction, if a WAL entry is pruned but nodes still reference its data,
the blob migrates to `kv.dat` (see `remapBlob` in `compactor.go`).

## Rollback

Append-only structure makes rollback a file truncation to the previous commit offset — O(1).
On crash recovery, entries for at most one version beyond the committed version are truncated
automatically. More than one is treated as corruption (see `wal_replay.go`).