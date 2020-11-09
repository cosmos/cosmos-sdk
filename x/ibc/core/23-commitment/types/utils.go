package types

import (
	ics23 "github.com/confio/ics23/go"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	crypto "github.com/tendermint/tendermint/proto/tendermint/crypto"
)

// ConvertProofs converts crypto.ProofOps into MerkleProof
func ConvertProofs(tmproof *crypto.ProofOps) (MerkleProof, error) {
	// Unmarshal all proof ops to CommitmentProof
	proofs := make([]*ics23.CommitmentProof, len(tmproof.Ops))
	for i, op := range tmproof.Ops {
		var p ics23.CommitmentProof
		err := p.Unmarshal(op.Data)
		if err != nil {
			return MerkleProof{}, sdkerrors.Wrapf(ErrInvalidMerkleProof, "could not unmarshal proof op into CommitmentProof at index: %d", i)
		}
		proofs[i] = &p
	}
	return MerkleProof{
		Proofs: proofs,
	}, nil
}
