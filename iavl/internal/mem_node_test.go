package internal

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMemNode_Getters(t *testing.T) {
	left := NewNodePointer(&MemNode{})
	right := NewNodePointer(&MemNode{})
	nodeId := NewNodeID(true, 5, 10)

	node := &MemNode{
		height:    3,
		version:   7,
		size:      42,
		key:       []byte("testkey"),
		value:     []byte("testvalue"),
		hash:      []byte("testhash"),
		left:      left,
		right:     right,
		nodeId:    nodeId,
		keyOffset: 100,
	}

	require.Equal(t, uint8(3), node.Height())
	require.Equal(t, uint32(7), node.Version())
	require.Equal(t, int64(42), node.Size())
	require.Equal(t, left, node.Left())
	require.Equal(t, right, node.Right())
	require.Equal(t, []byte("testhash"), node.Hash())
	require.Equal(t, nodeId, node.ID())

	key, err := node.Key()
	require.NoError(t, err)
	require.Equal(t, []byte("testkey"), key)

	value, err := node.Value()
	require.NoError(t, err)
	require.Equal(t, []byte("testvalue"), value)
}

func TestMemNode_IsLeaf(t *testing.T) {
	tests := []struct {
		name   string
		height uint8
		want   bool
	}{
		{name: "leaf", height: 0, want: true},
		{name: "branch height 1", height: 1, want: false},
		{name: "branch height 5", height: 5, want: false},
		{name: "branch max height", height: 255, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &MemNode{height: tt.height}
			require.Equal(t, tt.want, node.IsLeaf())
		})
	}
}

func TestMemNode_String(t *testing.T) {
	tests := []struct {
		name string
		node *MemNode
		want string
	}{
		{
			name: "leaf node",
			node: &MemNode{
				height:  0,
				version: 1,
				size:    1,
				key:     []byte{0xab, 0xcd},
				value:   []byte{0x12, 0x34},
			},
			want: "MemNode{key:abcd, version:1, size:1, value:1234}",
		},
		{
			name: "branch node",
			node: &MemNode{
				height:  2,
				version: 5,
				size:    10,
				key:     []byte{0xff},
				left:    &NodePointer{id: NewNodeID(true, 1, 1)},
				right:   &NodePointer{id: NewNodeID(true, 1, 2)},
			},
			want: "MemNode{key:ff, version:5, size:10, height:2, left:NodePointer{id: NodeID{leaf:true, version:1, index:1}, fileIdx: 0}, right:NodePointer{id: NodeID{leaf:true, version:1, index:2}, fileIdx: 0}}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.node.String())
		})
	}
}

func TestMemNode_MutateBranch(t *testing.T) {
	original := &MemNode{
		height:  2,
		version: 5,
		size:    10,
		key:     []byte("key"),
		hash:    []byte("oldhash"),
		left:    NewNodePointer(&MemNode{}),
		right:   NewNodePointer(&MemNode{}),
	}

	mutated, err := original.MutateBranch(12)
	require.NoError(t, err)

	// Version updated, hash cleared
	require.Equal(t, uint32(12), mutated.Version())
	require.Nil(t, mutated.Hash())

	// Other fields preserved
	require.Equal(t, original.Height(), mutated.Height())
	require.Equal(t, original.Size(), mutated.Size())
	key, _ := mutated.Key()
	require.Equal(t, []byte("key"), key)
	require.Equal(t, original.Left(), mutated.Left())
	require.Equal(t, original.Right(), mutated.Right())

	// Is a copy, not same pointer
	require.NotSame(t, original, mutated)

	// Original unchanged
	require.Equal(t, uint32(5), original.Version())
	require.Equal(t, []byte("oldhash"), original.Hash())
}
