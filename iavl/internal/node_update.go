package internal

import "bytes"

// SetRecursive do set operation.
// it always do modification and return new `MemNode`, even if the value is the same.
// also returns if it's an update or insertion, if update, the tree height and balance is not changed.
func SetRecursive(nodePtr *NodePointer, leafNode *MemNode, ctx *MutationContext) (*NodePointer, bool, error) {
	if nodePtr == nil {
		return NewNodePointer(leafNode), true, nil
	}

	node, pin, err := nodePtr.Resolve()
	defer pin.Unpin()
	if err != nil {
		return nil, false, err
	}

	if node.IsLeaf() {
		nodeKey, err := node.Key()
		if err != nil {
			return nil, false, err
		}
		leafNodePtr := NewNodePointer(leafNode)
		cmp := bytes.Compare(leafNode.key, nodeKey.UnsafeBytes())
		if cmp == 0 {
			ctx.addOrphan(nodePtr)
			return leafNodePtr, true, nil
		}
		var n *MemNode
		switch cmp {
		case -1:
			n = ctx.newBranchNode(leafNodePtr, nodePtr, nodeKey.SafeCopy())
		case 1:
			n = ctx.newBranchNode(nodePtr, leafNodePtr, leafNode.key)
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
		cmp, err := node.CmpKey(leafNode.key)
		if err != nil {
			return nil, false, err
		}
		if cmp == 1 {
			newChildPtr, updated, err = SetRecursive(node.Left(), leafNode, ctx)
			if err != nil {
				return nil, false, err
			}
			newNode, err = ctx.mutateBranch(node)
			if err != nil {
				return nil, false, err
			}
			newNode.left = newChildPtr
		} else {
			newChildPtr, updated, err = SetRecursive(node.Right(), leafNode, ctx)
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

// RemoveRecursive returns:
// - (false, origNode, nil) -> nothing changed in subtree
// - (true, nil, nil) -> leaf node is removed
// - (true, new node, newKey) -> subtree changed
// a previous version returned the value instead of the removed flag, but this is never used and it's just unnecessary copying and disk IO at this point
// if we ever want to implement a "get and remove" operation, we can add it back then, but there should be bool flag on this function to indicate whether to return the value or not
func RemoveRecursive(nodePtr *NodePointer, key []byte, ctx *MutationContext) (removed bool, newNodePtr *NodePointer, newKey []byte, err error) {
	if nodePtr == nil {
		return false, nil, nil, nil
	}

	node, pin, err := nodePtr.Resolve()
	defer pin.Unpin()
	if err != nil {
		return false, nil, nil, err
	}

	if node.IsLeaf() {
		nodeKey, err := node.Key()
		if err != nil {
			return false, nil, nil, err
		}

		if bytes.Equal(nodeKey.UnsafeBytes(), key) {
			ctx.addOrphan(nodePtr)
			return true, nil, nil, nil
		}
		return false, nodePtr, nil, nil
	}

	cmp, err := node.CmpKey(key)
	if err != nil {
		return false, nil, nil, err
	}
	if cmp == 1 {
		leftRemoved, newLeft, newKey, err := RemoveRecursive(node.Left(), key, ctx)
		if err != nil {
			return false, nil, nil, err
		}

		if !leftRemoved {
			return false, nodePtr, nil, nil
		}

		nodeKey, err := node.Key()
		if err != nil {
			return false, nil, nil, err
		}
		if newLeft == nil {
			ctx.addOrphan(nodePtr)
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

	rightRemoved, newRight, newKey, err := RemoveRecursive(node.Right(), key, ctx)
	if err != nil {
		return false, nil, nil, err
	}

	if !rightRemoved {
		return false, nodePtr, nil, nil
	}

	if newRight == nil {
		ctx.addOrphan(nodePtr)
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
