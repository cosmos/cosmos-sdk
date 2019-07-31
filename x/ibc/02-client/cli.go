package client

import (
	"github.com/cosmos/cosmos-sdk/client/context"

	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/merkle"
)

type CLIObject struct {
	obj Object
}

func (obj Object) CLIObject() CLIObject {
	return CLIObject{obj}
}

// (path, )

func (obj CLIObject) ConsensusState(ctx context.CLIContext) (res ConsensusState, proof merkle.Proof, err error) {
	tmproof, err := obj.obj.ConsensusState.Query(ctx, &res)
	proof = merkle.NewProofFromValue(tmproof, obj.obj.ConsensusState)
	return
}

func (obj CLIObject) Frozen(ctx context.CLIContext) (res bool, proof merkle.Proof, err error) {
	res, tmproof, err := obj.obj.Frozen.Query(ctx)
	proof = merkle.NewProofFromValue(tmproof, obj.obj.Frozen)
	return
}

