package internal

import (
	"errors"
	"io"

	"cosmossdk.io/store/cachekv"
	storetypes "cosmossdk.io/store/types"
	ics23 "github.com/cosmos/ics23/go"
)

type TreeReader struct {
	version uint32
	root    *NodePointer
}

func NewTreeReader(version uint32, root *NodePointer) TreeReader {
	return TreeReader{version: version, root: root}
}

func (t TreeReader) HasErr(key []byte) (bool, error) {
	if t.root == nil {
		return false, nil
	}
	root, pin, err := t.root.Resolve()
	defer pin.Unpin()
	if err != nil {
		return false, err
	}
	if root == nil {
		return false, nil
	}
	has, _, err := root.Has(key)
	return has, err
}

func (t TreeReader) GetErr(key []byte) ([]byte, error) {
	if t.root == nil {
		return nil, nil
	}
	root, pin, err := t.root.Resolve()
	defer pin.Unpin()
	if err != nil {
		return nil, err
	}
	if root == nil {
		return nil, nil
	}
	value, _, err := root.Get(key)
	if err != nil {
		return nil, err
	}
	return value.SafeCopy(), nil
}

func (t TreeReader) Size() int64 {
	if t.root == nil {
		return 0
	}
	root, pin, err := t.root.Resolve()
	defer pin.Unpin()
	if err != nil {
		return 0
	}
	return root.Size()
}

func (t TreeReader) Iterator(start, end []byte) storetypes.Iterator {
	return NewIterator(start, end, true, t.root)
}

func (t TreeReader) ReverseIterator(start, end []byte) storetypes.Iterator {
	return NewIterator(start, end, false, t.root)
}

func (t TreeReader) GetStoreType() storetypes.StoreType {
	return storetypes.StoreTypeIAVL
}

func (t TreeReader) CacheWrap() storetypes.CacheWrap {
	return cachekv.NewStore(t)
}

func (t TreeReader) CacheWrapWithTrace(io.Writer, storetypes.TraceContext) storetypes.CacheWrap {
	// TODO implement me
	return cachekv.NewStore(t)
}

func (t TreeReader) Get(key []byte) []byte {
	value, err := t.GetErr(key)
	if err != nil {
		panic(err)
	}
	return value
}

func (t TreeReader) Has(key []byte) bool {
	found, err := t.HasErr(key)
	if err != nil {
		panic(err)
	}
	return found
}

func (t TreeReader) Set([]byte, []byte) {
	panic("readonly store: cannot set value")
}

func (t TreeReader) Delete([]byte) {
	panic("readonly store: cannot delete")
}

func (t TreeReader) Version() uint32 {
	return t.version
}

var _ storetypes.KVStore = TreeReader{}

// GetMembershipProof will produce a CommitmentProof that the given key exists in the iavl tree.
// If the key does NOT exist in the tree, this will return an error.
func (t TreeReader) GetMembershipProof(key []byte) (*ics23.CommitmentProof, error) {
	if t.root == nil {
		return nil, errors.New("cannot create membership proof with nil root")
	}

	root, pin, err := t.root.Resolve()
	defer pin.Unpin()
	if err != nil {
		return nil, err
	}
	if root == nil {
		return nil, errors.New("cannot create membership proof with nil root")
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

// GetNonMembershipProof will produce a CommitmentProof that the given key doesn't exist in the iavl tree.
// If the key exists in the tree, this will return an error.
func (t TreeReader) GetNonMembershipProof(key []byte) (*ics23.CommitmentProof, error) {
	if t.root == nil {
		return nil, errors.New("cannot create non-membership proof with nil root")
	}

	// idx is one node right of what we want....
	exists := t.Has(key)
	if exists {
		return nil, errors.New("cannot create non-membership proof with key that exists in tree")
	}

	nonexist := &ics23.NonExistenceProof{
		Key: key,
	}

	root, pin, err := t.root.Resolve()
	defer pin.Unpin()
	if err != nil {
		return nil, err
	}
	if root == nil {
		return nil, errors.New("cannot create non-membership proof with nil root")
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

		nonexist.Left, err = createExistenceProof(root, leftKey.UnsafeBytes())
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

		if rightKey.UnsafeBytes() != nil {
			nonexist.Right, err = createExistenceProof(root, rightKey.UnsafeBytes())
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
func (t TreeReader) VerifyMembership(proof *ics23.CommitmentProof, key []byte) (bool, error) {
	if t.root == nil {
		return false, errors.New("cannot verify membership with nil root")
	}

	val := t.Get(key)
	if val == nil {
		return false, errors.New("key not found")
	}
	rootNode, pin, err := t.root.Resolve()
	defer pin.Unpin()
	if err != nil {
		return false, err
	}
	if rootNode == nil {
		return false, errors.New("cannot verify membership with nil root")
	}

	root := rootNode.Hash().UnsafeBytes()

	return ics23.VerifyMembership(ics23.IavlSpec, root, proof, key, val), nil
}

// VerifyNonMembership returns true iff proof is a NonExistenceProof for the given key.
func (t TreeReader) VerifyNonMembership(proof *ics23.CommitmentProof, key []byte) (bool, error) {
	if t.root == nil {
		return false, errors.New("cannot verify non-membership with nil root")
	}

	root, pin, err := t.root.Resolve()
	defer pin.Unpin()
	if err != nil {
		return false, err
	}
	if root == nil {
		return false, errors.New("cannot verify non-membership with nil root")
	}
	rootHash := root.Hash().UnsafeBytes()

	return ics23.VerifyNonMembership(ics23.IavlSpec, rootHash, proof, key), nil
}
