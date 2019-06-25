package client

import (
	"github.com/cosmos/cosmos-sdk/client/context"

	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/merkle"
)

// CLIObject stores the key for each object fields
type CLIObject struct {
	ID                string
	ConsensusStateKey []byte
	FrozenKey         []byte
}

func (obj Object) CLI() CLIObject {
	return CLIObject{
		ID:                obj.id,
		ConsensusStateKey: obj.consensusState.Key(),
		FrozenKey:         obj.frozen.Key(),
	}
}

func (obj CLIObject) ConsensusState(ctx context.CLIContext, root merkle.Root) (res ConsensusState, proof merkle..Proof) {
	val, proof, _, err := ctx.QueryProof(obj.ConsensusStateKey)
}
