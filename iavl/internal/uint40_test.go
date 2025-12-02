package internal

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUint40(t *testing.T) {
	tests := []struct {
		name        string
		value       uint64
		expectPanic bool
	}{
		{
			"zero",
			0,
			false,
		},
		{
			"max",
			1<<40 - 1,
			false,
		},
		{
			"arbitrary",
			109951162777,
			false,
		},
		{
			"overflow",
			1 << 40,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectPanic {
				require.Panics(t, func() {
					_ = NewUint40(tt.value)
				})
			} else {
				u := NewUint40(tt.value)
				got := u.ToUint64()
				require.Equal(t, tt.value, got)
			}
		})
	}
}
