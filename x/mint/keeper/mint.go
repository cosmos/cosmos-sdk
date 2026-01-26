package keeper

import (
	sdkmath "cosmossdk.io/math"

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

		// update minter's inflation and annual provisions
		minter.Inflation = ic(ctx, minter, params, bondedRatio)
		minter.AnnualProvisions = minter.NextAnnualProvisions(params, totalStakingSupply)
		if err = k.Minter.Set(ctx, minter); err != nil {
			return err
		}

		// calculate minted coins
		mintedCoin := minter.BlockProvision(params)

		maxSupply := params.MaxSupply
		totalSupply := k.bankKeeper.GetSupply(ctx, params.MintDenom).Amount // fetch total supply from the bank module

		// if maxSupply is not infinite, and minted coins exceeds maxSupply, adjust minted coins to be the diff
		if !maxSupply.IsZero() && totalSupply.Add(mintedCoin.Amount).GT(maxSupply) {
			// calculate the difference between maxSupply and totalSupply
			diff := maxSupply.Sub(totalSupply)
			if diff.Sign() == -1 {
				// mint nothing if total supply already exceeds max supply
				diff = sdkmath.ZeroInt()
			}
			// mint the difference
			mintedCoin.Amount = diff
		}

		// mint coins, will skip zero coin automatically
		mintedCoins := sdk.NewCoins(mintedCoin)
		if err := k.MintCoins(ctx, mintedCoins); err != nil {
			return err
		}

		// send the minted coins to the fee collector account
		err = k.AddCollectedFees(ctx, mintedCoins)
		if err != nil {
			return err
		}

		if mintedCoin.Amount.IsInt64() {
			defer telemetry.ModuleSetGauge(types.ModuleName, float32(mintedCoin.Amount.Int64()), "minted_tokens") //nolint:staticcheck // TODO: switch to OpenTelemetry
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
