package types

import (
	protoio "github.com/gogo/protobuf/io"
)

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

// ExtensionSnapshotter is an extension Snapshotter that is appended to the snapshot stream.
// ExtensionSnapshotter has an unique name and manages it's own internal formats.
type ExtensionSnapshotter interface {
	Snapshotter

	// SnapshotName returns the name of snapshotter, it should be unique in the manager.
	SnapshotName() string

	// SnapshotFormat returns the default format the extension snapshotter use to encode the
	// payloads when taking a snapshot.
	// It's defined within the extension, different from the global format for the whole state-sync snapshot.
	SnapshotFormat() uint32

	// SupportedFormats returns a list of formats it can restore from.
	SupportedFormats() []uint32
}
