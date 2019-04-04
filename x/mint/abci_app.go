package mint

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Inflate every block, update inflation parameters once per hour
func BeginBlocker(ctx sdk.Context, k Keeper) error {

	// fetch stored minter & params
	minter, err := k.GetMinter(ctx)
	if err != nil {
		return err
	}
	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}

	// recalculate inflation rate
	totalSupply := k.sk.TotalTokens(ctx)
	bondedRatio := k.sk.BondedRatio(ctx)
	minter.Inflation = minter.NextInflationRate(params, bondedRatio)
	minter.AnnualProvisions = minter.NextAnnualProvisions(params, totalSupply)
	k.SetMinter(ctx, minter)

	// mint coins, add to collected fees, update supply
	mintedCoin := minter.BlockProvision(params)
	k.fck.AddCollectedFees(ctx, sdk.Coins{mintedCoin})
	k.sk.InflateSupply(ctx, mintedCoin.Amount)
	return nil

}
