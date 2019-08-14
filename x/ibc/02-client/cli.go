package client

import (
	"bytes"

	"github.com/cosmos/cosmos-sdk/client/context"

	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/merkle"
)

func (obj Object) prefix() []byte {
	return bytes.Split(obj.ConsensusState.KeyBytes(), LocalRoot())[0]
}

func (obj Object) ConsensusStateCLI(ctx context.CLIContext) (res ConsensusState, proof merkle.Proof, err error) {
	tmproof, err := obj.ConsensusState.Query(ctx, &res)
	proof = merkle.NewProofFromValue(tmproof, obj.prefix(), obj.ConsensusState)
	return
}

func (obj Object) FrozenCLI(ctx context.CLIContext) (res bool, proof merkle.Proof, err error) {
	res, tmproof, err := obj.Frozen.Query(ctx)
	proof = merkle.NewProofFromValue(tmproof, obj.prefix(), obj.Frozen)
	return
}

