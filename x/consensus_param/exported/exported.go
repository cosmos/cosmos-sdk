package exported

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

// ProtocolVersionSetter defines the interface fulfilled by BaseApp
// which allows setting it's appVersion field.
type ConsensusParamSetter interface {
	Get(ctx sdk.Context, key []byte) (tmproto.ConsensusParams, error)
	Has(ctx sdk.Context, key []byte) bool
	Set(ctx sdk.Context, cp tmproto.ConsensusParams)
}
