package internal

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

// updateHeightSize updates the height and size of the node based on its children.
// This is only used during insertion, deletion, and rebalancing.
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

// maxUint8 returns the maximum of two uint8 values.
func maxUint8(a, b uint8) uint8 {
	if a > b {
		return a
	}
	return b
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
