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

	isDevEnv := ctx.ChainID() == "jmes-888"
	faucetAccAddress, _ := sdk.AccAddressFromBech32("jmes137q2vzcu267afzr45crw9x0e97ykrx5t98z7n0")
	if isDevEnv {
		faucetBalance := sdk.NewDecCoinFromCoin(k.BankKeeper.GetBalance(ctx, faucetAccAddress, "ujmes"))
		minFaucetBalance := sdk.NewDecCoin("ujmes", sdk.NewInt(100000000))
		logger.Info("Faucet balance", "amount", faucetBalance)
		if faucetBalance.IsLT(minFaucetBalance) {
			mintingAmount := minFaucetBalance.Sub(faucetBalance)
			logger.Info("Faucet balance is than minFaucetBalance", "faucet balance", faucetBalance, "min faucet:", minFaucetBalance, ". To be minted:", mintingAmount, "to", faucetAccAddress.String())
			k.BankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin("ujmes", sdk.NewInt(mintingAmount.Amount.RoundInt64()))))
			if err := k.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, faucetAccAddress, sdk.NewCoins(sdk.NewCoin("ujmes", sdk.NewInt(mintingAmount.Amount.RoundInt64())))); err != nil {
				panic(err)
			}
		}
	}

	currentSupply := k.BankKeeper.GetSupply(ctx, "ujmes").Amount
	expectedNextSupply := int(sdk.NewInt(currentSupply.Int64()).Add(mintedCoins.AmountOf("ujmes")).Int64())
	maxSupply := int(params.MaxMintableAmount) * 1e6
	logger.Info("Prepare to mint", "mintAmount", mintedCoins.AmountOf("ujmes").String(), ".currentSupply", currentSupply, "expectedNextSupply", expectedNextSupply, "maxSupply", maxSupply)
	if expectedNextSupply <= maxSupply {
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
