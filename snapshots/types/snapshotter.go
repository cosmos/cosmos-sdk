package types

import "io"

// Snapshotter is something that can take and restore snapshots, consisting of streamed binary
// chunks - all of which must be read from the channel and closed. If an unsupported format is
// given, it must return ErrUnknownFormat (possibly wrapped with fmt.Errorf).
type Snapshotter interface {
	// Snapshot takes a state snapshot.
	Snapshot(height uint64, format uint32) (<-chan io.ReadCloser, error)

	// Restore restores a state snapshot.
	Restore(height uint64, format uint32, chunks <-chan io.ReadCloser) error
}
