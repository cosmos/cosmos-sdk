# Node

The Node struct stores a node in the IAVL tree. 

### Structure

```golang
// Node represents a node in a Tree.
type Node struct {
	key       []byte // key for the node.
	value     []byte // value of leaf node. If inner node, value = nil
	version   int64  // The version of the IAVL that this node was first added in.
	height    int8   // The height of the node. Leaf nodes have height 0
	size      int64  // The number of leaves that are under the current node. Leaf nodes have size = 1
	hash      []byte // hash of above field and leftHash, rightHash
	leftHash  []byte // hash of left child
	leftNode  *Node  // pointer to left child
        rightHash []byte // hash of right child
	rightNode *Node  // pointer to right child
	persisted bool   // persisted to disk
}
```

Inner nodes have keys equal to the highest key on their left branch and have values set to nil.

The version of a node is the first version of the IAVL tree that the node gets added in. Future versions of the IAVL may point to this node if they also contain the node, however the node's version itself does not change.

Size is the number of leaves under a given node. With a full subtree, `node.size = 2^(node.height)`.

### Marshaling 

Every node is persisted by encoding the key, version, height, size and hash. If the node is a leaf node, then the value is persisted as well. If the node is not a leaf node, then the leftHash and rightHash are persisted as well.

```golang
// Writes the node as a serialized byte slice to the supplied io.Writer.
func (node *Node) writeBytes(w io.Writer) error {
	cause := encodeVarint(w, node.height)
	if cause != nil {
		return errors.Wrap(cause, "writing height")
	}
	cause = encodeVarint(w, node.size)
	if cause != nil {
		return errors.Wrap(cause, "writing size")
	}
	cause = encodeVarint(w, node.version)
	if cause != nil {
		return errors.Wrap(cause, "writing version")
	}

	// Unlike writeHashBytes, key is written for inner nodes.
	cause = encodeBytes(w, node.key)
	if cause != nil {
		return errors.Wrap(cause, "writing key")
	}

	if node.isLeaf() {
		cause = encodeBytes(w, node.value)
		if cause != nil {
			return errors.Wrap(cause, "writing value")
		}
	} else {
		if node.leftHash == nil {
			panic("node.leftHash was nil in writeBytes")
		}
		cause = encodeBytes(w, node.leftHash)
		if cause != nil {
			return errors.Wrap(cause, "writing left hash")
		}

		if node.rightHash == nil {
			panic("node.rightHash was nil in writeBytes")
		}
		cause = encodeBytes(w, node.rightHash)
		if cause != nil {
			return errors.Wrap(cause, "writing right hash")
		}
	}
	return nil
}
```

### Hashes

A node's hash is calculated by hashing the height, size, and version of the node. If the node is a leaf node, then the key and value are also hashed. If the node is an inner node, the leftHash and rightHash are included in hash but the key is not.

```golang
// Writes the node's hash to the given io.Writer. This function expects
// child hashes to be already set.
func (node *Node) writeHashBytes(w io.Writer) error {
	err := encodeVarint(w, node.height)
	if err != nil {
		return errors.Wrap(err, "writing height")
	}
	err = encodeVarint(w, node.size)
	if err != nil {
		return errors.Wrap(err, "writing size")
	}
	err = encodeVarint(w, node.version)
	if err != nil {
		return errors.Wrap(err, "writing version")
	}

	// Key is not written for inner nodes, unlike writeBytes.

	if node.isLeaf() {
		err = encodeBytes(w, node.key)
		if err != nil {
			return errors.Wrap(err, "writing key")
		}
		// Indirection needed to provide proofs without values.
		// (e.g. proofLeafNode.ValueHash)
		valueHash := tmhash.Sum(node.value)
		err = encodeBytes(w, valueHash)
		if err != nil {
			return errors.Wrap(err, "writing value")
		}
	} else {
		if node.leftHash == nil || node.rightHash == nil {
			panic("Found an empty child hash")
		}
		err = encodeBytes(w, node.leftHash)
		if err != nil {
			return errors.Wrap(err, "writing left hash")
		}
		err = encodeBytes(w, node.rightHash)
		if err != nil {
			return errors.Wrap(err, "writing right hash")
		}
	}

	return nil
}
```
