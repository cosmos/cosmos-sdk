package query

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClampPageRequestLimit(t *testing.T) {
	tests := []struct {
		name     string
		input    *PageRequest
		expected *PageRequest
	}{
		{
			name: "normal values within limits",
			input: &PageRequest{
				Offset: 10,
				Limit:  50,
			},
			expected: &PageRequest{
				Offset: 10,
				Limit:  50,
			},
		},
		{
			name:  "nil input returns empty PageRequest",
			input: nil,
			expected: &PageRequest{
				Offset: 0,
				Limit:  0,
			},
		},
		{
			name: "offset + limit would overflow",
			input: &PageRequest{
				Offset: math.MaxUint64 - 10,
				Limit:  50,
			},
			expected: &PageRequest{
				Offset: math.MaxUint64 - 10,
				Limit:  10,
			},
		},
		{
			name: "extreme limit overflow",
			input: &PageRequest{
				Offset: 100,
				Limit:  math.MaxUint64,
			},
			expected: &PageRequest{
				Offset: 100,
				Limit:  math.MaxUint64 - 100,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := clampPageRequestLimit(tc.input)
			require.Equal(t, tc.expected.Offset, result.Offset)
			require.Equal(t, tc.expected.Limit, result.Limit)
		})
	}
}
