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
			require.Equal(t, test.index, id.Index())
			require.Equal(t, test.version, id.Version())
			require.Equal(t, test.str, id.String())
		})
	}
}

func TestNodeID_IsEmpty(t *testing.T) {
	require.True(t, NodeID{}.IsEmpty())
	require.False(t, NewNodeID(true, 1, 1).IsEmpty())
}

func TestNodeID_Equal(t *testing.T) {
	tests := []struct {
		name  string
		a, b  NodeID
		equal bool
	}{
		{
			name:  "both empty",
			a:     NodeID{},
			b:     NodeID{},
			equal: true,
		},
		{
			name:  "same leaf",
			a:     NewNodeID(true, 1, 1),
			b:     NewNodeID(true, 1, 1),
			equal: true,
		},
		{
			name:  "same branch",
			a:     NewNodeID(false, 2, 3),
			b:     NewNodeID(false, 2, 3),
			equal: true,
		},
		{
			name:  "different version",
			a:     NewNodeID(true, 1, 1),
			b:     NewNodeID(true, 2, 1),
			equal: false,
		},
		{
			name:  "different index",
			a:     NewNodeID(true, 1, 1),
			b:     NewNodeID(true, 1, 2),
			equal: false,
		},
		{
			name:  "different leaf flag",
			a:     NewNodeID(true, 1, 1),
			b:     NewNodeID(false, 1, 1),
			equal: false,
		},
		{
			name:  "empty vs non-empty",
			a:     NodeID{},
			b:     NewNodeID(true, 1, 1),
			equal: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.equal, tt.a.Equal(tt.b))
			require.Equal(t, tt.equal, tt.b.Equal(tt.a)) // symmetry
		})
	}
}
