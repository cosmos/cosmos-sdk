package mint

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

// BeginBlocker mints new tokens for the previous block
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

	// mint coins, update supply
	mintedCoins := sdk.NewCoins(minter.BlockProvision(params))

	err := k.supplyKeeper.MintCoins(ctx, ModuleName, mintedCoins.Add(mintedCoins))
	if err != nil {
		panic(err)
	}

	// send the minted coins to the fee collector account
	err = k.supplyKeeper.SendCoinsPoolToAccount(ctx, ModuleName, auth.FeeCollectorAddr, mintedCoins)
	if err != nil {
		panic(err)
	}
}
