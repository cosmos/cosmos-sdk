package internal

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNodeID(t *testing.T) {
	tests := []struct {
		name    string
		leaf    bool
		version uint32
		index   uint32
	}{
		{
			"leaf1_1",
			true, 1, 1,
		},
		{
			"branch2_3",
			false, 2, 3,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			id := NewNodeID(test.leaf, test.version, test.index)
			require.Equal(t, test.leaf, id.IsLeaf())
			require.Equal(t, test.index, id.FlagIndex.Index())
			require.Equal(t, test.version, id.Version)
		})
	}
}
