package types

import "math"

// SnapshotOptions defines the snapshot strategy used when determining which
// heights are snapshotted for state sync.
type SnapshotOptions struct {
	// Interval defines at which heights the snapshot is taken.
	Interval uint64

	// KeepRecent defines how many snapshots to keep in heights.
	KeepRecent uint32
}

// NewSnapshotOptions creates and returns a new SnapshotOptions instance.
// It panics if the interval exceeds the maximum value for int64.
func NewSnapshotOptions(interval uint64, keepRecent uint32) SnapshotOptions {
	if interval > math.MaxInt64 {
		panic("interval must not exceed max int64")
	}
	return SnapshotOptions{
		Interval:   interval,
		KeepRecent: keepRecent,
	}
}
