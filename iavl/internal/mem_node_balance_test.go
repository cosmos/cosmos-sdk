package internal

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func newTestLeafNode(key string, version uint32) *MemNode {
	node := newLeafNode([]byte(key), []byte("value_"+key), version)
	assignTestID(node)
	return node
}

func newTestBranchNode(left, right *MemNode, key string, version uint32) *MemNode {
	node := &MemNode{
		key:     []byte(key), // branch key = smallest key in right subtree (caller must set correctly)
		left:    NewNodePointer(left),
		right:   NewNodePointer(right),
		version: version,
	}
	_ = node.updateHeightSize() // ignore error - test nodes always have valid children
	assignTestID(node)
	return node
}

// testIDCounter is used to assign unique node IDs in tests.
var testIDCounter uint32

// assignTestID assigns a unique test ID to a node based on its version and type.
// this is not a real node ID, but is extremely simple for testing purposes.
func assignTestID(node *MemNode) {
	testIDCounter++
	node.nodeId = NewNodeID(node.IsLeaf(), node.version, testIDCounter)
}

func TestCalcBalance(t *testing.T) {
	tests := []struct {
		name    string
		node    *MemNode
		balance int
	}{
		{
			name:    "balanced",
			node:    newTestBranchNode(&MemNode{height: 2}, &MemNode{height: 2}, "", 1),
			balance: 0,
		},
		{
			name:    "left heavy",
			node:    newTestBranchNode(&MemNode{height: 3}, &MemNode{height: 1}, "", 1),
			balance: 2,
		},
		{
			name:    "right heavy",
			node:    newTestBranchNode(&MemNode{height: 1}, &MemNode{height: 4}, "", 1),
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

func TestUpdateHeightSize(t *testing.T) {
	// Construct branch nodes manually (without newTestBranchNode which already calls updateHeightSize)
	tests := []struct {
		name   string
		left   *MemNode
		right  *MemNode
		height uint8
		size   int64
	}{
		{
			name:   "two leaves",
			left:   newTestLeafNode("A", 1),
			right:  newTestLeafNode("B", 1),
			height: 1,
			size:   2,
		},
		{
			name:   "left subtree taller",
			left:   newTestBranchNode(newTestLeafNode("A", 1), newTestLeafNode("B", 1), "B", 1),
			right:  newTestLeafNode("C", 1),
			height: 2,
			size:   3,
		},
		{
			name:   "right subtree taller",
			left:   newTestLeafNode("A", 1),
			right:  newTestBranchNode(newTestLeafNode("B", 1), newTestLeafNode("C", 1), "C", 1),
			height: 2,
			size:   3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &MemNode{
				left:  NewNodePointer(tt.left),
				right: NewNodePointer(tt.right),
			}
			err := node.updateHeightSize()
			require.NoError(t, err)
			require.Equal(t, tt.height, node.height)
			require.Equal(t, tt.size, node.size)
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
