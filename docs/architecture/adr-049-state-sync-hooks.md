# ADR 049: State Sync Hooks

## Changelog

- Jan 19, 2022: Initial Draft

## Status

Draft, Under Implementation

## Abstract

This ADR provides hooks for app modules to publish snapshot of additional state(outside of IAVL tree) for state-sync.

## Context

New clients use state-sync to download snapshots of module state from peer nodes. Currently, the snapshot consists of a
stream of `SnapshotStoreItem` and `SnapshotIAVLItem`, which means for app modules that maintain their states outside of
the IAVL tree, they can not add their states to the snapshot stream for state-sync.

## Decision

A simple proposal based on our existing implementation is that, we can add two new message types: `SnapshotExtensionMeta` 
and `SnapshotExtensionPayload`, and they are appended to the existing multi-store stream with `SnapshotExtensionMeta` 
acting as a delimiter between extensions. Even for modules that maintain the data outside of the tree, for determinism we
require that the hash of the external data should be posted in the IAVL tree. As the chunk hashes should be able to ensure 
data integrity, we don't need a delimiter to mark the end of the snapshot stream.

Besides, we provide `Snapshotter` and `ExtensionSnapshotter` interface for modules to implement snapshotters, which will handle both taking 
snapshot and the restoration. Each module could have mutiple snapshotters, and for modules with additional state, they should
implement `ExtensionSnapshotter` as extension snapshotters. When setting up the application, the snapshot `Manager` should call 
`RegisterExtensions([]ExtensionSnapshotter…)` to register all the extension snapshotters.

```proto
// SnapshotItem is an item contained in a rootmulti.Store snapshot.
message SnapshotItem {
  // item is the specific type of snapshot item.
  oneof item {
    SnapshotStoreItem        store             = 1;
    SnapshotIAVLItem         iavl              = 2 [(gogoproto.customname) = "IAVL"];
    SnapshotExtensionMeta    extension         = 3;
    SnapshotExtensionPayload extension_payload = 4;
  }
}

// SnapshotStoreItem contains metadata about a snapshotted store.
message SnapshotStoreItem {
  string name = 1;
}

// SnapshotIAVLItem is an exported IAVL node.
message SnapshotIAVLItem {
  bytes key     = 1;
  bytes value   = 2;
  int64 version = 3;
  int32 height  = 4;
}

// SnapshotExtensionMeta contains metadata about an external snapshotter.
// One module may need multiple snapshotters, so each module may have multiple SnapshotExtensionMeta.
message SnapshotExtensionMeta {
  string name   = 1;
  // format is used within the snapshotter/namespace, not global one for all modules
  uint32 format = 2;
}

// SnapshotExtensionPayload contains payloads of an external snapshotter.
message SnapshotExtensionPayload {
  bytes payload = 1;
}
```

The snapshot stream would look like this:

```go
// multi-store snapshot
{SnapshotStoreItem | SnapshotIAVLItem, ...}
// extension1 snapshot
SnapshotExtensionMeta
{SnapshotExtensionPayload, ...}
// extension2 snapshot
SnapshotExtensionMeta
{SnapshotExtensionPayload, ...}
```

Add `extensions` field to snapshot `Manager` for extension snapshotters. `multistore` snapshotter is a special one and it doesn't need a name because it is always placed at the beginning of the binary stream.

```go
type Manager struct {
	store      *Store
	multistore types.Snapshotter
	extensions map[string]types.ExtensionSnapshotter
    mtx                sync.Mutex
	operation          operation
	chRestore          chan<- io.ReadCloser
	chRestoreDone      <-chan restoreDone
	restoreChunkHashes [][]byte
	restoreChunkIndex  uint32
}
```

For extension snapshotters that implement the `ExtensionSnapshotter` interface, their names should be registered to the snapshot `Manager` by 
calling `RegisterExtensions` when setting up the application. And the snapshotters will handle both taking snapshot and restoration.

```go
// RegisterExtensions register extension snapshotters to manager
func (m *Manager) RegisterExtensions(extensions ...types.ExtensionSnapshotter) error 
```

Remain `Snapshotter` interface for `multistore` snapshotter. No need for `SnapshotFormat` and `SupportedFormats` as for `multistore` snapshotter
formats are handled in a higher level.
<ul>
<li> removal of format parameter: `Snapshotter` chooses a format autonomously, not pass in from the caller. </li>
<ul>

```go
// Snapshotter is something that can create and restore snapshots, consisting of streamed binary
// chunks - all of which must be read from the channel and closed. If an unsupported format is
// given, it must return ErrUnknownFormat (possibly wrapped with fmt.Errorf).
type Snapshotter interface {
	// Snapshot writes snapshot items into the protobuf writer.
	Snapshot(height uint64, protoWriter protoio.Writer) error

	// Restore restores a state snapshot from the protobuf items read from the reader.
	// If the ready channel is non-nil, it returns a ready signal (by being closed) once the
	// restorer is ready to accept chunks.
	Restore(height uint64, format uint32, protoReader protoio.Reader) (SnapshotItem, error)
}
```

Add `ExtensionSnapshotter` interface for extension snapshotters, and three more function signatures: `SnapshotFormat()` `SupportedFormats()` and
`SnapshotName()` are added.

```go
// ExtensionSnapshotter is an extension Snapshotter that is appended to the snapshot stream.
// ExtensionSnapshotter has an unique name and manages it's own internal formats.
type ExtensionSnapshotter interface {
	Snapshotter

	// SnapshotName returns the name of snapshotter, it should be unique in the manager.
	SnapshotName() string

	// SnapshotFormat returns the default format used to take a snapshot.
	SnapshotFormat() uint32

	// SupportedFormats returns a list of formats it can restore from.
	SupportedFormats() []uint32
}
```

## Consequences

As a result of this implementation, we are able to create snapshots of binary chunk stream for the state that we maintain outside of the IAVL Tree, CosmWasm blobs for example. And new clients are able to fetch sanpshots of state for all modules that have implemented the corresponding interface from peer nodes. 


### Backwards Compatibility

This ADR introduces new proto message types, add field for extension snapshotters to snapshot `Manager`, add new function signatures `SnapshotFormat()`, `SupportedFormats()` to `Snapshotter` interface and add new `ExtensionSnapshotter` interface, so this is not backwards compatible.

### Positive

State maintained outside of IAVL tree like CosmWasm blobs can create snapshots by implementing extension snapshotters, and being fetched by new clients via state-sync.

### Negative

// Todo

### Neutral

All modules that maintain state outside of IAVL tree need to implement `ExtensionSnapshotter` and the snapshot `Manager` need to call `RegisterExtensions` when setting up the application.

## Further Discussions

While an ADR is in the DRAFT or PROPOSED stage, this section should contain a summary of issues to be solved in future iterations (usually referencing comments from a pull-request discussion).
Later, this section can optionally list ideas or improvements the author or reviewers found during the analysis of this ADR.

## Test Cases [optional]

Test cases for an implementation are mandatory for ADRs that are affecting consensus changes. Other ADRs can choose to include links to test cases if applicable.

## References

- https://github.com/cosmos/cosmos-sdk/pull/10961
- https://github.com/cosmos/cosmos-sdk/issues/7340
- https://hackmd.io/gJoyev6DSmqqkO667WQlGw
