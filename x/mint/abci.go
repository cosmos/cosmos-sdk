package mint

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mint/internal/types"
)

func disableMining(minter *types.Minter) {
	minter.Inflation = sdk.ZeroDec()
}

var setInflationHandler func(minter *types.Minter)

// BeginBlocker mints new tokens for the previous block.
func beginBlocker(ctx sdk.Context, k Keeper) {

	logger := ctx.Logger().With("module", "mint")
	// fetch stored minter & params
	params := k.GetParams(ctx)
	minter := k.GetMinterCustom(ctx)
	if ctx.BlockHeight() == 0 || uint64(ctx.BlockHeight()) > minter.NextBlockToUpdate {
		k.UpdateMinterCustom(ctx, &minter, params)
	}

	logger.Debug(fmt.Sprintf(
		"total supply <%v>, "+
			"annual provisions <%v>, "+
			"params <%v>, "+
			"minted this block <%v>, "+
			"next block to update minted per block <%v>, ",
		sdk.NewDecCoinFromDec(params.MintDenom, k.StakingTokenSupply(ctx)),
		sdk.NewDecCoinFromDec(params.MintDenom, minter.AnnualProvisions),
		params,
		minter.MintedPerBlock,
		minter.NextBlockToUpdate))

	err := k.MintCoins(ctx, minter.MintedPerBlock)
	if err != nil {
		panic(err)
	}

	// send the minted coins to the fee collector account
	err = k.AddCollectedFees(ctx, minter.MintedPerBlock)
	if err != nil {
		panic(err)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeMint,
			sdk.NewAttribute(types.AttributeKeyInflation, params.InflationRate.String()),
			sdk.NewAttribute(types.AttributeKeyAnnualProvisions, minter.AnnualProvisions.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, minter.MintedPerBlock.String()),
		),
	)
}

// BeginBlocker mints new tokens for the previous block.
func BeginBlocker(ctx sdk.Context, k Keeper) {
	setInflationHandler = disableMining
	beginBlocker(ctx, k)
}
