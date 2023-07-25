package mint

import (
	"context"
	"time"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mint/keeper"
	"github.com/cosmos/cosmos-sdk/x/mint/types"
)

// BeginBlocker mints new tokens for the previous block.
func BeginBlocker(ctx context.Context, k keeper.Keeper, ic types.InflationCalculationFn) error {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

	logger := k.Logger(ctx)

	// fetch stored minter & params
	minter, err := k.Minter.Get(ctx)
	if err != nil {
		return err
	}

	params, err := k.Params.Get(ctx)
	if err != nil {
		return err
	}

	// recalculate inflation rate
	totalStakingSupply, err := k.StakingTokenSupply(ctx)
	if err != nil {
		return err
	}

	bondedRatio, err := k.BondedRatio(ctx)
	if err != nil {
		return err
	}

	minter.BlockHeader = ctx.BlockHeader()
	minter.Inflation = ic(ctx, minter, params, bondedRatio)
	minter.AnnualProvisions = minter.NextAnnualProvisions(params, totalStakingSupply)
	if err = k.Minter.Set(ctx, minter); err != nil {
		return err
	}

	// mint coins, update supply
	mintedCoin := minter.BlockProvision(params)
	mintedCoins := sdk.NewCoins(mintedCoin)

	divider, _ := math.LegacyNewDecFromStr("0.9")
	mintedCoinsDec := mintedCoins.AmountOf("ujmes")
	unlockedVesting := mintedCoinsDec.Quo(math.Int(divider)).Sub(mintedCoinsDec)
	totalAmount := mintedCoins.AmountOf("ujmes").Add(unlockedVesting)
	foreverVestingAccounts := k.GetAuthKeeper().GetAllForeverVestingAccounts(ctx)
	percentageVestingOfSupply := math.LegacyNewDec(0)

	totalVestedAmount := math.LegacyNewDec(0)
	for _, account := range foreverVestingAccounts {
		vestingSupplyPercentage, _ := math.LegacyNewDecFromStr(account.VestingSupplyPercentage)
		vestedForBlock := sdk.NewCoin("ujmes", totalAmount.ToDec().Mul(vestingSupplyPercentage).TruncateInt())

		percentageVestingOfSupply = percentageVestingOfSupply.Add(vestingSupplyPercentage)
		account.AlreadyVested = account.AlreadyVested.Add(vestedForBlock)
		totalVestedAmount = totalVestedAmount.Add(account.AlreadyVested.AmountOf("ujmes").ToDec())
		k.GetAuthKeeper().SetAccount(ctx, &account)
	}

	logger.Info("=============== =============== ===============")
	logger.Info("=============== MINTER.BeginBlocker(%v)", "height", ctx.BlockHeader().Height)
	logger.Info("=============== MINTER.coinbaseReward(%v)", "reward", mintedCoin)
	logger.Info("=============== =============== ===============")

	currentSupply := k.GetSupply(ctx, "ujmes").Amount
	expectedNextSupply := math.NewInt(currentSupply.Int64()).Add(mintedCoins.AmountOf("ujmes")).Uint64()

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

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeMint,
			sdk.NewAttribute(types.AttributeKeyBondedRatio, bondedRatio.String()),
			sdk.NewAttribute(types.AttributeKeyInflation, minter.Inflation.String()),
			sdk.NewAttribute(types.AttributeKeyAnnualProvisions, minter.AnnualProvisions.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, mintedCoin.Amount.String()),
		),
	)

	return nil
}
