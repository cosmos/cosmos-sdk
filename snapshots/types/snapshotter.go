package types

import "io"

// Snapshotter is something that can take and restore snapshots. Snapshot data consists of streamed
// chunks, all of which must be read from the channel and closed. If an unsupported format is given,
// it must return ErrUnknownFormat.
type Snapshotter interface {
	Snapshot(height uint64, format uint32) (<-chan io.ReadCloser, error)
	Restore(height uint64, format uint32, chunks <-chan io.ReadCloser) error
}
