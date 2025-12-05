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
		str     string
	}{
		{
			name: "leaf1_1",
			leaf: true, version: 1, index: 1,
			str: "NodeID{leaf:true, version:1, index:1}",
		},
		{
			name: "branch2_3", version: 2, index: 3,
			str: "NodeID{leaf:false, version:2, index:3}",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			id := NewNodeID(test.leaf, test.version, test.index)
			require.Equal(t, test.leaf, id.IsLeaf())
			require.Equal(t, test.index, id.FlagIndex.Index())
			require.Equal(t, test.version, id.Version)
			require.Equal(t, test.str, id.String())
		})
	}
}

func TestNodeID_IsEmpty(t *testing.T) {
	require.True(t, NodeID{}.IsEmpty())
	require.False(t, NewNodeID(true, 1, 1).IsEmpty())
}
