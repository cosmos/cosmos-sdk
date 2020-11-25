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
	if ctx.BlockHeight() == 0 || uint64(ctx.BlockHeight()) >= minter.NextBlockToUpdate {
		k.UpdateMinterCustom(ctx, &minter, params)
	}

	if minter.MintedPerBlock.AmountOf(params.MintDenom).LTE(sdk.ZeroDec()) {
		logger.Debug(fmt.Sprintf("No more <%v> to mint", params.MintDenom))
		return
	}

	err := k.MintCoins(ctx, minter.MintedPerBlock)
	if err != nil {
		panic(err)
	}

	farmingAmount := minter.MintedPerBlock.MulDecTruncate(params.FarmProportion)
	// send the minted coins to the fee collector account
	err = k.AddCollectedFees(ctx, minter.MintedPerBlock.Sub(farmingAmount))
	if err != nil {
		panic(err)
	}
	// send the minted coins to the farm module account
	err = k.AddYieldFarming(ctx, farmingAmount)
	if err != nil {
		panic(err)
	}

	logger.Debug(fmt.Sprintf(
		"total supply <%v>, "+
			"\nparams <%v>, "+
			"\nminted amount<%v>, "+
			"staking amount <%v>, "+
			"yield farming amount <%v>, "+
			"\nnext block to update minted per block <%v>, ",
		sdk.NewDecCoinFromDec(params.MintDenom, k.StakingTokenSupply(ctx)),
		params,
		minter.MintedPerBlock,
		minter.MintedPerBlock.Sub(farmingAmount),
		farmingAmount,
		minter.NextBlockToUpdate))

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeMint,
			sdk.NewAttribute(types.AttributeKeyInflation, params.DeflationRate.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, minter.MintedPerBlock.String()),
		),
	)
}

// BeginBlocker mints new tokens for the previous block.
func BeginBlocker(ctx sdk.Context, k Keeper) {
	setInflationHandler = disableMining
	beginBlocker(ctx, k)
}

