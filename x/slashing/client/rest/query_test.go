package rest

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAdjustPagination(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name  string
		size  uint
		page  uint
		limit uint
		start uint
		end   uint
	}{
		{"Ok", 3, 0, 1, 0, 1},
		{"Limit too big", 3, 1, 5, 0, 3},
		{"Page over limit", 3, 2, 3, 0, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end := adjustPagination(tt.size, tt.page, tt.limit)
			require.Equal(t, tt.start, start)
			require.Equal(t, tt.end, end)
		})
	}
}
