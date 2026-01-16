package iavlx

import (
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
	root, err := tree.root.Resolve()
	if err != nil {
		return nil, err
	}
	exist, err := createExistenceProof(root, key)
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

/*
GetNonMembershipProof will produce a CommitmentProof that the given key doesn't exist in the iavl tree.
If the key exists in the tree, this will return an error.
*/
func (t *ImmutableTree) GetNonMembershipProof(key []byte) (*ics23.CommitmentProof, error) {
	// idx is one node right of what we want....
	exists := t.Has(key)
	if exists {
		return nil, errors.New("cannot create non-membership proof with key that exists in tree")
	}

	nonexist := &ics23.NonExistenceProof{
		Key: key,
	}

	root, err := t.root.Resolve()
	if err != nil {
		return nil, err
	}
	idx, err := nextIndex(root, key)
	if err != nil {
		return nil, err
	}

	if idx >= 1 {
		leftNode, err := getByIndex(root, idx-1)
		if err != nil {
			return nil, err
		}
		leftKey, err := leftNode.Key()
		if err != nil {
			return nil, err
		}

		nonexist.Left, err = createExistenceProof(root, leftKey)
		if err != nil {
			return nil, err
		}
	}

	// this will be nil if nothing right of the queried key
	rightNode, err := getByIndex(root, idx)
	if err != nil {
		return nil, err
	}
	if rightNode != nil {
		rightKey, err := rightNode.Key()

		if rightKey != nil {
			nonexist.Right, err = createExistenceProof(root, rightKey)
			if err != nil {
				return nil, err
			}
		}
	}

	proof := &ics23.CommitmentProof{
		Proof: &ics23.CommitmentProof_Nonexist{
			Nonexist: nonexist,
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

// VerifyNonMembership returns true iff proof is a NonExistenceProof for the given key.
func (t *ImmutableTree) VerifyNonMembership(proof *ics23.CommitmentProof, key []byte) (bool, error) {
	root, err := t.root.Resolve()
	if err != nil {
		return false, err
	}
	rootHash := root.Hash()

	return ics23.VerifyNonMembership(ics23.IavlSpec, rootHash, proof, key), nil
}

var (
	_ storetypes.KVStore = (*ImmutableTree)(nil)
)
