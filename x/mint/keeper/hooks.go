package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/core/event"
	epochstypes "cosmossdk.io/x/epochs/types"
	"cosmossdk.io/x/mint/types"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BeforeEpochStart is a hook which is executed before the start of an epoch. It is a no-op for mint module.
func (k Keeper) BeforeEpochStart(ctx context.Context, epochIdentifier string, epochNumber int64) error {
	// no-op
	return nil
}

// AfterEpochEnd is a hook which is executed after the end of an epoch.
// This hook should attempt to mint and distribute coins according to
// the configuration set via parameters. In addition, it handles the logic
// for reducing minted coins according to the parameters.
// For an attempt to mint to occur:
// - given epochIdentifier must be equal to the mint epoch identifier set via parameters.
// - given epochNumber must be greater than or equal to the mint start epoch set via parameters.
func (k Keeper) AfterEpochEnd(ctx context.Context, epochIdentifier string, epochNumber int64) error {
	defer telemetry.ModuleMeasureSince(types.ModuleName, telemetry.Now(), telemetry.MetricKeyBeginBlocker)

	// fetch stored params
	params, err := k.Params.Get(ctx)
	if err != nil {
		return err
	}

	if epochIdentifier == params.EpochIdentifier {
		// not distribute rewards if it's not time yet for rewards distribution
		if epochNumber < params.MintingRewardsDistributionStartEpoch {
			return nil
		} else if epochNumber == params.MintingRewardsDistributionStartEpoch {
			err := k.setLastReductionEpochNum(ctx, epochNumber)
			if err != nil {
				return err
			}
		}
		// fetch stored minter
		minter, err := k.Minter.Get(ctx)
		if err != nil {
			return err
		}

		// Check if we have hit an epoch where we update the inflation parameter.
		// We measure time between reductions in number of epochs.
		// This avoids issues with measuring in block numbers, as epochs have fixed intervals, with very
		// low variance at the relevant sizes. As a result, it is safe to store the epoch number
		// of the last reduction to be later retrieved for comparison.
		lastReductionEpochNum, err := k.getLastReductionEpochNum(ctx)
		if err != nil {
			return err
		}

		totalStakingSupply, err := k.StakingTokenSupply(ctx)
		if err != nil {
			return err
		}

		// bondedRatio, err := k.BondedRatio(ctx)
		// if err != nil {
		// 	return err
		// }

		if epochNumber >= params.ReductionPeriodInEpochs+lastReductionEpochNum {
			// update minter's inflation and annual provisions
			// minter.Inflation = ic(ctx, minter, params, bondedRatio)

			// Reduce the reward per reduction period: TODO REMOVE COMMENT
			minter.EpochProvisions = minter.NextEpochProvisions(params, totalStakingSupply)
			err := k.Minter.Set(ctx, minter)
			if err != nil {
				return err
			}
			err = k.setLastReductionEpochNum(ctx, epochNumber)
			if err != nil {
				return err
			}

			// calculate minted coins
			// mintedCoin := minter.BlockProvision(params)
			mintedCoin := minter.EpochProvision(params)
			mintedCoins := sdk.NewCoins(mintedCoin)

			maxSupply := params.MaxSupply
			totalSupply := k.bankKeeper.GetSupply(ctx, params.MintDenom).Amount // fetch total supply from the bank module

			// if maxSupply is not infinite, check against max_supply parameter
			if !maxSupply.IsZero() {
				if totalSupply.Add(mintedCoins.AmountOf(params.MintDenom)).GT(maxSupply) {
					// calculate the difference between maxSupply and totalSupply
					diff := maxSupply.Sub(totalSupply)
					// mint the difference
					diffCoin := sdk.NewCoin(params.MintDenom, diff)
					diffCoins := sdk.NewCoins(diffCoin)

					// mint coins
					if err := k.MintCoins(ctx, diffCoins); err != nil {
						return err
					}
					mintedCoins = diffCoins
				}
			}

			// mint coins if maxSupply is infinite or total staking supply is less than maxSupply
			if maxSupply.IsZero() || totalSupply.Add(mintedCoins.AmountOf(params.MintDenom)).LT(maxSupply) {
				// mint coins
				if err := k.MintCoins(ctx, mintedCoins); err != nil {
					return err
				}
			}

			// send the minted coins to the fee collector account
			err = k.AddCollectedFees(ctx, mintedCoins)
			if err != nil {
				return err
			}

			if mintedCoin.Amount.IsInt64() {
				defer telemetry.ModuleSetGauge(types.ModuleName, float32(mintedCoin.Amount.Int64()), "minted_tokens")
			}

			if err := k.EventService.EventManager(ctx).EmitKV(
				types.EventTypeMint,
				event.NewAttribute(types.AttributeEpochNumber, fmt.Sprintf("%d", epochNumber)),
				event.NewAttribute(types.AttributeKeyEpochProvisions, minter.EpochProvisions.String()),
				event.NewAttribute(sdk.AttributeKeyAmount, mintedCoin.Amount.String()),
			); err != nil {
				return err
			}
		}
	}
	return nil
}

// ___________________________________________________________________________________________________

// Hooks wrapper struct for mint keeper.
type Hooks struct {
	k Keeper
}

var _ epochstypes.EpochHooks = Hooks{}

// Return the wrapper struct.
func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

// GetModuleName implements types.EpochHooks.
func (Hooks) GetModuleName() string {
	return types.ModuleName
}

// epochs hooks.
func (h Hooks) BeforeEpochStart(ctx context.Context, epochIdentifier string, epochNumber int64) error {
	return h.k.BeforeEpochStart(ctx, epochIdentifier, epochNumber)
}

func (h Hooks) AfterEpochEnd(ctx context.Context, epochIdentifier string, epochNumber int64) error {
	return h.k.AfterEpochEnd(ctx, epochIdentifier, epochNumber)
}
