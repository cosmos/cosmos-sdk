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
		str         string
	}{
		{
			name: "zero",
			str:  "0",
		},
		{
			name:  "max",
			value: 1<<40 - 1,
			str:   "1099511627775",
		},
		{
			name:  "arbitrary",
			value: 109951162777,
			str:   "109951162777",
		},
		{
			name:        "overflow",
			value:       1 << 40,
			expectPanic: true,
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
				require.Equal(t, tt.str, u.String())
			}
		})
	}
}
