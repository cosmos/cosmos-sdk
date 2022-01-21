package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// TokensToConsensusPower - convert input tokens to potential consensus-engine power
func (k Keeper) TokensToConsensusPower(ctx sdk.Context, tokens sdk.Int) int64 {
	return types.TokensToConsensusPower(tokens, k.PowerReduction(ctx))
}

// TokensFromConsensusPower - convert input power to tokens
func (k Keeper) TokensFromConsensusPower(ctx sdk.Context, power int64) sdk.Int {
	return types.TokensFromConsensusPower(power, k.PowerReduction(ctx))
}
