package proofs

import (
	"errors"
	"fmt"
	"sort"

	ics23 "github.com/confio/ics23/go"

	sdkmaps "github.com/cosmos/cosmos-sdk/store/internal/maps"
)

var (
	ErrEmptyKey       = errors.New("key is empty")
	ErrEmptyKeyInData = errors.New("data contains empty key")
)

// TendermintSpec constrains the format from ics23-tendermint (crypto/merkle SimpleProof)
var TendermintSpec = &ics23.ProofSpec{
	LeafSpec: &ics23.LeafOp{
		Prefix:       []byte{0},
		Hash:         ics23.HashOp_SHA256,
		PrehashValue: ics23.HashOp_SHA256,
		Length:       ics23.LengthOp_VAR_PROTO,
	},
	InnerSpec: &ics23.InnerSpec{
		ChildOrder:      []int32{0, 1},
		MinPrefixLength: 1,
		MaxPrefixLength: 1,  // fixed prefix + one child
		ChildSize:       32, // (no length byte)
		Hash:            ics23.HashOp_SHA256,
	},
}

/*
CreateMembershipProof will produce a CommitmentProof that the given key (and queries value) exists in the iavl tree.
If the key doesn't exist in the tree, this will return an error.
*/
func CreateMembershipProof(data map[string][]byte, key []byte) (*ics23.CommitmentProof, error) {
	if len(key) == 0 {
		return nil, ErrEmptyKey
	}
	exist, err := createExistenceProof(data, key)
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
CreateNonMembershipProof will produce a CommitmentProof that the given key doesn't exist in the iavl tree.
If the key exists in the tree, this will return an error.
*/
func CreateNonMembershipProof(data map[string][]byte, key []byte) (*ics23.CommitmentProof, error) {
	if len(key) == 0 {
		return nil, ErrEmptyKey
	}
	// ensure this key is not in the store
	if _, ok := data[string(key)]; ok {
		return nil, fmt.Errorf("cannot create non-membership proof if key is in map")
	}

	keys := SortedKeys(data)
	rightidx := sort.SearchStrings(keys, string(key))

	var err error
	nonexist := &ics23.NonExistenceProof{
		Key: key,
	}

	// include left proof unless key is left of entire map
	if rightidx >= 1 {
		leftkey := keys[rightidx-1]
		nonexist.Left, err = createExistenceProof(data, []byte(leftkey))
		if err != nil {
			return nil, err
		}
	}

	// include right proof unless key is right of entire map
	if rightidx < len(keys) {
		rightkey := keys[rightidx]
		nonexist.Right, err = createExistenceProof(data, []byte(rightkey))
		if err != nil {
			return nil, err
		}

	}

	proof := &ics23.CommitmentProof{
		Proof: &ics23.CommitmentProof_Nonexist{
			Nonexist: nonexist,
		},
	}
	return proof, nil
}

func createExistenceProof(data map[string][]byte, key []byte) (*ics23.ExistenceProof, error) {
	for k := range data {
		if k == "" {
			return nil, ErrEmptyKeyInData
		}
	}
	value, ok := data[string(key)]
	if !ok {
		return nil, fmt.Errorf("cannot make existence proof if key is not in map")
	}

	_, ics23, _ := sdkmaps.ProofsFromMap(data)
	proof := ics23[string(key)]
	if proof == nil {
		return nil, fmt.Errorf("returned no proof for key")
	}

	return ConvertExistenceProof(proof, key, value)
}
