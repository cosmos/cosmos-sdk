package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
)

// type declaration for parameters
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable(
		ParamStoreKeyCommunityTax, sdk.Dec{},
		ParamStoreKeyProposerReward, sdk.Dec{},
		ParamStoreKeySignerReward, sdk.Dec{},
		ParamStoreKeyWithdrawAddrEnabled, false,
	)
}

// returns the current CommunityTax rate from the global param store
// nolint: errcheck
func (k Keeper) GetCommunityTax(ctx sdk.Context) sdk.Dec {
	var fraction sdk.Dec
	k.paramSpace.Get(ctx, ParamStoreKeyCommunityTax, &fraction)
	return fraction
}

// nolint: errcheck
func (k Keeper) SetCommunityTax(ctx sdk.Context, fraction sdk.Dec) {
	k.paramSpace.Set(ctx, ParamStoreKeyCommunityTax, &fraction)
}

// returns the current ProposerReward rate from the global param store
// nolint: errcheck
func (k Keeper) GetProposerReward(ctx sdk.Context) sdk.Dec {
	var fraction sdk.Dec
	k.paramSpace.Get(ctx, ParamStoreKeyProposerReward, &fraction)
	return fraction
}

// nolint: errcheck
func (k Keeper) SetProposerReward(ctx sdk.Context, fraction sdk.Dec) {
	k.paramSpace.Set(ctx, ParamStoreKeyProposerReward, &fraction)
}

// returns the current SignerReward rate from the global param store
func (k Keeper) GetSignerReward(ctx sdk.Context) sdk.Dec {
	var fraction sdk.Dec
	k.paramSpace.Get(ctx, ParamStoreKeySignerReward, &fraction)
	return fraction
}

// nolint: errcheck
func (k Keeper) SetSignerReward(ctx sdk.Context, fraction sdk.Dec) {
	k.paramSpace.Set(ctx, ParamStoreKeySignerReward, &fraction)
}

// returns the current WithdrawAddrEnabled
// nolint: errcheck
func (k Keeper) GetWithdrawAddrEnabled(ctx sdk.Context) bool {
	var enabled bool
	k.paramSpace.Get(ctx, ParamStoreKeyWithdrawAddrEnabled, &enabled)
	return enabled
}

// nolint: errcheck
func (k Keeper) SetWithdrawAddrEnabled(ctx sdk.Context, enabled bool) {
	k.paramSpace.Set(ctx, ParamStoreKeyWithdrawAddrEnabled, &enabled)
}
