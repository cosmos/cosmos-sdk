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
	// Construct the tree from rotateLeft docs (all version 1):
	//
	//	X (mutable)
	//	  /  \
	//	[X]   Y
	//	     / \
	//	   [Y] [Z]
	//
	leafX := newTestLeafNode("X", 1)
	leafY := newTestLeafNode("Y", 1)
	leafZ := newTestLeafNode("Z", 1)

	// Y's key is Z (smallest key in right subtree, which is just [Z])
	Y := newTestBranchNode(leafY, leafZ, "Z", 1)

	// X's key is Y (smallest key in right subtree, which starts at [Y] in Y's left)
	X := newTestBranchNode(leafX, Y, "Y", 1)

	// After rotation, the tree should look like:
	//
	//	     copy of Y
	//	  /            \
	//	X (immutable)  [Z]
	//	   /  \
	//	 [X]  [Y]
	//
	// orphans: Y

	ctx := &mutationContext{version: 2}
	newRoot, err := X.rotateLeft(ctx)
	require.NoError(t, err)

	// Verify new root is copy of Y with new version
	// Y's key was Z (smallest in right subtree [Z])
	require.Equal(t, []byte("Z"), newRoot.key)
	require.Equal(t, uint32(2), newRoot.version)

	// Verify left child is X (now immutable, still version 1)
	leftChild, leftPin, err := newRoot.left.Resolve()
	require.NoError(t, err)
	defer leftPin.Unpin()
	require.Equal(t, []byte("Y"), leftChild.(*MemNode).key) // X's key was Y (first key of right subtree)
	require.Equal(t, uint32(1), leftChild.Version())

	// Verify right child is [Z]
	rightChild, rightPin, err := newRoot.right.Resolve()
	require.NoError(t, err)
	defer rightPin.Unpin()
	require.Equal(t, []byte("Z"), rightChild.(*MemNode).key)
	require.True(t, rightChild.IsLeaf())

	// Verify X's children are now [X] and [Y]
	xLeft, xLeftPin, err := leftChild.Left().Resolve()
	require.NoError(t, err)
	defer xLeftPin.Unpin()
	require.Equal(t, []byte("X"), xLeft.(*MemNode).key)
	require.True(t, xLeft.IsLeaf())

	xRight, xRightPin, err := leftChild.Right().Resolve()
	require.NoError(t, err)
	defer xRightPin.Unpin()
	require.Equal(t, []byte("Y"), xRight.(*MemNode).key)
	require.True(t, xRight.IsLeaf())

	// Verify orphans contains Y
	require.Len(t, ctx.orphans, 1)

	// Verify heights and sizes are correct
	require.Equal(t, uint8(2), newRoot.height)
	require.Equal(t, int64(3), newRoot.size)
	require.Equal(t, uint8(1), leftChild.Height())
	require.Equal(t, int64(2), leftChild.Size())

	// Verify all invariants
	require.NoError(t, verifyAVLInvariants(newRoot))
}

func TestRotateRight(t *testing.T) {
	// Construct the tree from rotateRight docs (all version 1):
	//
	//	  Y (mutable)
	//	    /  \
	//	   X   [Y]
	//	  / \
	//	[W] [X]
	//
	leafW := newTestLeafNode("W", 1)
	leafX := newTestLeafNode("X", 1)
	leafY := newTestLeafNode("Y", 1)

	// X's key is X (smallest key in right subtree, which is just [X])
	X := newTestBranchNode(leafW, leafX, "X", 1)

	// Y's key is Y (smallest key in right subtree, which is just [Y])
	Y := newTestBranchNode(X, leafY, "Y", 1)

	// After rotation, the tree should look like:
	//
	//	copy of X
	//	  /      \
	//	[W]    Y (immutable)
	//	         /  \
	//	       [X]  [Y]
	//
	// orphans: X

	ctx := &mutationContext{version: 2}
	newRoot, err := Y.rotateRight(ctx)
	require.NoError(t, err)

	// Verify new root is copy of X with new version
	// X's key was X (smallest in right subtree [X])
	require.Equal(t, []byte("X"), newRoot.key)
	require.Equal(t, uint32(2), newRoot.version)

	// Verify left child is [W]
	leftChild, leftPin, err := newRoot.left.Resolve()
	require.NoError(t, err)
	defer leftPin.Unpin()
	require.Equal(t, []byte("W"), leftChild.(*MemNode).key)
	require.True(t, leftChild.IsLeaf())

	// Verify right child is Y (now immutable, still version 1)
	rightChild, rightPin, err := newRoot.right.Resolve()
	require.NoError(t, err)
	defer rightPin.Unpin()
	require.Equal(t, []byte("Y"), rightChild.(*MemNode).key) // Y's key was Y (first key of right subtree)
	require.Equal(t, uint32(1), rightChild.Version())

	// Verify Y's children are now [X] and [Y]
	yLeft, yLeftPin, err := rightChild.Left().Resolve()
	require.NoError(t, err)
	defer yLeftPin.Unpin()
	require.Equal(t, []byte("X"), yLeft.(*MemNode).key)
	require.True(t, yLeft.IsLeaf())

	yRight, yRightPin, err := rightChild.Right().Resolve()
	require.NoError(t, err)
	defer yRightPin.Unpin()
	require.Equal(t, []byte("Y"), yRight.(*MemNode).key)
	require.True(t, yRight.IsLeaf())

	// Verify orphans contains X
	require.Len(t, ctx.orphans, 1)

	// Verify heights and sizes are correct
	require.Equal(t, uint8(2), newRoot.height)
	require.Equal(t, int64(3), newRoot.size)
	require.Equal(t, uint8(1), rightChild.Height())
	require.Equal(t, int64(2), rightChild.Size())

	// Verify all invariants
	require.NoError(t, verifyAVLInvariants(newRoot))
}
