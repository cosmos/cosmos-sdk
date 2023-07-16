package mint

import (
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mint/keeper"
	"github.com/cosmos/cosmos-sdk/x/mint/types"
)

// BeginBlocker mints new tokens for the previous block.
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)
	logger := k.Logger(ctx)

	// fetch stored minter & params
	minter := k.GetMinter(ctx)
	params := k.GetParams(ctx)

	// recalculate inflation rate
	totalStakingSupply := k.StakingTokenSupply(ctx)
	bondedRatio := k.BondedRatio(ctx)
	minter.BlockHeader = ctx.BlockHeader()
	minter.Inflation = minter.NextInflationRate(params, bondedRatio)
	minter.AnnualProvisions = minter.NextAnnualProvisions(params, totalStakingSupply)
	k.SetMinter(ctx, minter)

	// mint coins, update supply
	mintedCoin := minter.BlockProvision(params)
	mintedCoins := sdk.NewCoins(mintedCoin)

	divider, _ := sdk.NewDecFromStr("0.9")
	mintedCoinsDec := mintedCoins.AmountOf("ujmes").ToDec()
	unlockedVesting := mintedCoinsDec.Quo(divider).RoundInt().Sub(mintedCoinsDec.RoundInt())
	totalAmount := mintedCoins.AmountOf("ujmes").ToDec().Add(unlockedVesting.ToDec()).RoundInt()

	foreverVestingAccounts := k.GetAuthKeeper().GetAllForeverVestingAccounts(ctx)
	percentageVestingOfSupply := sdk.NewDec(0)
	totalVestedAmount := sdk.NewDec(0)
	for _, account := range foreverVestingAccounts {
		vestingSupplyPercentage, _ := sdk.NewDecFromStr(account.VestingSupplyPercentage)
		vestedForBlock := sdk.NewCoin("ujmes", totalAmount.ToDec().Mul(vestingSupplyPercentage).TruncateInt())

		percentageVestingOfSupply = percentageVestingOfSupply.Add(vestingSupplyPercentage)
		account.AlreadyVested = account.AlreadyVested.Add(vestedForBlock)
		totalVestedAmount = totalVestedAmount.Add(account.AlreadyVested.AmountOf("ujmes").ToDec())
		k.GetAuthKeeper().SetAccount(ctx, &account)
	}

	currentSupply := k.GetSupply(ctx, "ujmes").Amount
	expectedNextSupply := sdk.NewInt(currentSupply.Int64()).Add(mintedCoins.AmountOf("ujmes")).Uint64()
	maxMintableAmount := params.GetMaxMintableAmount()

	logger.Info("Prepare to mint", "blockheight", ctx.BlockHeight(), "mintAmount", mintedCoins.AmountOf("ujmes").String(), ".currentSupply", currentSupply, "expectedNextSupply", expectedNextSupply, "maxSupply", maxMintableAmount, "unlockedVesting", unlockedVesting)
	if expectedNextSupply <= maxMintableAmount {
		err := k.MintCoins(ctx, mintedCoins)
		if err != nil {
			panic(err)
		}
		// send the minted coins to the fee collector account
		err = k.AddCollectedFees(ctx, mintedCoins)
		if err != nil {
			panic(err)
		}

	} else {
		logger.Info("Abort minting. ", "total", expectedNextSupply, "would exceed", params.MaxMintableAmount)
	}

	if mintedCoin.Amount.IsInt64() {
		defer telemetry.ModuleSetGauge(types.ModuleName, float32(mintedCoin.Amount.Int64()), "minted_tokens")
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeMint,
			sdk.NewAttribute(types.AttributeKeyBondedRatio, bondedRatio.String()),
			sdk.NewAttribute(types.AttributeKeyInflation, minter.Inflation.String()),
			sdk.NewAttribute(types.AttributeKeyAnnualProvisions, minter.AnnualProvisions.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, mintedCoin.Amount.String()),
		),
	)
}
