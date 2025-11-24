package iavlx

import (
	"bytes"
	"encoding/binary"
	"errors"

	ics23 "github.com/cosmos/ics23/go"
)

type proofInnerNode struct {
	Height  int8   `json:"height"`
	Size    int64  `json:"size"`
	Version int64  `json:"version"`
	Left    []byte `json:"left"`
	Right   []byte `json:"right"`
}

type leafPath []proofInnerNode

// nextIndex returns the index that would be assigned to the key.
// This method assumes the key does not exist.
// Callers are expected to check that t.Has(key) == false.
func nextIndex(node Node, key []byte) (int64, error) {
	if node == nil {
		return 0, nil
	}
	nKey, err := node.Key()
	if err != nil {
		return 0, err
	}
	if node.IsLeaf() {
		switch bytes.Compare(nKey, key) {
		case -1:
			return 1, nil
		case 1:
			return 0, nil
		default:
			return 0, nil
		}
	}
	if bytes.Compare(key, nKey) < 0 {
		leftNode, err := node.Left().Resolve()
		if err != nil {
			return 0, err
		}

		return nextIndex(leftNode, key)
	}

	rightNode, err := node.Right().Resolve()
	if err != nil {
		return 0, err
	}

	index, err := nextIndex(rightNode, key)
	if err != nil {
		return 0, err
	}

	index += node.Size() - rightNode.Size()
	return index, nil
}

func createExistenceProof(root Node, key []byte) (*ics23.ExistenceProof, error) {
	path := new(leafPath)
	leafVersion := root.Version()

	leaf, err := pathToLeaf(root, key, uint64(leafVersion), path)
	if err != nil {
		return nil, err
	}
	leafVersion = leaf.Version()

	leafKey, err := leaf.Key()
	if err != nil {
		return nil, err
	}

	leafValue, err := leaf.Value()
	if err != nil {
		return nil, err
	}
	return &ics23.ExistenceProof{
		Key:   leafKey,
		Value: leafValue,
		Leaf:  convertLeafOp(int64(leafVersion)),
		Path:  convertInnerOps(*path),
	}, nil
}

func getByIndex(node Node, index int64) (Node, error) {
	if node == nil {
		return nil, nil
	}
	if node.IsLeaf() {
		if index == 0 {
			return node, nil
		}
		return nil, nil
	}
	leftNode, err := node.Left().Resolve()
	if err != nil {
		return nil, err
	}

	if index < leftNode.Size() {
		return getByIndex(leftNode, index)
	}

	rightNode, err := node.Right().Resolve()
	if err != nil {
		return nil, err
	}

	return getByIndex(rightNode, index-leftNode.Size())
}
func convertLeafOp(version int64) *ics23.LeafOp {
	var varintBuf [binary.MaxVarintLen64]byte
	// this is adapted from iavl/proof.go:proofLeafNode.Hash()
	prefix := convertVarIntToBytes(0, varintBuf)
	prefix = append(prefix, convertVarIntToBytes(1, varintBuf)...)
	prefix = append(prefix, convertVarIntToBytes(version, varintBuf)...)

	return &ics23.LeafOp{
		Hash:         ics23.HashOp_SHA256,
		PrehashValue: ics23.HashOp_SHA256,
		Length:       ics23.LengthOp_VAR_PROTO,
		Prefix:       prefix,
	}
}

func convertVarIntToBytes(orig int64, buf [binary.MaxVarintLen64]byte) []byte {
	n := binary.PutVarint(buf[:], orig)
	return buf[:n]
}

func convertInnerOps(path leafPath) []*ics23.InnerOp {
	steps := make([]*ics23.InnerOp, 0, len(path))

	// lengthByte is the length prefix prepended to each of the sha256 sub-hashes
	var lengthByte byte = 0x20

	var varintBuf [binary.MaxVarintLen64]byte

	// we need to go in reverse order, iavl starts from root to leaf,
	// we want to go up from the leaf to the root
	for i := len(path) - 1; i >= 0; i-- {
		// this is adapted from iavl/proof.go:proofInnerNode.Hash()
		// prefix = bytes of height-size-version ++ <length>-leftHash-<length>
		// suffix = <length>-rightHash
		prefix := convertVarIntToBytes(int64(path[i].Height), varintBuf)
		prefix = append(prefix, convertVarIntToBytes(path[i].Size, varintBuf)...)
		prefix = append(prefix, convertVarIntToBytes(path[i].Version, varintBuf)...)

		var suffix []byte
		if len(path[i].Left) > 0 {
			// length prefixed left side
			prefix = append(prefix, lengthByte)
			prefix = append(prefix, path[i].Left...)
			// prepend the length prefix for child
			prefix = append(prefix, lengthByte)
		} else {
			// prepend the length prefix for child
			prefix = append(prefix, lengthByte)
			// length-prefixed right side
			suffix = []byte{lengthByte}
			suffix = append(suffix, path[i].Right...)
		}

		op := &ics23.InnerOp{
			Hash:   ics23.HashOp_SHA256,
			Prefix: prefix,
			Suffix: suffix,
		}
		steps = append(steps, op)
	}
	return steps
}

func pathToLeaf(node Node, key []byte, version uint64, path *leafPath) (Node, error) {
	nodeKey, err := node.Key()
	if err != nil {
		return nil, err
	}
	if node.IsLeaf() {
		if bytes.Equal(nodeKey, key) {
			return node, nil
		} else {
			return node, errors.New("key does not exist")
		}
	}
	nodeVersion := version
	if node.ID().Index() != 0 {
		nodeVersion = node.ID().Version()
	}
	if bytes.Compare(key, nodeKey) < 0 {
		// left side
		rightNodePtr := node.Right()
		rightNode, err := rightNodePtr.Resolve()
		if err != nil {
			return nil, err
		}
		pin := proofInnerNode{
			Height:  int8(node.Height()),
			Size:    node.Size(),
			Version: int64(nodeVersion),
			Left:    nil,
			Right:   rightNode.Hash(),
		}
		*path = append(*path, pin)

		leftNodePtr := node.Left()

		leftNode, err := leftNodePtr.Resolve()
		if err != nil {
			return nil, err
		}
		n, err := pathToLeaf(leftNode, key, version, path)
		return n, err
	}
	// right side
	leftNode, err := node.Left().Resolve()
	if err != nil {
		return nil, err
	}
	pin := proofInnerNode{
		Height:  int8(node.Height()),
		Size:    node.Size(),
		Version: int64(nodeVersion),
		Left:    leftNode.Hash(),
		Right:   nil,
	}
	*path = append(*path, pin)

	rightNode, err := node.Right().Resolve()
	if err != nil {
		return nil, err
	}

	n, err := pathToLeaf(rightNode, key, version, path)
	return n, err
}
