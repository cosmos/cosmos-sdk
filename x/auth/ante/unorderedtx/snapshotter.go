package unorderedtx

import (
	snapshot "cosmossdk.io/store/snapshots/types"
)

var _ snapshot.ExtensionSnapshotter = &Snapshotter{}

// SnapshotFormat defines the snapshot format of exported unordered transactions.
// No protobuf envelope, no metadata.
const SnapshotFormat = 1

type Snapshotter struct {
	m *Manager
}

func (s *Snapshotter) SnapshotName() string {
	panic("not implemented!")
}

func (s *Snapshotter) SnapshotFormat() uint32 {
	panic("not implemented!")
}

func (s *Snapshotter) SupportedFormats() []uint32 {
	panic("not implemented!")
}

func (s *Snapshotter) SnapshotExtension(height uint64, payloadWriter snapshot.ExtensionPayloadWriter) error {
	panic("not implemented!")
}

func (s *Snapshotter) RestoreExtension(height uint64, format uint32, payloadReader snapshot.ExtensionPayloadReader) error {
	panic("not implemented!")
}
