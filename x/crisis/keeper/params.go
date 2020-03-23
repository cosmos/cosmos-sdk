package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/crisis/types"
)

// GetConstantFee get's the constant fee from the paramSpace
func (k Keeper) GetConstantFee(ctx sdk.Context) (constantFee sdk.Coin) {
	k.paramSpace.Get(ctx, types.ParamStoreKeyConstantFee, &constantFee)
	return
}

// GetConstantFee set's the constant fee in the paramSpace
func (k Keeper) SetConstantFee(ctx sdk.Context, constantFee sdk.Coin) {
	k.paramSpace.Set(ctx, types.ParamStoreKeyConstantFee, constantFee)
}
