package internal

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func newTestLeafNode(version uint32, index uint32, key string) *MemNode {
	node := newLeafNode([]byte(key), []byte("value_"+key), version)
	node.nodeId = NewNodeID(true, version, index)
	return node
}

func newTestBranchNode(version, index uint32, left, right *MemNode) *MemNode {
	var getSmallestKey func(n *MemNode) []byte
	getSmallestKey = func(n *MemNode) []byte {
		if n.IsLeaf() {
			return n.key
		} else {
			return getSmallestKey(n.left.mem.Load())
		}
	}

	node := &MemNode{
		key:     getSmallestKey(right), // branch key = smallest key in right subtree
		left:    NewNodePointer(left),
		right:   NewNodePointer(right),
		version: version,
	}
	_ = node.updateHeightSize() // ignore error - test nodes always have valid children
	node.nodeId = NewNodeID(false, version, index)
	return node
}

func TestCalcBalance(t *testing.T) {
	tests := []struct {
		name    string
		node    *MemNode
		balance int
	}{
		{
			name: "balanced",
			node: newTestBranchNode(1, 1,
				newTestLeafNode(1, 2, "A"),
				newTestLeafNode(1, 3, "B"),
			),
			balance: 0,
		},
		{
			name: "left heavy",
			node: newTestBranchNode(1, 1,
				newTestBranchNode(1, 2,
					newTestLeafNode(1, 3, "A"),
					newTestLeafNode(1, 4, "B"),
				),
				newTestLeafNode(1, 5, "C"),
			),
			balance: 1,
		},
		{
			name: "right heavy",
			node: newTestBranchNode(1, 1,
				newTestLeafNode(1, 2, "A"),
				newTestBranchNode(1, 3,
					newTestLeafNode(1, 4, "B"),
					newTestBranchNode(1, 5,
						newTestLeafNode(1, 6, "C"),
						newTestLeafNode(1, 7, "D"),
					),
				)),
			balance: -2,
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
			left:   newTestLeafNode(1, 1, "A"),
			right:  newTestLeafNode(1, 2, "B"),
			height: 1,
			size:   2,
		},
		{
			name: "left subtree taller",
			left: newTestBranchNode(1, 1,
				newTestLeafNode(1, 2, "A"),
				newTestLeafNode(1, 3, "B"),
			),
			right:  newTestLeafNode(1, 4, "C"),
			height: 2,
			size:   3,
		},
		{
			name: "right subtree taller",
			left: newTestLeafNode(1, 1, "A"),
			right: newTestBranchNode(1, 2,
				newTestLeafNode(1, 3, "B"),
				newTestLeafNode(1, 4, "C"),
			),
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

	Y := newTestBranchNode(2, 1,
		newTestLeafNode(1, 2, "X"),
		newTestBranchNode(1, 3,
			newTestLeafNode(1, 4, "Y"),
			newTestLeafNode(1, 5, "Z"),
		),
	)

	require.Equal(t, "(Y.2.1 [X.1.2] (Z.1.3 [Y.1.4] [Z.1.5]))", printTreeStructure(t, Y))

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

	require.Equal(t, "(Z.2.2 (Y.2.1 [X.1.2] [Y.1.4]) [Z.1.5])", printTreeStructure(t, newRoot))
	require.Equal(t, []NodeID{NewNodeID(false, 1, 3)}, ctx.orphans)
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
	Y := newTestBranchNode(2, 1,
		newTestBranchNode(1, 2,
			newTestLeafNode(1, 3, "W"),
			newTestLeafNode(1, 4, "X")),
		newTestLeafNode(1, 5, "Y"))

	require.Equal(t, "(Y.2.1 (X.1.2 [W.1.3] [X.1.4]) [Y.1.5])", printTreeStructure(t, Y))

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

	require.Equal(t, "(X.2.2 [W.1.3] (Y.2.1 [X.1.4] [Y.1.5]))", printTreeStructure(t, newRoot))
	require.Equal(t, []NodeID{NewNodeID(false, 1, 2)}, ctx.orphans)
	require.NoError(t, verifyAVLInvariants(newRoot))
}

func TestNodeRebalance(t *testing.T) {
	tests := []struct {
		name           string
		root           *MemNode
		beforeRotation string
		afterRotation  string
		orphans        []NodeID
	}{
		{
			name: "balanced tree - no rotation",
			root: newTestBranchNode(2, 1,
				newTestLeafNode(1, 2, "X"),
				newTestBranchNode(1, 3,
					newTestLeafNode(1, 4, "Y"),
					newTestLeafNode(1, 5, "Z"),
				),
			),
			//         Y.2.1
			//       /      \
			//   [X.1.2]     Z.1.3
			//              /     \
			//          [Y.1.4] [Z.1.5]
			beforeRotation: "(Y.2.1 [X.1.2] (Z.1.3 [Y.1.4] [Z.1.5]))",
			afterRotation:  "(Y.2.1 [X.1.2] (Z.1.3 [Y.1.4] [Z.1.5]))", // unchanged
		},
		{
			name: "left-left case",
			root: newTestBranchNode(2, 1,
				newTestBranchNode(1, 2,
					newTestBranchNode(1, 3,
						newTestLeafNode(1, 4, "W"),
						newTestLeafNode(1, 5, "X"),
					),
					newTestLeafNode(1, 6, "Y"),
				),
				newTestLeafNode(1, 7, "Z"),
			),
			//        Z.2.1 (mutable)
			//       /           \
			//      Y.1.2      [Z.1.7]
			//     /     \
			//    X.1.3  [Y.1.6]
			//    /    \
			// [W.1.4] [X.1.5]
			beforeRotation: "(Z.2.1 (Y.1.2 (X.1.3 [W.1.4] [X.1.5]) [Y.1.6]) [Z.1.7])",
			//         Y.2.2 (copy of Y.1.2)
			//        /                \
			//      X.1.3         Z.2.1 (immutable)
			//     /     \           /     \
			//   [W.1.4] [X.1.5]  [Y.1.6] [Z.1.7]
			//
			// orphans: Y.1.2 (the original branch, now replaced by Y.2.2)
			afterRotation: "(Y.2.2 (X.1.3 [W.1.4] [X.1.5]) (Z.2.1 [Y.1.6] [Z.1.7]))",
			orphans:       []NodeID{NewNodeID(false, 1, 2)},
		},
		{
			name: "left-right case",
			root: newTestBranchNode(2, 1,
				newTestBranchNode(1, 2,
					newTestLeafNode(1, 3, "W"),
					newTestBranchNode(1, 4,
						newTestLeafNode(1, 5, "X"),
						newTestLeafNode(1, 6, "Y"),
					),
				),
				newTestLeafNode(1, 7, "Z"),
			),
			//        Z.2.1 (mutable)
			//       /           \
			//     X.1.2       [Z.1.7]
			//    /      \
			// [W.1.3]  Y.1.4
			//          /    \
			//      [X.1.5] [Y.1.6]
			beforeRotation: "(Z.2.1 (X.1.2 [W.1.3] (Y.1.4 [X.1.5] [Y.1.6])) [Z.1.7])",
			//              Y.2.2 (copy of Y.1.4)
			//             /                \
			//  X.2.3 (copy of X.1.2)   Z.2.1 (immutable)
			//    /      \                 /     \
			// [W.1.3] [X.1.5]        [Y.1.6] [Z.1.7]
			//
			// orphans: X.1.2 (replaced by X.2.3), Y.1.4 (replaced by Y.2.2)
			afterRotation: "(Y.2.2 (X.2.3 [W.1.3] [X.1.5]) (Z.2.1 [Y.1.6] [Z.1.7]))",
			orphans:       []NodeID{NewNodeID(false, 1, 2), NewNodeID(false, 1, 4)},
		},
		{
			name: "right-right case",
			root: newTestBranchNode(2, 1,
				newTestLeafNode(1, 2, "W"),
				newTestBranchNode(1, 3,
					newTestLeafNode(1, 4, "X"),
					newTestBranchNode(1, 5,
						newTestLeafNode(1, 6, "Y"),
						newTestLeafNode(1, 7, "Z"),
					),
				),
			),
			//     X.2.1 (mutable)
			//    /           \
			// [W.1.2]       Y.1.3
			//              /     \
			//          [X.1.4]  Z.1.5
			//                   /    \
			//               [Y.1.6] [Z.1.7]
			beforeRotation: "(X.2.1 [W.1.2] (Y.1.3 [X.1.4] (Z.1.5 [Y.1.6] [Z.1.7])))",
			//         Y.2.2 (copy of Y.1.3)
			//        /                \
			//     X.2.1 (immutable)  Z.1.5
			//    /      \           /     \
			// [W.1.2] [X.1.4]  [Y.1.6] [Z.1.7]
			//
			// orphans: Y.1.3 (replaced by Y.2.2)
			afterRotation: "(Y.2.2 (X.2.1 [W.1.2] [X.1.4]) (Z.1.5 [Y.1.6] [Z.1.7]))",
			orphans:       []NodeID{NewNodeID(false, 1, 3)},
		},
		{
			name: "right-left case",
			root: newTestBranchNode(2, 1,
				newTestLeafNode(1, 2, "W"),
				newTestBranchNode(1, 3,
					newTestBranchNode(1, 4,
						newTestLeafNode(1, 5, "X"),
						newTestLeafNode(1, 6, "Y"),
					),
					newTestLeafNode(1, 7, "Z"),
				),
			),
			//     X.2.1 (mutable)
			//    /           \
			// [W.1.2]       Z.1.3
			//              /     \
			//           Y.1.4  [Z.1.7]
			//           /    \
			//       [X.1.5] [Y.1.6]
			beforeRotation: "(X.2.1 [W.1.2] (Z.1.3 (Y.1.4 [X.1.5] [Y.1.6]) [Z.1.7]))",
			//              Y.2.2 (copy of Y.1.4)
			//             /                \
			//  X.2.1 (immutable)    Z.2.3 (copy of Z.1.3)
			//    /      \                 /     \
			// [W.1.2] [X.1.5]        [Y.1.6] [Z.1.7]
			//
			// orphans: Z.1.3 (replaced by Z.2.3), Y.1.4 (replaced by Y.2.2)
			afterRotation: "(Y.2.2 (X.2.1 [W.1.2] [X.1.5]) (Z.2.3 [Y.1.6] [Z.1.7]))",
			orphans:       []NodeID{NewNodeID(false, 1, 3), NewNodeID(false, 1, 4)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.beforeRotation, printTreeStructure(t, tt.root), "tree structure before reBalance")
			ctx := &mutationContext{version: 2}
			newRoot, err := tt.root.reBalance(ctx)
			require.NoError(t, err, "reBalance error")
			require.Equal(t, tt.afterRotation, printTreeStructure(t, newRoot), "tree structure after reBalance")
			require.NoError(t, verifyAVLInvariants(newRoot))
			// check orphans
			require.Equal(t, tt.orphans, ctx.orphans, "orphans after reBalance")
		})
	}
}

// printTreeStructure returns a string representation of the tree structure.
// Leaves are formatted as [key.version], branches as (key.version left right).
// Example: (Y.1 [X.1] (Z.1 [Y.1] [Z.1]))
func printTreeStructure(t *testing.T, node *MemNode) string {
	seen := map[NodeID]bool{}
	// collect all existing IDs to avoid temporary ID collisions
	var collectIds func(node *MemNode)
	collectIds = func(node *MemNode) {
		if !node.nodeId.IsEmpty() {
			seen[node.nodeId] = true
		}
		if !node.IsLeaf() {
			collectIds(node.left.mem.Load())
			collectIds(node.right.mem.Load())
		}
	}
	collectIds(node)

	var doPrintNode func(node *MemNode) string
	doPrintNode = func(node *MemNode) string {
		id := node.nodeId
		// assign temporary ID if missing
		if id.IsEmpty() {
			n := uint32(1)
			for {
				id = NewNodeID(node.IsLeaf(), node.version, n)
				if !seen[id] {
					seen[id] = true
					break
				}
				n++
			}
		}

		key := node.key
		if node.IsLeaf() {
			return fmt.Sprintf("[%s.%d.%d]", key, id.Version(), id.Index())
		}
		return fmt.Sprintf("(%s.%d.%d %s %s)",
			key, id.Version(), id.Index(),
			doPrintNode(node.left.mem.Load()),
			doPrintNode(node.right.mem.Load()),
		)
	}
	return doPrintNode(node)
}
