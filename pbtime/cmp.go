package pbtime

import (
	"time"

	tspb "google.golang.org/protobuf/types/known/timestamppb"
)

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

func Add(t *tspb.Timestamp, d time.Duration) *tspb.Timestamp {
	if d == 0 {
		t2 := *t
		return &t2
	}
	t2 := t.AsTime()
	return tspb.New(t2.Add(d))
}
