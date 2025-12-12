package internal

import (
	"fmt"
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
	// Construct tree:
	//
	//	 Y (mutable)
	//	  /   \
	//	[X]    Z
	//	      / \
	//	    [Y] [Z]
	//
	leafX := newTestLeafNode("X", 1)
	leafY := newTestLeafNode("Y", 1)
	leafZ := newTestLeafNode("Z", 1)
	Z := newTestBranchNode(leafY, leafZ, "Z", 1)
	Y := newTestBranchNode(leafX, Z, "Y", 2) // a new mutable node in version 2 at the root

	require.Equal(t, "(Y.2 [X.1] (Z.1 [Y.1] [Z.1]))", printTreeStructure(t, Y))

	// After rotation:
	//
	//	        copy of Z
	//	    /              \
	//	   Y (immutable)    [Z]
	//	  / \
	//	[X] [Y]
	//
	// orphans: Z (the original branch, now replaced by newRoot)

	ctx := &mutationContext{version: 2}
	newRoot, err := Y.rotateLeft(ctx)
	require.NoError(t, err)

	require.Equal(t, "(Z.2 (Y.2 [X.1] [Y.1]) [Z.1])", printTreeStructure(t, newRoot))
	require.Equal(t, []NodeID{Z.ID()}, ctx.orphans)
	require.NoError(t, verifyAVLInvariants(newRoot))
}

func TestRotateRight(t *testing.T) {
	// Construct tree:
	//
	//	   Y (mutable)
	//	    /   \
	//	   X    [Y]
	//	  / \
	//	[W] [X]
	//
	leafW := newTestLeafNode("W", 1)
	leafX := newTestLeafNode("X", 1)
	leafY := newTestLeafNode("Y", 1)
	X := newTestBranchNode(leafW, leafX, "X", 1)
	Y := newTestBranchNode(X, leafY, "Y", 2) // a new mutable node in version 2 at the root

	require.Equal(t, "(Y.2 (X.1 [W.1] [X.1]) [Y.1])", printTreeStructure(t, Y))

	// After rotation:
	//
	//	  copy of X
	//	  /   \
	//	[W]    Y (immutable)
	//	      / \
	//	    [X] [Y]
	//
	// orphans: X (the original branch, now replaced by newRoot)

	ctx := &mutationContext{version: 2}
	newRoot, err := Y.rotateRight(ctx)
	require.NoError(t, err)

	require.Equal(t, "(X.2 [W.1] (Y.2 [X.1] [Y.1]))", printTreeStructure(t, newRoot))
	require.Equal(t, []NodeID{X.ID()}, ctx.orphans)
	require.NoError(t, verifyAVLInvariants(newRoot))
}

func TestNodeRebalance(t *testing.T) {

}

// printTreeStructure returns a string representation of the tree structure.
// Leaves are formatted as [key.version], branches as (key.version left right).
// Example: (Y.1 [X.1] (Z.1 [Y.1] [Z.1]))
func printTreeStructure(t *testing.T, node Node) string {
	key, err := node.Key()
	require.NoError(t, err)
	if node.IsLeaf() {
		return fmt.Sprintf("[%s.%d]", key.UnsafeBytes(), node.Version())
	}
	leftNode, leftPin, err := node.Left().Resolve()
	require.NoError(t, err)
	defer leftPin.Unpin()
	rightNode, rightPin, err := node.Right().Resolve()
	require.NoError(t, err)
	defer rightPin.Unpin()
	return fmt.Sprintf("(%s.%d %s %s)", key.UnsafeBytes(), node.Version(), printTreeStructure(t, leftNode), printTreeStructure(t, rightNode))
}
