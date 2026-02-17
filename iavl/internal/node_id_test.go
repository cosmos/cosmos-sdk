package internal

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNodeID(t *testing.T) {
	tests := []struct {
		name       string
		leaf       bool
		checkpoint uint32
		index      uint32
		str        string
	}{
		{
			name: "leaf1_1",
			leaf: true, checkpoint: 1, index: 1,
			str: "L:1:1",
		},
		{
			name: "branch2_3", checkpoint: 2, index: 3,
			str: "B:2:3",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			id := NewNodeID(test.leaf, test.checkpoint, test.index)
			require.Equal(t, test.leaf, id.IsLeaf())
			require.Equal(t, test.index, id.Index())
			require.Equal(t, test.checkpoint, id.Checkpoint())
			require.Equal(t, test.str, id.String())
		})
	}
}

func TestNodeID_IsEmpty(t *testing.T) {
	require.True(t, NodeID{}.IsEmpty())
	require.False(t, NewNodeID(true, 1, 1).IsEmpty())
	require.False(t, NewEmptyTreeNodeID(5).IsEmpty())
}

func TestNodeID_EmptyTree(t *testing.T) {
	// NewEmptyTreeNodeID creates a sentinel distinct from the zero value
	et := NewEmptyTreeNodeID(5)
	require.True(t, et.IsEmptyTree())
	require.False(t, et.IsEmpty())
	require.Equal(t, uint32(5), et.Checkpoint())

	// Zero value is not an empty tree
	require.False(t, NodeID{}.IsEmptyTree())

	// Regular nodes are not empty trees
	require.False(t, NewNodeID(true, 1, 1).IsEmptyTree())
	require.False(t, NewNodeID(false, 1, 1).IsEmptyTree())

	// checkpoint 0 panics
	require.Panics(t, func() { NewEmptyTreeNodeID(0) })

	// String round-trip via MarshalText/UnmarshalText
	require.Equal(t, "empty:5", et.String())
	txt, err := et.MarshalText()
	require.NoError(t, err)
	var parsed NodeID
	require.NoError(t, parsed.UnmarshalText(txt))
	require.True(t, et.Equal(parsed))
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
			name:  "different checkpoint",
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
