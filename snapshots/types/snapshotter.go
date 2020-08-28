package types

import "io"

// Snapshotter is something that can create and restore snapshots, consisting of streamed binary
// chunks - all of which must be read from the channel and closed. If an unsupported format is
// given, it must return ErrUnknownFormat (possibly wrapped with fmt.Errorf).
type Snapshotter interface {
	// Snapshot creates a state snapshot, returning a channel of snapshot chunk readers.
	Snapshot(height uint64, format uint32) (<-chan io.ReadCloser, error)

	// Restore restores a state snapshot, taking snapshot chunk readers as input.
	Restore(height uint64, format uint32, chunks <-chan io.ReadCloser) error
}
