package internal

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMemNode_Getters(t *testing.T) {
	left := NewNodePointer(&MemNode{})
	right := NewNodePointer(&MemNode{})
	nodeId := NewNodeID(true, 5, 10)

	testKey := []byte("testkey")
	testValue := []byte("testvalue")
	testHash := []byte("testhash")
	node := &MemNode{
		height:         3,
		version:        7,
		size:           42,
		key:            testKey,
		value:          testValue,
		hash:           testHash,
		left:           left,
		right:          right,
		nodeId:         nodeId,
		walKeyOffset:   NewKVOffset(100, false),
		walValueOffset: NewKVOffset(200, false),
	}

	require.Equal(t, uint8(3), node.Height())
	require.Equal(t, uint32(7), node.Version())
	require.Equal(t, int64(42), node.Size())
	require.Equal(t, left, node.Left())
	require.Equal(t, right, node.Right())
	require.Equal(t, testHash, node.Hash().UnsafeBytes())
	require.Equal(t, nodeId, node.ID())

	key, err := node.Key()
	require.NoError(t, err)
	require.Equal(t, testKey, key.UnsafeBytes())

	value, err := node.Value()
	require.NoError(t, err)
	require.Equal(t, testValue, value.UnsafeBytes())
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
			want: "MemNode{key:abcd, layer:1, size:1, value:1234}",
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
			want: "MemNode{key:ff, layer:5, size:10, height:2, left:NodePointer{id: NodeID{leaf:true, layer:1, index:1}, fileIdx: 0}, right:NodePointer{id: NodeID{leaf:true, layer:1, index:2}, fileIdx: 0}}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.node.String())
		})
	}
}

func TestMemNode_MutateBranch(t *testing.T) {
	key := []byte("key")
	origHash := []byte("origHash")
	origID := NewNodeID(false, 5, 42)
	original := &MemNode{
		height:  2,
		version: 5,
		size:    10,
		key:     key,
		hash:    origHash,
		nodeId:  origID,
		left:    NewNodePointer(&MemNode{}),
		right:   NewNodePointer(&MemNode{}),
	}

	mutated, err := original.MutateBranch(12)
	require.NoError(t, err)

	// Version updated, hash and ID cleared
	require.Equal(t, uint32(12), mutated.Version())
	require.Nil(t, mutated.Hash().UnsafeBytes())
	require.True(t, mutated.ID().IsEmpty())

	// Other fields preserved
	require.Equal(t, original.Height(), mutated.Height())
	require.Equal(t, original.Size(), mutated.Size())
	key2, _ := mutated.Key()
	require.Equal(t, key, key2.UnsafeBytes())
	require.Equal(t, original.Left(), mutated.Left())
	require.Equal(t, original.Right(), mutated.Right())

	// Is a copy, not same pointer
	require.NotSame(t, original, mutated)

	// Original unchanged
	require.Equal(t, uint32(5), original.Version())
	require.Equal(t, origHash, original.Hash().UnsafeBytes())
	require.Equal(t, origID, original.ID())
}

func TestMemNode_Get_Leaf(t *testing.T) {
	// When Get is called on a leaf node:
	// - If key matches: returns (value, 0, nil)
	// - If key not found: returns (nil, index, nil) where index is the insertion point
	//   - key < nodeKey: index=0 (would insert before this leaf)
	//   - key > nodeKey: index=1 (would insert after this leaf)
	tests := []struct {
		name      string
		nodeKey   string
		nodeValue string
		searchKey string
		wantValue []byte
		wantIndex int64
	}{
		{
			name:      "exact match",
			nodeKey:   "b",
			nodeValue: "val_b",
			searchKey: "b",
			wantValue: []byte("val_b"),
			wantIndex: 0,
		},
		{
			name:      "search key less than node key",
			nodeKey:   "b",
			nodeValue: "val_b",
			searchKey: "a",
			wantValue: nil,
			wantIndex: 0, // "a" would be inserted before "b"
		},
		{
			name:      "search key greater than node key",
			nodeKey:   "b",
			nodeValue: "val_b",
			searchKey: "c",
			wantValue: nil,
			wantIndex: 1, // "c" would be inserted after "b"
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &MemNode{
				height: 0,
				size:   1,
				key:    []byte(tt.nodeKey),
				value:  []byte(tt.nodeValue),
			}
			val, idx, err := node.Get([]byte(tt.searchKey))
			require.NoError(t, err)
			require.Equal(t, tt.wantValue, val.UnsafeBytes())
			require.Equal(t, tt.wantIndex, idx)
		})
	}
}

func TestMemNode_Get_Branch(t *testing.T) {
	// Hand-construct a simple tree:
	//
	//       [b]          <- branch, key="b", size=2
	//      /   \
	//    [a]   [b]       <- leaves (index 0, index 1)
	//
	// In IAVL, branch key = smallest key in right subtree
	//
	// Index is the 0-based position in sorted leaf order:
	// - "a" is at index 0, "b" is at index 1
	// - Keys not found return the insertion point

	leftLeaf := &MemNode{
		height: 0,
		size:   1,
		key:    []byte("a"),
		value:  []byte("val_a"),
	}
	rightLeaf := &MemNode{
		height: 0,
		size:   1,
		key:    []byte("b"),
		value:  []byte("val_b"),
	}
	root := &MemNode{
		height: 1,
		size:   2,
		key:    []byte("b"), // smallest key in right subtree
		left:   NewNodePointer(leftLeaf),
		right:  NewNodePointer(rightLeaf),
	}

	tests := []struct {
		name      string
		searchKey string
		wantValue []byte
		wantIndex int64
	}{
		{
			name:      "find in left subtree",
			searchKey: "a",
			wantValue: []byte("val_a"),
			wantIndex: 0,
		},
		{
			name:      "find in right subtree",
			searchKey: "b",
			wantValue: []byte("val_b"),
			wantIndex: 1,
		},
		{
			name:      "key not found - less than all",
			searchKey: "0",
			wantValue: nil,
			wantIndex: 0, // "0" would be inserted at position 0
		},
		{
			name:      "key not found - greater than all",
			searchKey: "z",
			wantValue: nil,
			wantIndex: 2, // "z" would be inserted at position 2 (after both leaves)
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, idx, err := root.Get([]byte(tt.searchKey))
			require.NoError(t, err)
			require.Equal(t, tt.wantValue, val.UnsafeBytes())
			require.Equal(t, tt.wantIndex, idx)
		})
	}
}

func TestMemNode_Get_DeeperTree(t *testing.T) {
	// Hand-construct a 3-level tree:
	//
	//            [c]              <- root, size=4
	//          /     \
	//       [b]       [d]         <- branches, size=2 each
	//      /   \     /   \
	//    [a]   [b] [c]   [d]      <- leaves
	//
	// Sorted keys: a=0, b=1, c=2, d=3

	leafA := &MemNode{height: 0, size: 1, key: []byte("a"), value: []byte("val_a")}
	leafB := &MemNode{height: 0, size: 1, key: []byte("b"), value: []byte("val_b")}
	leafC := &MemNode{height: 0, size: 1, key: []byte("c"), value: []byte("val_c")}
	leafD := &MemNode{height: 0, size: 1, key: []byte("d"), value: []byte("val_d")}

	branchLeft := &MemNode{
		height: 1,
		size:   2,
		key:    []byte("b"),
		left:   NewNodePointer(leafA),
		right:  NewNodePointer(leafB),
	}
	branchRight := &MemNode{
		height: 1,
		size:   2,
		key:    []byte("d"),
		left:   NewNodePointer(leafC),
		right:  NewNodePointer(leafD),
	}
	root := &MemNode{
		height: 2,
		size:   4,
		key:    []byte("c"), // smallest key in right subtree
		left:   NewNodePointer(branchLeft),
		right:  NewNodePointer(branchRight),
	}

	tests := []struct {
		searchKey string
		wantValue []byte
		wantIndex int64
	}{
		{"a", []byte("val_a"), 0},
		{"b", []byte("val_b"), 1},
		{"c", []byte("val_c"), 2},
		{"d", []byte("val_d"), 3},
	}
	for _, tt := range tests {
		t.Run(tt.searchKey, func(t *testing.T) {
			val, idx, err := root.Get([]byte(tt.searchKey))
			require.NoError(t, err)
			require.Equal(t, tt.wantValue, val.UnsafeBytes())
			require.Equal(t, tt.wantIndex, idx)
		})
	}
}
