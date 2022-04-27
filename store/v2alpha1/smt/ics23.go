// Here we implement proof generation according to the ICS-23 specification:
// https://github.com/cosmos/ibc/tree/master/spec/core/ics-023-vector-commitments

package smt

import (
	"crypto/sha256"
	"fmt"

	dbm "github.com/cosmos/cosmos-sdk/db"

	ics23 "github.com/confio/ics23/go"
)

func createIcs23Proof(store *Store, key []byte) (*ics23.CommitmentProof, error) {
	ret := &ics23.CommitmentProof{}
	path := sha256.Sum256(key)
	has, err := store.tree.Has(key)
	if err != nil {
		return nil, err
	}
	if has { // Membership proof
		value, err := store.values.Get(path[:])
		if err != nil {
			return nil, err
		}
		if value == nil {
			return nil, fmt.Errorf("value not found for: %v", key)
		}
		proof, err := store.tree.Prove(key)
		if err != nil {
			return nil, err
		}
		ret.Proof = &ics23.CommitmentProof_Exist{&ics23.ExistenceProof{
			Key:   path[:],
			Value: value,
			Leaf:  ics23.SmtSpec.LeafSpec,
			Path:  convertInnerOps(path[:], proof.SideNodes),
		}}
	} else { // Non-membership
		nonexist, err := toNonExistenceProof(store, path)
		if err != nil {
			return nil, err
		}
		ret.Proof = &ics23.CommitmentProof_Nonexist{nonexist}
	}
	return ret, nil
}

func toNonExistenceProof(store *Store, path [32]byte) (*ics23.NonExistenceProof, error) {
	// Seek to our neighbors via the backing DB
	getNext := func(it dbm.Iterator) (*ics23.ExistenceProof, error) {
		defer it.Close()
		if it.Next() {
			value, err := store.values.Get(it.Key())
			if err != nil {
				return nil, err
			}
			if value == nil {
				return nil, fmt.Errorf("value not found for: %v", it.Value())
			}
			proof, err := store.tree.Prove(it.Value()) // pass the preimage to Prove
			if err != nil {
				return nil, err
			}
			return &ics23.ExistenceProof{
				Key:   it.Key(),
				Value: value,
				Leaf:  ics23.SmtSpec.LeafSpec,
				Path:  convertInnerOps(it.Key(), proof.SideNodes),
			}, nil
		}
		return nil, nil
	}
	var lproof, rproof *ics23.ExistenceProof
	it, err := store.preimages.ReverseIterator(nil, path[:])
	if err != nil {
		return nil, err
	}
	lproof, err = getNext(it)
	if err != nil {
		return nil, err
	}
	it, err = store.preimages.Iterator(path[:], nil)
	if err != nil {
		return nil, err
	}
	rproof, err = getNext(it)
	if err != nil {
		return nil, err
	}
	return &ics23.NonExistenceProof{
		Key:   path[:],
		Left:  lproof,
		Right: rproof,
	}, nil
}

func convertInnerOps(path []byte, sideNodes [][]byte) []*ics23.InnerOp {
	depth := len(sideNodes)
	inners := make([]*ics23.InnerOp, 0, depth)
	for i := 0; i < len(sideNodes); i++ {
		op := &ics23.InnerOp{
			Hash:   ics23.HashOp_SHA256,
			Prefix: []byte{1},
		}
		if getBitAtFromMSB(path[:], depth-1-i) == 1 {
			// right child is on path
			op.Prefix = append(op.Prefix, sideNodes[i]...)
		} else {
			op.Suffix = sideNodes[i]
		}
		inners = append(inners, op)
	}
	return inners
}

// getBitAtFromMSB gets the bit at an offset from the most significant bit
func getBitAtFromMSB(data []byte, position int) int {
	if int(data[position/8])&(1<<(8-1-uint(position)%8)) > 0 {
		return 1
	}
	return 0
}
