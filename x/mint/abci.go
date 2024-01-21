package mint

import (
	"context"
	"time"

	"cosmossdk.io/x/mint/keeper"
	"cosmossdk.io/x/mint/types"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BeginBlocker mints new tokens for the previous block.
func BeginBlocker(ctx context.Context, k keeper.Keeper) error {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// fetch stored minter
	m, err := k.Minter.Get(ctx)
	if err != nil {
		return err
	}

	// corner case during processing the genesis block
	if sdkCtx.HeaderInfo().Height == 1 {
		genesisTime := sdkCtx.HeaderInfo().Time
		m.GenesisTime = &genesisTime
	}

	// recalculate inflation rate
	inflation := types.InflationRate(*m.GenesisTime, sdkCtx.HeaderInfo().Time)

	// no need to modify any params if inflation is already correct
	// zero annual provision is a corner case when processing the genesis block
	if !m.Inflation.Equal(inflation) || m.AnnualProvisions.IsZero() {
		totalStakingSupply, errX := k.StakingTokenSupply(ctx)
		if errX != nil {
			return errX
		}

		m.Inflation = inflation
		m.AnnualProvisions = types.AnnualProvisions(inflation, totalStakingSupply)
	}

	// mint block provision
	err = mintCoins(sdkCtx, m, k)
	if err != nil {
		return err
	}

	// update last block time
	blockTime := sdkCtx.HeaderInfo().Time
	m.PreviousBlockTime = &blockTime

	// save all miter changes
	err = k.Minter.Set(ctx, m)
	if err != nil {
		return err
	}

	return nil
}

func mintCoins(ctx sdk.Context, m types.Minter, k keeper.Keeper) error {
	if m.PreviousBlockTime == nil {
		// this is expected to happen for the genesis block
		return nil
	}

	var (
		blockProvision = types.BlockProvision(ctx.HeaderInfo().Time, *m.PreviousBlockTime, m.AnnualProvisions)
		mintedCoin     = sdk.NewCoin(m.MintDenom, blockProvision)
		mintedCoins    = sdk.NewCoins()
	)

	err := k.MintCoins(ctx, mintedCoins)
	if err != nil {
		return err
	}

	// send the minted coins to the fee collector account
	err = k.AddCollectedFees(ctx, mintedCoins)
	if err != nil {
		return err
	}

	if mintedCoin.Amount.IsInt64() {
		defer telemetry.ModuleSetGauge(types.ModuleName, float32(mintedCoin.Amount.Int64()), "minted_tokens")
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeMint,
		sdk.NewAttribute(types.AttributeKeyInflation, m.Inflation.String()),
		sdk.NewAttribute(types.AttributeKeyAnnualProvisions, m.AnnualProvisions.String()),
		sdk.NewAttribute(types.AttributeKeyPreviousBlockTime, m.PreviousBlockTime.String()),
		sdk.NewAttribute(types.AttributeKeyCurrentBlockTime, ctx.HeaderInfo().Time.String()),
		sdk.NewAttribute(types.AttributeKeyBlockProvision, blockProvision.String()),
	))

	return nil
}
