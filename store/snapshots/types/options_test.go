package types

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSnapshotOptions(t *testing.T) {
	specs := map[string]struct {
		srcInterval uint64
		expPanic    bool
	}{
		"valid ": {
			srcInterval: 1,
		},
		"max interval ": {
			srcInterval: math.MaxInt64,
		},
		"exceeds max interval ": {
			srcInterval: math.MaxInt64 + 1,
			expPanic:    true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			if spec.expPanic {
				assert.Panics(t, func() {
					NewSnapshotOptions(spec.srcInterval, 2)
				})
				return
			}
			NewSnapshotOptions(spec.srcInterval, 10)
		})
	}
}
