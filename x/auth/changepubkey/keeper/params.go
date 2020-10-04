package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/changepubkey/types"
)

// SetParams sets the changepubkey module's parameters.
func (pk ChangePubKeyKeeper) SetParams(ctx sdk.Context, params types.Params) {
	pk.paramSubspace.SetParamSet(ctx, &params)
}

// GetParams gets the changepubkey module's parameters.
func (pk ChangePubKeyKeeper) GetParams(ctx sdk.Context) (params types.Params) {
	pk.paramSubspace.GetParamSet(ctx, &params)
	return
}
