package mint

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mint/keeper"
	"github.com/cosmos/cosmos-sdk/x/supply"
)

// BeginBlocker inflates every block and updates inflation parameters once per hour
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {

	// fetch stored minter & params
	minter := k.GetMinter(ctx)
	params := k.GetParams(ctx)

	// recalculate inflation rate
	bondedRatio := k.sk.BondedRatio(ctx)
	minter.Inflation = minter.NextInflationRate(params, bondedRatio)
	minter.AnnualProvisions = minter.NextAnnualProvisions(params, k.sk.StakingTokenSupply(ctx))
	k.SetMinter(ctx, minter)

	// mint coins, add to collected fees, update supply by adding it to the fee collector
	mintedCoin := minter.BlockProvision(params)
	k.fck.AddCollectedFees(ctx, sdk.NewCoins(mintedCoin))

	// // passively keep track of the total and the not bonded supply
	k.supplyKeeper.InflateSupply(ctx, supply.TypeTotal, sdk.NewCoins(mintedCoin))
	k.sk.InflateNotBondedTokenSupply(ctx, mintedCoin.Amount)
}
