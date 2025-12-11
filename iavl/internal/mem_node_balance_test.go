package internal

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func newTestBranchNode(left, right *MemNode) *MemNode {
	return &MemNode{
		left:  NewNodePointer(left),
		right: NewNodePointer(right),
	}
}

func TestCalcBalance(t *testing.T) {
	tests := []struct {
		name    string
		node    *MemNode
		balance int
	}{
		{
			name:    "balanced",
			node:    newTestBranchNode(&MemNode{height: 2}, &MemNode{height: 2}),
			balance: 0,
		},
		{
			name:    "left heavy",
			node:    newTestBranchNode(&MemNode{height: 3}, &MemNode{height: 1}),
			balance: 2,
		},
		{
			name:    "right heavy",
			node:    newTestBranchNode(&MemNode{height: 1}, &MemNode{height: 4}),
			balance: -3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			balance, err := calcBalance(tt.node)
			require.NoError(t, err)
			require.Equal(t, tt.balance, balance)
		})
	}
}

func TestCalcBalance_UpdateHeightSize(t *testing.T) {
	tests := []struct {
		name   string
		node   *MemNode
		height uint8
		size   int64
	}{
		{
			name:   "two leaves",
			node:   newTestBranchNode(&MemNode{height: 0, size: 1}, &MemNode{height: 0, size: 1}),
			height: 1,
			size:   2,
		},
		{
			name:   "left child h 2 s 3, right child h 1 s 2",
			node:   newTestBranchNode(&MemNode{height: 2, size: 3}, &MemNode{height: 1, size: 2}),
			height: 3,
			size:   5,
		},
		{
			name:   "left child h 1 s 2, right child h 3 s 4",
			node:   newTestBranchNode(&MemNode{height: 1, size: 2}, &MemNode{height: 3, size: 4}),
			height: 4,
			size:   6,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.node.updateHeightSize()
			require.NoError(t, err)
			require.Equal(t, tt.height, tt.node.height)
			require.Equal(t, tt.size, tt.node.size)
		})
	}
}

func TestRotateLeft(t *testing.T) {
	// Construct a simple tree (all version 1):
	//	node (mutable)
	//	   /  \
	//	  A  right
	//	     /  \
	//	    B    C
	A := newLeafNode([]byte("A"), []byte("valueA"), 1)
	B := newLeafNode([]byte("B"), []byte("valueB"), 1)
	C := newLeafNode([]byte("C"), []byte("valueC"), 1)
	left := newTestBranchNode(A, B)
	node := newTestBranchNode(left, C)

	// After rotation, the tree should look like:
	//	copy of right (new version)
	//	   /                      \
	//	  node (now immutable)     C
	//	     /  \
	//	    A    B
	// with orphans: right
	ctx := &mutationContext{version: 2}
	newRoot, err := node.rotateRight(ctx)
	require.NoError(t, err)

}
