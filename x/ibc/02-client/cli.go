package client

import (
	"github.com/cosmos/cosmos-sdk/client/context"

	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/merkle"
)

func (obj Object) ConsensusStateCLI(ctx context.CLIContext, path merkle.Path) (res ConsensusState, proof merkle.Proof, err error) {
	tmproof, err := obj.ConsensusState.Query(ctx, &res)
	proof = merkle.NewProofFromValue(tmproof, path, obj.ConsensusState)
	return
}

func (obj Object) FrozenCLI(ctx context.CLIContext, path merkle.Path) (res bool, proof merkle.Proof, err error) {
	res, tmproof, err := obj.Frozen.Query(ctx)
	proof = merkle.NewProofFromValue(tmproof, path, obj.Frozen)
	return
}

