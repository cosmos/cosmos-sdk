package exported

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

// ProtocolVersionSetter defines the interface fulfilled by BaseApp
// which allows setting it's appVersion field.
type ConsensusParamSetter interface {
	StoreConsensusParams(ctx sdk.Context, cp *tmproto.ConsensusParams)
	GetConsensusParams(ctx sdk.Context) *tmproto.ConsensusParams
}
