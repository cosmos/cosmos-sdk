package internal

import "bytes"

func newLeafNode(key, value []byte, version uint32) *MemNode {
	return &MemNode{
		height:  0,
		size:    1,
		version: version,
		key:     key,
		value:   value,
	}
}

// setRecursive do set operation.
// it always do modification and return new `MemNode`, even if the value is the same.
// also returns if it's an update or insertion, if update, the tree height and balance is not changed.
func setRecursive(nodePtr *NodePointer, leafNode *MemNode, ctx *mutationContext) (*NodePointer, bool, error) {
	if nodePtr == nil {
		return NewNodePointer(leafNode), true, nil
	}

	node, pin, err := nodePtr.Resolve()
	defer pin.Unpin()
	if err != nil {
		return nil, false, err
	}

	nodeKey, err := node.Key()
	if err != nil {
		return nil, false, err
	}
	if node.IsLeaf() {
		leafNodePtr := NewNodePointer(leafNode)
		cmp := bytes.Compare(leafNode.key, nodeKey.UnsafeBytes())
		if cmp == 0 {
			ctx.addOrphan(nodePtr.id)
			return leafNodePtr, true, nil
		}
		n := &MemNode{
			height:  1,
			size:    2,
			version: ctx.version,
		}
		switch cmp {
		case -1:
			n.left = leafNodePtr
			n.right = nodePtr
			n.key = nodeKey.SafeCopy()
			// n._keyRef = node
		case 1:
			n.left = nodePtr
			n.right = leafNodePtr
			n.key = leafNode.key
			// n._keyRef = leafNode
		default:
			panic("unreachable")
		}
		return NewNodePointer(n), false, nil
	} else {
		var (
			newChildPtr *NodePointer
			newNode     *MemNode
			updated     bool
			err         error
		)
		if bytes.Compare(leafNode.key, nodeKey.UnsafeBytes()) == -1 {
			newChildPtr, updated, err = setRecursive(node.Left(), leafNode, ctx)
			if err != nil {
				return nil, false, err
			}
			newNode, err = ctx.mutateBranch(node)
			if err != nil {
				return nil, false, err
			}
			newNode.left = newChildPtr
		} else {
			newChildPtr, updated, err = setRecursive(node.Right(), leafNode, ctx)
			if err != nil {
				return nil, false, err
			}
			newNode, err = ctx.mutateBranch(node)
			if err != nil {
				return nil, false, err
			}
			newNode.right = newChildPtr
		}

		if !updated {
			err = newNode.updateHeightSize()
			if err != nil {
				return nil, false, err
			}

			newNode, err = newNode.reBalance(ctx)
			if err != nil {
				return nil, false, err
			}
		}

		return NewNodePointer(newNode), updated, nil
	}
}

// removeRecursive returns:
// - (false, origNode, nil) -> nothing changed in subtree
// - (true, nil, nil) -> leaf node is removed
// - (true, new node, newKey) -> subtree changed
// a previous version returned the value instead of the removed flag, but this is never used and it's just unnecessary copying and disk IO at this point
// if we ever want to implement a "get and remove" operation, we can add it back then, but there should be bool flag on this function to indicate whether to return the value or not
func removeRecursive(nodePtr *NodePointer, key []byte, ctx *mutationContext) (removed bool, newNodePtr *NodePointer, newKey []byte, err error) {
	if nodePtr == nil {
		return false, nil, nil, nil
	}

	node, pin, err := nodePtr.Resolve()
	defer pin.Unpin()
	if err != nil {
		return false, nil, nil, err
	}

	nodeKey, err := node.Key()
	if err != nil {
		return false, nil, nil, err
	}

	if node.IsLeaf() {
		if bytes.Equal(nodeKey.UnsafeBytes(), key) {
			ctx.addOrphan(nodePtr.id)
			return true, nil, nil, nil
		}
		return false, nodePtr, nil, nil
	}

	if bytes.Compare(key, nodeKey.UnsafeBytes()) == -1 {
		leftRemoved, newLeft, newKey, err := removeRecursive(node.Left(), key, ctx)
		if err != nil {
			return false, nil, nil, err
		}

		if !leftRemoved {
			return false, nodePtr, nil, nil
		}

		if newLeft == nil {
			ctx.addOrphan(nodePtr.id)
			return true, node.Right(), nodeKey.SafeCopy(), nil
		}

		newNode, err := ctx.mutateBranch(node)
		if err != nil {
			return false, nil, nil, err
		}
		newNode.left = newLeft
		err = newNode.updateHeightSize()
		if err != nil {
			return false, nil, nil, err
		}
		newNode, err = newNode.reBalance(ctx)
		if err != nil {
			return false, nil, nil, err
		}

		return true, NewNodePointer(newNode), newKey, nil
	}

	rightRemoved, newRight, newKey, err := removeRecursive(node.Right(), key, ctx)
	if err != nil {
		return false, nil, nil, err
	}

	if !rightRemoved {
		return false, nodePtr, nil, nil
	}

	if newRight == nil {
		ctx.addOrphan(nodePtr.id)
		return true, node.Left(), nil, nil
	}

	newNode, err := ctx.mutateBranch(node)
	if err != nil {
		return false, nil, nil, err
	}

	newNode.right = newRight
	if newKey != nil {
		newNode.key = newKey
	}

	err = newNode.updateHeightSize()
	if err != nil {
		return false, nil, nil, err
	}

	newNode, err = newNode.reBalance(ctx)
	if err != nil {
		return false, nil, nil, err
	}

	return true, NewNodePointer(newNode), nil, nil
}

// IMPORTANT: nodes called with this method must be new or copies first.
// Code reviewers should use find usages to ensure that all callers follow this rule!
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

func maxUint8(a, b uint8) uint8 {
	if a > b {
		return a
	}
	return b
}

// IMPORTANT: nodes called with this method must be new or copies first.
// Code reviewers should use find usages to ensure that all callers follow this rule!
func (node *MemNode) reBalance(ctx *mutationContext) (*MemNode, error) {
	balance, err := calcBalance(node)
	if err != nil {
		return nil, err
	}
	switch {
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

		if leftBalance >= 0 {
			// left left
			return node.rotateRight(ctx)
		}

		// left right
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

		if rightBalance <= 0 {
			// right right
			return node.rotateLeft(ctx)
		}

		// right left
		newRight, err := ctx.mutateBranch(right)
		if err != nil {
			return nil, err
		}
		newRight, err = newRight.rotateRight(ctx)
		node.right = NewNodePointer(newRight)
		return node.rotateLeft(ctx)
	default:
		// nothing changed
		return node, err
	}
}

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

// IMPORTANT: nodes called with this method must be new or copies first.
// Code reviewers should use find usages to ensure that all callers follow this rule!
func (node *MemNode) rotateRight(ctx *mutationContext) (*MemNode, error) {
	left, leftPin, err := node.left.Resolve()
	defer leftPin.Unpin()
	if err != nil {
		return nil, err
	}
	newSelf, err := ctx.mutateBranch(left)
	if err != nil {
		return nil, err
	}
	node.left = left.Right()
	newSelf.right = NewNodePointer(node)

	err = node.updateHeightSize()
	if err != nil {
		return nil, err
	}
	err = newSelf.updateHeightSize()
	if err != nil {
		return nil, err
	}

	return newSelf, nil
}

// IMPORTANT: nodes called with this method must be new or copies first.
// Code reviewers should use find usages to ensure that all callers follow this rule!
func (node *MemNode) rotateLeft(ctx *mutationContext) (*MemNode, error) {
	right, rightPin, err := node.right.Resolve()
	defer rightPin.Unpin()
	if err != nil {
		return nil, err
	}

	newSelf, err := ctx.mutateBranch(right)
	if err != nil {
		return nil, err
	}

	node.right = right.Left()
	newSelf.left = NewNodePointer(node)

	err = node.updateHeightSize()
	if err != nil {
		return nil, err
	}

	err = newSelf.updateHeightSize()
	if err != nil {
		return nil, err
	}

	return newSelf, nil
}

type mutationContext struct {
	version uint32
	orphans []NodeID
}

func (ctx *mutationContext) mutateBranch(node Node) (*MemNode, error) {
	id := node.ID()
	if !id.IsEmpty() {
		ctx.orphans = append(ctx.orphans, id)
	}
	return node.MutateBranch(ctx.version)
}

func (ctx *mutationContext) addOrphan(id NodeID) {
	if !id.IsEmpty() {
		ctx.orphans = append(ctx.orphans, id)
	}
}
