package pbtime

import (
	"time"

	durpb "google.golang.org/protobuf/types/known/durationpb"
	tspb "google.golang.org/protobuf/types/known/timestamppb"
)

// IsZero returns true when t is nil or is zero unix timestamp (1970-01-01)
func IsZero(t *tspb.Timestamp) bool {
	return t == nil || t.Nanos == 0 && t.Seconds == 0
}

// Commpare t1 and t2 and returns -1 when t1 < t2, 0 when t1 == t2 and 1 otherwise.
// Panics if t1 or t2 is nil
func Compare(t1, t2 *tspb.Timestamp) int {
	if t1.Seconds == t2.Seconds && t1.Nanos == t2.Nanos {
		return 0
	}
	if t1.Seconds < t2.Seconds || t1.Seconds == t2.Seconds && t1.Nanos < t2.Nanos {
		return -1
	}
	return 1
}

// AddStd returns a new timestamp with value t + d, where d is stdlib Duration
// If t is nil then nil is returned
func AddStd(t *tspb.Timestamp, d time.Duration) *tspb.Timestamp {
	if t == nil {
		return nil
	}
	if d == 0 {
		t2 := *t
		return &t2
	}
	t2 := t.AsTime()
	return tspb.New(t2.Add(d))
}

const second = int32(time.Second)

// Add returns a new timestamp with value t + d, where d is protobuf Duration
// If t is nil then nil is returned. Panics on overflow.
func Add(t *tspb.Timestamp, d durpb.Duration) *tspb.Timestamp {
	if t == nil {
		return nil
	}
	if d.Seconds == 0 && d.Nanos == 0 {
		t2 := *t
		return &t2
	}
	t2 := tspb.Timestamp{
		Seconds: t.Seconds + d.Seconds,
		Nanos:   t.Nanos + d.Nanos,
	}
	if t2.Nanos >= second {
		t2.Nanos -= second
		t2.Seconds++
	}
	return &t2
}
