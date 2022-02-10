package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPruningOptions_Validate(t *testing.T) {
	testCases := []struct {
		keepRecent uint64
		interval   uint64
		expectErr  bool
	}{
		{100, 10, false}, // default
		{0, 10, false},   // everything
		{0, 0, false},    // nothing
		{100, 0, true},   // invalid interval
	}

	for _, tc := range testCases {
		po := NewPruningOptions(tc.keepRecent, tc.interval)
		err := po.Validate()
		require.Equal(t, tc.expectErr, err != nil, "options: %v, err: %s", po, err)
	}
}
