package keeper

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/protocolpool/types"
)

func (k Keeper) InitGenesis(ctx sdk.Context, data *types.GenesisState) error {
	currentTime := ctx.BlockTime()

	err := k.Params.Set(ctx, *data.Params)
	if err != nil {
		return fmt.Errorf("failed to set params: %w", err)
	}

	for _, cf := range data.ContinuousFunds {
		// ignore expired ContinuousFunds
		if cf.Expiry != nil && cf.Expiry.Before(currentTime) {
			continue
		}

		recipientAddress, err := k.authKeeper.AddressCodec().StringToBytes(cf.Recipient)
		if err != nil {
			return fmt.Errorf("failed to decode recipient address: %w", err)
		}
		if err := k.ContinuousFunds.Set(ctx, recipientAddress, cf); err != nil {
			return fmt.Errorf("failed to set continuous fund for recipient %s: %w", recipientAddress, err)
		}
	}

	// todo:  validate all continuous funds

	// sanity check to avoid trying to distribute more than what is available

	return nil
}

func (k Keeper) ExportGenesis(ctx sdk.Context) (*types.GenesisState, error) {
	// refresh all funds
	if err := k.IterateAndUpdateFundsDistribution(ctx); err != nil {
		return nil, err
	}

	// withdraw all rewards before exporting genesis
	if err := k.RecipientFundDistributions.Walk(ctx, nil, func(key sdk.AccAddress, value types.DistributionAmount) (stop bool, err error) {
		if _, err := k.withdrawRecipientFunds(ctx, key.Bytes()); err != nil {
			return true, err
		}
		return false, nil
	}); err != nil {
		return nil, err
	}

	var cf []*types.ContinuousFund
	err := k.ContinuousFunds.Walk(ctx, nil, func(key sdk.AccAddress, value types.ContinuousFund) (stop bool, err error) {
		recipient, err := k.authKeeper.AddressCodec().BytesToString(key)
		if err != nil {
			return true, err
		}
		cf = append(cf, &types.ContinuousFund{
			Recipient:  recipient,
			Percentage: value.Percentage,
			Expiry:     value.Expiry,
		})
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	var budget []*types.Budget
	err = k.Budgets.Walk(ctx, nil, func(key sdk.AccAddress, value types.Budget) (stop bool, err error) {
		recipient, err := k.authKeeper.AddressCodec().BytesToString(key)
		if err != nil {
			return true, err
		}
		budget = append(budget, &types.Budget{
			RecipientAddress: recipient,
			ClaimedAmount:    value.ClaimedAmount,
			LastClaimedAt:    value.LastClaimedAt,
			TranchesLeft:     value.TranchesLeft,
			Period:           value.Period,
			BudgetPerTranche: value.BudgetPerTranche,
		})
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	genState := types.NewGenesisState(cf, budget)

	lastBalance, err := k.LastBalance.Get(ctx)
	if err != nil {
		return nil, err
	}

	genState.LastBalance = lastBalance

	err = k.Distributions.Walk(ctx, nil, func(key time.Time, value types.DistributionAmount) (stop bool, err error) {
		genState.Distributions = append(genState.Distributions, &types.Distribution{
			Time:   &key,
			Amount: value,
		})

		return false, nil
	})
	if err != nil {
		return nil, err
	}

	params, err := k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}

	genState.Params = &params

	return genState, nil
}
