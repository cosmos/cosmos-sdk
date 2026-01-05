package internal

// reBalance implements AVL tree rebalancing.
//
// AVL trees maintain a balance invariant: for every node, the height of its
// left and right subtrees may differ by at most 1. This ensures O(log n) operations.
//
// The balance factor of a node is: height(left) - height(right)
//   - balance > 1:  left-heavy, needs right rotation(s)
//   - balance < -1: right-heavy, needs left rotation(s)
//   - -1 <= balance <= 1: balanced, no rotation needed
//
// There are four imbalance cases, each resolved by one or two rotations:
//
//	Left-Left (LL):   Single right rotation
//	Left-Right (LR):  Left rotation on left child, then right rotation
//	Right-Right (RR): Single left rotation
//	Right-Left (RL):  Right rotation on right child, then left rotation

// reBalance checks the balance factor of a node and performs rotations if needed
// to restore the AVL balance invariant.
//
// Also, an IAVL tree is an AVL+ tree, so all values are contained in leaf nodes only.
// This means that a branch node's key is always the smallest key in its right subtree.
//
// Because an IAVL tree is immutable, we must copy existing branch nodes before modifying them.
//
// In the diagrams below [] indicates leaf nodes with concrete keys, and uppercase letters
// (P, Q, R, S) are abstract names for branch nodes (their actual keys follow IAVL rules).
//
// When looking at each diagram, you should verify the following invariants are maintained after rotation:
// 1. Leaf keys remain in sorted order (left to right): [W] < [X] < [Y] < [Z]
// 2. Each branch key equals its right child's key (the smallest key in its right subtree)
// 3. The tree is balanced (no node has children differing in height by more than 1)
// 4. Mutated branch nodes are copied and originals are marked as orphans
//
// Additional invariants maintained by the code but not shown in diagrams:
// 5. Height = max(left.height, right.height) + 1 for all branch nodes
// 6. Size = left.size + right.size for all branch nodes (total leaf count)
//
// The four cases handled:
//
//	Left-Left - left heavy and left child is also left heavy (balance > 1, leftBalance >= 0)
//	Needs single right rotation on root (node):
//	        P (mutable)              copy of Q
//	       / \                        /      \
//	      Q   [Z]                    R    P (immutable)
//	     / \        =>              / \      / \
//	    R   [Y]                   [W] [X]  [Y] [Z]
//	   / \
//	 [W] [X]
//
//	orphans: Q
//
//	Left-Right - left heavy but left child is right heavy (balance > 1, leftBalance < 0)
//	Needs left rotation on left child (Q), then right rotation on root (node):
//	      P (mutable)         P                   copy of R
//	       / \                 / \                  /      \
//	      Q   [Z]       copy of R  [Z]      copy of Q     P (immutable)
//	     / \        =>        / \        =>        / \      / \
//	   [W]  R          copy of Q  [Y]            [W] [X]  [Y] [Z]
//	       / \              / \
//	     [X] [Y]          [W] [X]
//
//	orphans: Q, R
//
//	Right-Right - right heavy and right child is also right heavy (balance < -1, rightBalance <= 0)
//	Needs single left rotation on root (node):
//	       P (mutable)                copy of Q
//	       / \                        /      \
//	     [W]  Q          P (immutable)        R
//	         / \       =>        / \        / \
//	       [X]  R              [W] [X]   [Y] [Z]
//	           / \
//	         [Y] [Z]
//
//	orphans: Q
//
//	Right-Left - right heavy but right child is left heavy (balance < -1, rightBalance > 0)
//	Needs right rotation on right child (Q), then left rotation on root (node):
//	      P (mutable)         P                    copy of R
//	       / \                 / \                  /      \
//	     [W]  Q              [W]  copy of R   P (immutable)  copy of Q
//	         / \      =>         / \       =>       / \        / \
//	        R  [Z]            [X]  copy of Q      [W] [X]   [Y] [Z]
//	       / \                    / \
//	     [X] [Y]                [Y] [Z]
//
//	orphans: Q, R
//
// IMPORTANT: This method must only be called on newly created or copied nodes.
// Code reviewers should check that the node is new or copied by doing a find usages check on this method.
func (node *MemNode) reBalance(ctx *mutationContext) (*MemNode, error) {
	balance, err := calcBalance(node)
	if err != nil {
		return nil, err
	}
	switch {
	// left heavy
	case balance > 1:
		left, leftPin, err := node.left.Resolve()
		defer leftPin.Unpin()
		if err != nil {
			return nil, err
		}

		leftBalance, err := calcBalance(left)
		if err != nil {
			return nil, err
		}

		// Left-Left (LL) case, needs single right rotation on root
		if leftBalance >= 0 {
			// left left
			return node.rotateRight(ctx)
		}

		// Left-Right (LR) case, needs left rotation on left-child then right rotation on root
		newLeft, err := ctx.mutateBranch(left)
		if err != nil {
			return nil, err
		}
		newLeft, err = newLeft.rotateLeft(ctx)
		if err != nil {
			return nil, err
		}
		node.left = NewNodePointer(newLeft)
		return node.rotateRight(ctx)
	case balance < -1:
		right, rightPin, err := node.right.Resolve()
		defer rightPin.Unpin()
		if err != nil {
			return nil, err
		}

		rightBalance, err := calcBalance(right)
		if err != nil {
			return nil, err
		}

		// Right-Right (RR) case, needs single left rotation on root
		if rightBalance <= 0 {
			// right right
			return node.rotateLeft(ctx)
		}

		// Right-Left (RL) case, needs right rotation on right-child then left rotation on root
		newRight, err := ctx.mutateBranch(right)
		if err != nil {
			return nil, err
		}
		newRight, err = newRight.rotateRight(ctx)
		if err != nil {
			return nil, err
		}
		node.right = NewNodePointer(newRight)
		return node.rotateLeft(ctx)
	default:
		// nothing changed
		return node, err
	}
}

// calcBalance computes the balance factor of a branch node: height(left) - height(right).
// This method MUST NOT be called on leaf nodes, or it will panic.
//
// Return values:
//   - positive: left subtree is taller (left-heavy)
//   - negative: right subtree is taller (right-heavy)
//   - zero: subtrees have equal height (perfectly balanced)
func calcBalance(node Node) (int, error) {
	leftNode, leftPin, err := node.Left().Resolve()
	defer leftPin.Unpin()
	if err != nil {
		return 0, err
	}

	rightNode, rightPin, err := node.Right().Resolve()
	defer rightPin.Unpin()
	if err != nil {
		return 0, err
	}

	return int(leftNode.Height()) - int(rightNode.Height()), nil
}

// updateHeightSize recalculates the height and size of a branch node from its children.
// This method MUST NOT be called on leaf nodes, or it will panic.
//
// Height is set to max(left.height, right.height) + 1 (for the current node).
// Size is set to left.size + right.size (total leaf node count in subtree).
//
// This must be called after any structural change to a node's children
// (insertion, deletion, rotation) to maintain correct metadata.
//
// IMPORTANT: This method must only be called on newly created or copied nodes.
// Code reviewers should check that the node is new or copied by doing a find usages check on this method.
func (node *MemNode) updateHeightSize() error {
	leftNode, leftPin, err := node.left.Resolve()
	defer leftPin.Unpin()
	if err != nil {
		return err
	}

	rightNode, rightPin, err := node.right.Resolve()
	defer rightPin.Unpin()
	if err != nil {
		return err
	}

	node.height = maxUint8(leftNode.Height(), rightNode.Height()) + 1
	node.size = leftNode.Size() + rightNode.Size()
	return nil
}

// maxUint8 returns the maximum of two uint8 values.
func maxUint8(a, b uint8) uint8 {
	if a > b {
		return a
	}
	return b
}

// rotateRight performs a right rotation around this node.
//
// A new node is created by copying the left child, which becomes
// the new root of this subtree. The original left child is orphaned.
// The current node (which must be new and mutable) becomes the right child of the new root.
// The original left child's right subtree becomes the current node's new left subtree.
//
//	  P (mutable)                copy of Q
//	    /  \                      /      \
//	   Q   [Y]      =>          [W]    P (immutable)     +    orphans: original Q
//	  / \                                /  \
//	[W] [X]                            [X]  [Y]
//
// After rotation, heights and sizes are recalculated bottom-up (Y first,
// then the new root) since Y is now a child of the new root.
//
// IMPORTANT: This method must only be called on newly created or copied nodes.
// Code reviewers should check that the node is new or copied by doing a find usages check on this method.
func (node *MemNode) rotateRight(ctx *mutationContext) (*MemNode, error) {
	left, leftPin, err := node.left.Resolve()
	defer leftPin.Unpin()
	if err != nil {
		return nil, err
	}
	newRoot, err := ctx.mutateBranch(left)
	if err != nil {
		return nil, err
	}

	// move left's right subtree (B) to node's left
	node.left = left.Right()
	// node becomes the right child of the new root
	newRoot.right = NewNodePointer(node)

	// update node's size/height first (it's now a child), then the new root's
	if err := node.updateHeightSize(); err != nil {
		return nil, err
	}
	if err := newRoot.updateHeightSize(); err != nil {
		return nil, err
	}

	return newRoot, nil
}

// rotateLeft performs a left rotation around this node.
//
// A new node is created by copying the right child, which becomes
// the new root of this subtree. The original right child is orphaned.
// The current node (which must be new and mutable) becomes the left child of the new root.
// The original right child's left subtree becomes the current node's new right subtree.
//
//	Q (mutable)                  copy of R
//	  /  \                        /      \
//	[X]   R      =>       Q (immutable)  [Z]     +    orphans: original R
//	     / \                   /  \
//	   [Y] [Z]               [X]  [Y]
//
// After rotation, heights and sizes are recalculated bottom-up (Y first,
// then the new root) since Y is now a child of the new root.
//
// IMPORTANT: This method must only be called on newly created or copied nodes.
// Code reviewers should check that the node is new or copied by doing a find usages check on this method.
func (node *MemNode) rotateLeft(ctx *mutationContext) (*MemNode, error) {
	right, rightPin, err := node.right.Resolve()
	defer rightPin.Unpin()
	if err != nil {
		return nil, err
	}

	newRoot, err := ctx.mutateBranch(right)
	if err != nil {
		return nil, err
	}

	// move right's left subtree (B) to node's right
	node.right = right.Left()
	// node becomes the left child of the new root
	newRoot.left = NewNodePointer(node)

	// update node's height/size first (it's now a child), then the new root's
	if err := node.updateHeightSize(); err != nil {
		return nil, err
	}
	if err := newRoot.updateHeightSize(); err != nil {
		return nil, err
	}

	return newRoot, nil
}
