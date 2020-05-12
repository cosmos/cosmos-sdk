package types

import (
	"bytes"
	"errors"

	commitmentexported "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/exported"
)

// Mocked types
// TODO: fix tests and replace for real proofs

var (
	_ commitmentexported.Proof = ValidProof{nil, nil, nil}
	_ commitmentexported.Proof = InvalidProof{}
)

type (
	ValidProof struct {
		root  commitmentexported.Root
		path  commitmentexported.Path
		value []byte
	}
	InvalidProof struct{}
)

func (ValidProof) GetCommitmentType() commitmentexported.Type {
	return commitmentexported.Merkle
}

func (proof ValidProof) VerifyMembership(
	root commitmentexported.Root, path commitmentexported.Path, value []byte,
) error {
	if bytes.Equal(root.GetHash(), proof.root.GetHash()) &&
		path.String() == proof.path.String() &&
		bytes.Equal(value, proof.value) {
		return nil
	}

	return errors.New("invalid proof")
}

func (ValidProof) VerifyNonMembership(root commitmentexported.Root, path commitmentexported.Path) error {
	return nil
}

func (ValidProof) ValidateBasic() error {
	return nil
}

func (ValidProof) IsEmpty() bool {
	return false
}

func (InvalidProof) GetCommitmentType() commitmentexported.Type {
	return commitmentexported.Merkle
}

func (InvalidProof) VerifyMembership(
	root commitmentexported.Root, path commitmentexported.Path, value []byte) error {
	return errors.New("proof failed")
}

func (InvalidProof) VerifyNonMembership(root commitmentexported.Root, path commitmentexported.Path) error {
	return errors.New("proof failed")
}

func (InvalidProof) ValidateBasic() error {
	return errors.New("invalid proof")
}

func (InvalidProof) IsEmpty() bool {
	return true
}
