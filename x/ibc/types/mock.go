package types

import (
	"bytes"
	"errors"

	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

// Mocked types
// TODO: fix tests and replace for real proofs

var (
	_ commitment.ProofI = ValidProof{nil, nil, nil}
	_ commitment.ProofI = InvalidProof{}
)

type (
	ValidProof struct {
		root  commitment.RootI
		path  commitment.PathI
		value []byte
	}
	InvalidProof struct{}
)

func (ValidProof) GetCommitmentType() commitment.Type {
	return commitment.Merkle
}

func (proof ValidProof) VerifyMembership(
	root commitment.RootI, path commitment.PathI, value []byte) error {
	if bytes.Equal(root.GetHash(), proof.root.GetHash()) &&
		(path.String() == proof.path.String()) &&
		bytes.Equal(value, proof.value) {
		return nil
	} else {
		return errors.New("invalid proof")
	}
}

func (ValidProof) VerifyNonMembership(root commitment.RootI, path commitment.PathI) error {
	return nil
}

func (ValidProof) ValidateBasic() error {
	return nil
}

func (ValidProof) IsEmpty() bool {
	return false
}

func (InvalidProof) GetCommitmentType() commitment.Type {
	return commitment.Merkle
}

func (InvalidProof) VerifyMembership(
	root commitment.RootI, path commitment.PathI, value []byte) error {
	return errors.New("proof failed")
}

func (InvalidProof) VerifyNonMembership(root commitment.RootI, path commitment.PathI) error {
	return errors.New("proof failed")
}

func (InvalidProof) ValidateBasic() error {
	return errors.New("invalid proof")
}

func (InvalidProof) IsEmpty() bool {
	return true
}
