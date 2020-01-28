package types

import (
	"errors"

	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

// Mocked types
// TODO: fix tests and replace for real proofs

var (
	_ commitment.ProofI = ValidProof{}
	_ commitment.ProofI = InvalidProof{}
)

type (
	ValidProof   struct{}
	InvalidProof struct{}
)

func (ValidProof) GetCommitmentType() commitment.Type {
	return commitment.Merkle
}

func (ValidProof) VerifyMembership(
	root commitment.RootI, path commitment.PathI, value []byte) error {
	return nil
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
