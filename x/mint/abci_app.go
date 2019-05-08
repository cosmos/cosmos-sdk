package mint

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

// BeginBlocker inflates every block, update inflation parameters once per hour
func BeginBlocker(ctx sdk.Context, k Keeper) {

	// fetch stored minter & params
	minter := k.GetMinter(ctx)
	params := k.GetParams(ctx)

	// recalculate inflation rate
	stakingSupply := k.sk.StakingTokenSupply(ctx)
	bondedRatio := k.sk.BondedRatio(ctx)
	minter.Inflation = minter.NextInflationRate(params, bondedRatio)
	minter.AnnualProvisions = minter.NextAnnualProvisions(params, stakingSupply)
	k.SetMinter(ctx, minter)

	// mint coins, add to collected fees, update supply
	mintedCoins := sdk.NewCoins(minter.BlockProvision(params))

	// we mint the coins twice to send them to fee collection and staking pool accounts
	k.skk.MintCoins(ctx, ModuleName, mintedCoins.Add(mintedCoins))

	// the fee collector is represented as a base account
	err := k.skk.SendCoinsPoolToAccount(ctx, ModuleName, auth.FeeCollectorAddr, mintedCoins)
	if err != nil {
		panic(err)
	}

	// TODO: get the name from staking module
	err = k.skk.SendCoinsPoolToPool(ctx, ModuleName, "UnbondedTokenSupply", mintedCoins)
	if err != nil {
		panic(err)
	}

	// inflate the total supply tracker
	k.supplyKeeper.Inflate(ctx, mintedCoins)
}
