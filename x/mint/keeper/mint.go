package keeper

import (
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mint/types"
)

// MintFn defines the function that needs to be implemented in order to customize the minting process.
type MintFn func(ctx sdk.Context, k *Keeper) error

// MintFn runs the mintFn of the keeper.
func (k *Keeper) MintFn(ctx sdk.Context) error {
	return k.mintFn(ctx, k)
}

// DefaultMintFn returns a default mint function.
// The default MintFn has a requirement on staking as it uses bond to calculate inflation.
func DefaultMintFn(ic types.InflationCalculationFn) MintFn {
	return func(ctx sdk.Context, k *Keeper) error {
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

		minter.Inflation = ic(ctx, minter, params, bondedRatio)
		minter.AnnualProvisions = minter.NextAnnualProvisions(params, totalStakingSupply)
		if err = k.Minter.Set(ctx, minter); err != nil {
			return err
		}

		// mint coins, update supply
		mintedCoin := minter.BlockProvision(params)
		mintedCoins := sdk.NewCoins(mintedCoin)

		err = k.MintCoins(ctx, mintedCoins)
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

		ctx.EventManager().EmitEvent(
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
}
