package types

import (
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/exported"
)

var _ exported.Proof = SignatureProof{}

// SignatureProof is a signature used as proof for verification.
type SignatureProof struct {
	Signature []byte
}

// GetCommitmentType implements ProofI.
func (SignatureProof) GetCommitmentType() exported.Type {
	return exported.Signature
}

// VerifyMembership implements ProofI.
func (SignatureProof) VerifyMembership(exported.Root, exported.Path, []byte) error {
	return nil
}

// VerifyNonMembership implements ProofI.
func (SignatureProof) VerifyNonMembership(exported.Root, exported.Path) error {
	return nil
}

// IsEmpty returns trie if the signature is emtpy.
func (proof SignatureProof) IsEmpty() bool {
	return len(proof.Signature) == 0
}

// ValidateBasic checks if the proof is empty.
func (proof SignatureProof) ValidateBasic() error {
	if proof.IsEmpty() {
		return ErrInvalidProof
	}
	return nil
}
