package interfaces

import (
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ConsensusParamsGetter is an interface to retrieve consensus parameters for a given context.
type ConsensusParamsGetter interface {
	GetConsensusParams(ctx sdk.Context) tmproto.ConsensusParams
}
