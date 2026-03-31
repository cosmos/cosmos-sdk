package internal

import "iter"

// Export returns an iterator that yields all nodes in the tree in post-order
// (left subtree, right subtree, parent). This is the standard IAVL export order
// and is compatible with the Importer's expected input for tree reconstruction.
func (t TreeReader) Export() iter.Seq2[ExportNode, error] {
	rootPtr := t.root
	if rootPtr == nil {
		return func(yield func(ExportNode, error) bool) {}
	}

	return func(yield func(ExportNode, error) bool) {
		exportSubTree(rootPtr, yield)
	}
}

func exportSubTree(ptr *NodePointer, yield func(ExportNode, error) bool) bool {
	node, pin, err := ptr.Resolve()
	defer pin.Unpin()
	if err != nil {
		yield(ExportNode{}, err)
		return false
	}

	key, err := node.Key()
	if err != nil {
		yield(ExportNode{}, err)
		return false
	}

	exportNode := ExportNode{
		Key:     key.SafeCopy(),
		Version: int64(node.Version()),
		Height:  int8(node.Height()),
	}

	if node.IsLeaf() {
		value, err := node.Value()
		if err != nil {
			yield(ExportNode{}, err)
			return false
		}

		exportNode.Value = value.SafeCopy()
		if !yield(exportNode, nil) {
			return false
		}
	} else {
		if !exportSubTree(node.Left(), yield) {
			return false
		}
		if !exportSubTree(node.Right(), yield) {
			return false
		}

		if !yield(exportNode, nil) {
			return false
		}
	}

	return true
}
