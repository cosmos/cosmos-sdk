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
		wantStr     string
	}{
		{
			name:    "zero",
			value:   0,
			wantStr: "0",
		},
		{
			name:    "max",
			value:   MaxUint40,
			wantStr: "1099511627775",
		},
		{
			name:    "arbitrary",
			value:   12345678,
			wantStr: "12345678",
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
				require.Equal(t, tt.value, u.ToUint64())
				require.Equal(t, tt.wantStr, u.String())
			}
		})
	}
}

func TestUint40_IsZero(t *testing.T) {
	require.True(t, NewUint40(0).IsZero())
	require.False(t, NewUint40(1).IsZero())
}
