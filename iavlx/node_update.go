package iavlx

import "bytes"

// setRecursive do set operation.
// it always do modification and return new `MemNode`, even if the value is the same.
// also returns if it's an update or insertion, if update, the tree height and balance is not changed.
func setRecursive(nodePtr *NodePointer, leafNode *MemNode, ctx *MutationContext) (*NodePointer, bool, error) {
	if nodePtr == nil {
		return NewNodePointer(leafNode), true, nil
	}

	node, err := nodePtr.Resolve()
	if err != nil {
		return nil, false, err
	}

	nodeKey, err := node.Key()
	if err != nil {
		return nil, false, err
	}
	if node.IsLeaf() {
		leafNodePtr := NewNodePointer(leafNode)
		cmp := bytes.Compare(leafNode.key, nodeKey)
		if cmp == 0 {
			ctx.AddOrphan(nodePtr.id)
			return leafNodePtr, true, nil
		}
		n := &MemNode{
			height:  1,
			size:    2,
			version: ctx.Version,
		}
		switch cmp {
		case -1:
			n.left = leafNodePtr
			n.right = nodePtr
			n.key = nodeKey
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
		if bytes.Compare(leafNode.key, nodeKey) == -1 {
			newChildPtr, updated, err = setRecursive(node.Left(), leafNode, ctx)
			if err != nil {
				return nil, false, err
			}
			newNode, err = ctx.MutateBranch(node)
			if err != nil {
				return nil, false, err
			}
			newNode.left = newChildPtr
		} else {
			newChildPtr, updated, err = setRecursive(node.Right(), leafNode, ctx)
			if err != nil {
				return nil, false, err
			}
			newNode, err = ctx.MutateBranch(node)
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

type newKeyWrapper struct {
	key []byte
	// keyRef keyRefLink
}

// removeRecursive returns:
// - (nil, origNode, nil) -> nothing changed in subtree
// - (value, nil, nil) -> leaf node is removed
// - (value, new node, newKey) -> subtree changed
func removeRecursive(nodePtr *NodePointer, key []byte, ctx *MutationContext) (value []byte, newNodePtr *NodePointer, newKey *newKeyWrapper, err error) {
	if nodePtr == nil {
		return nil, nil, nil, nil
	}

	node, err := nodePtr.Resolve()
	if err != nil {
		return nil, nil, nil, err
	}

	nodeKey, err := node.Key()
	if err != nil {
		return nil, nil, nil, err
	}

	if node.IsLeaf() {
		if bytes.Equal(nodeKey, key) {
			ctx.AddOrphan(nodePtr.id)
			value, err := node.Value()
			return value, nil, nil, err
		}
		return nil, nodePtr, nil, nil
	}

	if bytes.Compare(key, nodeKey) == -1 {
		value, newLeft, newKey, err := removeRecursive(node.Left(), key, ctx)
		if err != nil {
			return nil, nil, nil, err
		}

		if value == nil {
			return nil, nodePtr, nil, nil
		}

		if newLeft == nil {
			ctx.AddOrphan(nodePtr.id)
			return value, node.Right(), &newKeyWrapper{
				key: nodeKey,
				// keyRef: nodePtr,
			}, nil
		}

		newNode, err := ctx.MutateBranch(node)
		if err != nil {
			return nil, nil, nil, err
		}
		newNode.left = newLeft
		err = newNode.updateHeightSize()
		if err != nil {
			return nil, nil, nil, err
		}
		newNode, err = newNode.reBalance(ctx)
		if err != nil {
			return nil, nil, nil, err
		}

		return value, NewNodePointer(newNode), newKey, nil
	}

	value, newRight, newKey, err := removeRecursive(node.Right(), key, ctx)
	if err != nil {
		return nil, nil, nil, err
	}

	if value == nil {
		return nil, nodePtr, nil, nil
	}

	if newRight == nil {
		ctx.AddOrphan(nodePtr.id)
		return value, node.Left(), nil, nil
	}

	newNode, err := ctx.MutateBranch(node)
	if err != nil {
		return nil, nil, nil, err
	}

	newNode.right = newRight
	if newKey != nil {
		newNode.key = newKey.key
		// newNode._keyRef = newKey.keyRef
	}

	err = newNode.updateHeightSize()
	if err != nil {
		return nil, nil, nil, err
	}

	newNode, err = newNode.reBalance(ctx)
	if err != nil {
		return nil, nil, nil, err
	}

	return value, NewNodePointer(newNode), nil, nil
}

// IMPORTANT: nodes called with this method must be new or copies first.
// Code reviewers should use find usages to ensure that all callers follow this rule!
func (node *MemNode) updateHeightSize() error {
	leftNode, err := node.left.Resolve()
	if err != nil {
		return err
	}

	rightNode, err := node.right.Resolve()
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
func (node *MemNode) reBalance(ctx *MutationContext) (*MemNode, error) {
	balance, err := calcBalance(node)
	if err != nil {
		return nil, err
	}
	switch {
	case balance > 1:
		left, err := node.left.Resolve()
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
		newLeft, err := ctx.MutateBranch(left)
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
		right, err := node.right.Resolve()
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
		newRight, err := ctx.MutateBranch(right)
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
	leftNode, err := node.Left().Resolve()
	if err != nil {
		return 0, err
	}

	rightNode, err := node.Right().Resolve()
	if err != nil {
		return 0, err
	}

	return int(leftNode.Height()) - int(rightNode.Height()), nil
}

// IMPORTANT: nodes called with this method must be new or copies first.
// Code reviewers should use find usages to ensure that all callers follow this rule!
func (node *MemNode) rotateRight(ctx *MutationContext) (*MemNode, error) {
	left, err := node.left.Resolve()
	if err != nil {
		return nil, err
	}
	newSelf, err := ctx.MutateBranch(left)
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
func (node *MemNode) rotateLeft(ctx *MutationContext) (*MemNode, error) {
	right, err := node.right.Resolve()
	if err != nil {
		return nil, err
	}

	newSelf, err := ctx.MutateBranch(right)
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
