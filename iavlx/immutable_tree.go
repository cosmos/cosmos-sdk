package iavlx

import (
	"bytes"
	"encoding/binary"
	"errors"
	io "io"

	ics23 "github.com/cosmos/ics23/go"

	storetypes "cosmossdk.io/store/types"
)

type ImmutableTree struct {
	root *NodePointer
}

func NewImmutableTree(root *NodePointer) *ImmutableTree {
	return &ImmutableTree{
		root: root,
	}
}

func (tree *ImmutableTree) GetStoreType() storetypes.StoreType {
	return storetypes.StoreTypeIAVL
}

func (tree *ImmutableTree) CacheWrap() storetypes.CacheWrap {
	return NewCacheTree(tree)
}

func (tree *ImmutableTree) CacheWrapWithTrace(w io.Writer, tc storetypes.TraceContext) storetypes.CacheWrap {
	// TODO support tracing
	return tree.CacheWrap()
}

func (tree *ImmutableTree) Get(key []byte) []byte {
	if tree.root == nil {
		return nil
	}

	root, err := tree.root.Resolve()
	if err != nil {
		panic(err)
	}

	value, _, err := root.Get(key)
	if err != nil {
		panic(err)
	}

	return value
}

func (tree *ImmutableTree) Set(key, value []byte) {
	panic("cannot set in immutable tree")
}

func (tree *ImmutableTree) Delete(key []byte) {
	panic("cannot delete from immutable tree")
}

func (tree *ImmutableTree) Has(key []byte) bool {
	val := tree.Get(key)
	return val != nil
}

func (tree *ImmutableTree) Iterator(start, end []byte) storetypes.Iterator {
	return NewIterator(start, end, true, tree.root, true)
}

func (tree *ImmutableTree) ReverseIterator(start, end []byte) storetypes.Iterator {
	return NewIterator(start, end, false, tree.root, true)
}

func (tree *ImmutableTree) GetMembershipProof(key []byte) (*ics23.CommitmentProof, error) {
	exist, err := tree.createExistenceProof(key)
	if err != nil {
		return nil, err
	}
	proof := &ics23.CommitmentProof{
		Proof: &ics23.CommitmentProof_Exist{
			Exist: exist,
		},
	}
	return proof, nil
}

// VerifyMembership returns true iff proof is an ExistenceProof for the given key.
func (t *ImmutableTree) VerifyMembership(proof *ics23.CommitmentProof, key []byte) (bool, error) {
	val := t.Get(key)
	if val == nil {
		return false, errors.New("key not found")
	}
	rootNode, err := t.root.Resolve()
	if err != nil {
		return false, err
	}

	root := rootNode.Hash()

	return ics23.VerifyMembership(ics23.IavlSpec, root, proof, key, val), nil
}

func (tree *ImmutableTree) createExistenceProof(key []byte) (*ics23.ExistenceProof, error) {
	nodeVersion := tree.root.id.Version()
	path := new(PathToLeaf)
	root, err := tree.root.Resolve()
	if err != nil {
		return nil, err
	}
	node, err := pathToLeaf(tree, root, key, nodeVersion, path)
	if err != nil {
		return nil, err
	}
	nodeVersion = node.ID().Version()

	nodeKey, err := node.Key()
	if err != nil {
		return nil, err
	}

	nodeValue, err := node.Value()
	if err != nil {
		return nil, err
	}
	return &ics23.ExistenceProof{
		Key:   nodeKey,
		Value: nodeValue,
		Leaf:  convertLeafOp(nodeVersion),
		Path:  convertInnerOps(*path),
	}, nil
}

func convertLeafOp(version uint64) *ics23.LeafOp {
	var varintBuf [binary.MaxVarintLen64]byte
	// this is adapted from iavl/proof.go:proofLeafNode.Hash()
	prefix := convertVarIntToBytes(0, varintBuf)
	prefix = append(prefix, convertVarIntToBytes(1, varintBuf)...)
	prefix = append(prefix, convertVarIntToBytes(int64(version), varintBuf)...)

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

func convertInnerOps(path PathToLeaf) []*ics23.InnerOp {
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
		prefix = append(prefix, convertVarIntToBytes(int64(path[i].Version), varintBuf)...)

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

func pathToLeaf(tree *ImmutableTree, node Node, key []byte, version uint64, path *PathToLeaf) (Node, error) {
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
	nodeVersion := node.ID().Version()
	if bytes.Compare(key, nodeKey) < 0 {
		// left side
		rightNodePtr := node.Right()
		rightNode, err := rightNodePtr.Resolve()
		if err != nil {
			return nil, err
		}
		pin := ProofInnerNode{
			Height:  node.Height(),
			Size:    node.Size(),
			Version: nodeVersion,
			Left:    nil,
			Right:   rightNode.Hash(),
		}
		*path = append(*path, pin)

		leftNodePtr := node.Left()

		leftNode, err := leftNodePtr.Resolve()
		if err != nil {
			return nil, err
		}
		n, err := pathToLeaf(tree, leftNode, key, version, path)
		return n, err
	}
	// right side
	leftNode, err := node.Left().Resolve()
	if err != nil {
		return nil, err
	}
	pin := ProofInnerNode{
		Height:  node.Height(),
		Size:    node.Size(),
		Version: nodeVersion,
		Left:    leftNode.Hash(),
		Right:   nil,
	}
	*path = append(*path, pin)

	rightNode, err := node.Right().Resolve()
	if err != nil {
		return nil, err
	}

	n, err := pathToLeaf(tree, rightNode, key, version, path)
	return n, err
}

var (
	_ storetypes.KVStore = (*ImmutableTree)(nil)
)
