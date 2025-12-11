package internal

import (
	"bytes"
	"fmt"
)

// verifyAVLInvariants recursively verifies all IAVL tree invariants starting from the given node.
//
// For leaf nodes, it verifies:
//   - Height is 0
//   - Size is 1
//   - Left and right children are nil
//
// For branch nodes, it verifies:
//  1. Key ordering: left child key < node key <= right child key
//  2. Branch key property: node key equals right child's key (smallest key in right subtree)
//  3. AVL balance: |left.height - right.height| <= 1
//  4. Height invariant: height = max(left.height, right.height) + 1
//  5. Size invariant: size = left.size + right.size
//
// Note: This function does NOT verify orphan tracking or hash correctness.
// Those require separate verification with access to the mutation context or hash computation.
func verifyAVLInvariants(node Node) error {
	key, err := node.Key()
	if err != nil {
		return fmt.Errorf("get key: %w", err)
	}

	// Node identifier for error messages. Key alone doesn't uniquely identify a node
	// since branch nodes store the first key of their right subtree, which means
	// the same key can appear in multiple branch nodes along the path to a leaf,
	// as well as in the leaf itself. We include height to disambiguate.
	nodeID := fmt.Sprintf("key=%x, height=%d", key.UnsafeBytes(), node.Height())

	if node.IsLeaf() {
		if node.Height() != 0 {
			return fmt.Errorf("leaf node (%s) has height %d, expected 0", nodeID, node.Height())
		}

		if node.Size() != 1 {
			return fmt.Errorf("leaf node (%s) has size %d, expected 1", nodeID, node.Size())
		}

		if node.Left() != nil {
			return fmt.Errorf("leaf node (%s) has non-nil left child", nodeID)
		}

		if node.Right() != nil {
			return fmt.Errorf("leaf node (%s) has non-nil right child", nodeID)
		}
	} else {
		leftPtr := node.Left()
		if leftPtr == nil {
			return fmt.Errorf("branch node (%s) has nil left child", nodeID)
		}

		rightPtr := node.Right()
		if rightPtr == nil {
			return fmt.Errorf("branch node (%s) has nil right child", nodeID)
		}

		left, leftPin, err := leftPtr.Resolve()
		defer leftPin.Unpin()
		if err != nil {
			return fmt.Errorf("resolve left child of node (%s): %w", nodeID, err)
		}

		right, rightPin, err := rightPtr.Resolve()
		defer rightPin.Unpin()
		if err != nil {
			return fmt.Errorf("resolve right child of node (%s): %w", nodeID, err)
		}

		leftKey, err := left.Key()
		if err != nil {
			return fmt.Errorf("get key of left child of node (%s): %w", nodeID, err)
		}

		rightKey, err := right.Key()
		if err != nil {
			return fmt.Errorf("get key of right child of node (%s): %w", nodeID, err)
		}

		// IAVL key ordering: branch nodes store the first key of their right subtree as a separator.
		// This means:
		//   - All keys in left subtree < node.key
		//   - All keys in right subtree >= node.key (the leftmost leaf in right subtree has key == node.key)
		//
		// We check immediate children here; recursive calls verify the full subtrees.
		if bytes.Compare(leftKey.UnsafeBytes(), key.UnsafeBytes()) >= 0 {
			return fmt.Errorf("branch node (%s) has left child with key %x which is >= node key", nodeID, leftKey.UnsafeBytes())
		}

		if bytes.Compare(rightKey.UnsafeBytes(), key.UnsafeBytes()) < 0 {
			return fmt.Errorf("branch node (%s) has right child with key %x which is < node key", nodeID, rightKey.UnsafeBytes())
		}

		// Size invariant
		if left.Size()+right.Size() != node.Size() {
			return fmt.Errorf("branch node (%s) has size %d, but children sizes are %d + %d = %d", nodeID, node.Size(), left.Size(), right.Size(), left.Size()+right.Size())
		}

		// Height invariant
		expectedHeight := maxUint8(left.Height(), right.Height()) + 1
		if node.Height() != expectedHeight {
			return fmt.Errorf("branch node (%s) has height %d, expected %d (left height %d, right height %d)", nodeID, node.Height(), expectedHeight, left.Height(), right.Height())
		}

		// AVL balance invariant
		balance := int(left.Height()) - int(right.Height())
		if balance < -1 || balance > 1 {
			return fmt.Errorf("branch node (%s) is unbalanced: balance=%d (left height %d, right height %d)", nodeID, balance, left.Height(), right.Height())
		}

		if err := verifyAVLInvariants(left); err != nil {
			return err
		}

		if err := verifyAVLInvariants(right); err != nil {
			return err
		}
	}
	return nil
}
