package internal

import (
	"bytes"
	"fmt"
)

func verifyAVLInvariants(np *NodePointer) error {
	node, pin, err := np.Resolve()
	defer pin.Unpin()
	if err != nil {
		return fmt.Errorf("resolve node %s: %w", np.id, err)
	}

	if node.Version() != np.id.Version {
		return fmt.Errorf("node %s has version %d, expected %d", np.id, node.Version(), np.id.Version)
	}

	if node.IsLeaf() {
		if node.Height() != 0 {
			return fmt.Errorf("leaf node %s has height %d", np.id, node.Height())
		}

		if node.Size() != 1 {
			return fmt.Errorf("leaf node %s has size %d, expected 1", np.id, node.Size())
		}

		if node.Left() != nil {
			return fmt.Errorf("leaf node %s has non-nil left child", np.id)
		}

		if node.Right() != nil {
			return fmt.Errorf("leaf node %s has non-nil right child", np.id)
		}
	} else {
		leftPtr := node.Left()
		if leftPtr == nil {
			return fmt.Errorf("branch node %s has nil left child", np.id)
		}

		rightPtr := node.Right()
		if rightPtr == nil {
			return fmt.Errorf("branch node %s has nil right child", np.id)
		}

		left, leftPin, err := leftPtr.Resolve()
		defer leftPin.Unpin()
		if err != nil {
			return fmt.Errorf("resolve left child of node %s: %w", np.id, err)
		}

		right, rightPin, err := rightPtr.Resolve()
		defer rightPin.Unpin()
		if err != nil {
			return fmt.Errorf("resolve right child of node %s: %w", np.id, err)
		}

		key, err := node.Key()
		if err != nil {
			return fmt.Errorf("get key of node %s: %w", np.id, err)
		}

		leftKey, err := left.Key()
		if err != nil {
			return fmt.Errorf("get key of left child of node %s: %w", np.id, err)
		}

		rightKey, err := right.Key()
		if err != nil {
			return fmt.Errorf("get key of right child of node %s: %w", np.id, err)
		}

		if bytes.Compare(leftKey.UnsafeBytes(), key.UnsafeBytes()) >= 0 {
			return fmt.Errorf("branch node %s with id %s has key %x, but left child %s, has key %x", node, np.id, key, left, leftKey)
		}

		if bytes.Compare(rightKey.UnsafeBytes(), key.UnsafeBytes()) < 0 {
			return fmt.Errorf("branch node %s with id %s has key %x, but right child %s, has key %x", node, np.id, key, right, rightKey)
		}

		if left.Size()+right.Size() != node.Size() {
			return fmt.Errorf("branch node %s has size %d, but children sizes are %d and %d", np.id, node.Size(), left.Size(), right.Size())
		}

		expectedHeight := maxUint8(left.Height(), right.Height()) + 1
		if node.Height() != expectedHeight {
			return fmt.Errorf("branch node %s has height %d, expected %d, left height %d, right height %d", np.id, node.Height(), expectedHeight, left.Height(), right.Height())
		}

		// ensure balanced
		balance := int(left.Height()) - int(right.Height())
		if balance < -1 || balance > 1 {
			return fmt.Errorf("branch node %s is unbalanced: left height %d, right height %d", np.id, left.Height(), right.Height())
		}

		if err := verifyAVLInvariants(leftPtr); err != nil {
			return err
		}

		if err := verifyAVLInvariants(rightPtr); err != nil {
			return err
		}
	}
	return nil
}
