package mint

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mint/internal/types"
)

// BeginBlocker mints new tokens for the previous block.
func BeginBlocker(ctx sdk.Context, k Keeper) {
	// fetch stored minter & params
	minter := k.GetMinter(ctx)
	params := k.GetParams(ctx)

	// recalculate inflation rate
	totalSupply := k.TotalTokens(ctx)
	bondedRatio := k.BondedRatio(ctx)
	minter.Inflation = minter.NextInflationRate(params, bondedRatio)
	minter.AnnualProvisions = minter.NextAnnualProvisions(params, totalSupply)
	k.SetMinter(ctx, minter)

	// mint coins, add to collected fees, update supply
	mintedCoin := minter.BlockProvision(params)
	k.AddCollectedFees(ctx, sdk.Coins{mintedCoin})
	k.InflateSupply(ctx, mintedCoin.Amount)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.Mint,
			sdk.NewAttribute(types.BondedRatio, bondedRatio.String()),
			sdk.NewAttribute(types.Inflation, minter.Inflation.String()),
			sdk.NewAttribute(types.AnnualProvisions, minter.AnnualProvisions.String()),
			sdk.NewAttribute(types.Amount, mintedCoin.Amount.String()),
		),
	})
}
