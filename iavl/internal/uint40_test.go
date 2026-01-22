package internal

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKVOffset(t *testing.T) {
	tests := []struct {
		name        string
		value       uint64
		inKVData    bool
		expectPanic bool
		wantStr     string
	}{
		{
			name:     "zero WAL",
			value:    0,
			inKVData: false,
			wantStr:  "0(WAL)",
		},
		{
			name:     "zero KV",
			value:    0,
			inKVData: true,
			wantStr:  "0",
		},
		{
			name:     "max WAL",
			value:    MaxKVOffset,
			inKVData: false,
			wantStr:  "549755813887(WAL)",
		},
		{
			name:     "max KV",
			value:    MaxKVOffset,
			inKVData: true,
			wantStr:  "549755813887",
		},
		{
			name:     "arbitrary WAL",
			value:    12345678,
			inKVData: false,
			wantStr:  "12345678(WAL)",
		},
		{
			name:     "arbitrary KV",
			value:    12345678,
			inKVData: true,
			wantStr:  "12345678",
		},
		{
			name:        "overflow",
			value:       1 << 39,
			inKVData:    false,
			expectPanic: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectPanic {
				require.Panics(t, func() {
					_ = NewKVOffset(tt.value, tt.inKVData)
				})
			} else {
				u := NewKVOffset(tt.value, tt.inKVData)
				require.Equal(t, tt.value, u.Offset())
				require.Equal(t, !tt.inKVData, u.IsWAL())
				require.Equal(t, tt.inKVData, u.IsKVData())
				require.Equal(t, tt.wantStr, u.String())
			}
		})
	}
}

func TestKVOffset_IsZero(t *testing.T) {
	require.True(t, NewKVOffset(0, false).IsZero())
	require.False(t, NewKVOffset(1, false).IsZero())
	require.False(t, NewKVOffset(0, true).IsZero()) // KV flag set, not zero
}
