package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/x/protocolpool/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) InitGenesis(ctx context.Context, data *types.GenesisState) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	currentTime := sdkCtx.BlockTime()
	for _, cf := range data.ContinuousFund {
		// ignore expired ContinuousFunds
		if cf.Expiry != nil && cf.Expiry.Before(currentTime) {
			continue
		}

		recipientAddress, err := k.authKeeper.AddressCodec().StringToBytes(cf.Recipient)
		if err != nil {
			return fmt.Errorf("failed to decode recipient address: %w", err)
		}
		if err := k.ContinuousFund.Set(ctx, recipientAddress, *cf); err != nil {
			return fmt.Errorf("failed to set continuous fund for recipient %s: %w", recipientAddress, err)
		}
	}
	for _, budget := range data.Budget {
		// Validate StartTime
		if budget.StartTime.IsZero() || budget.StartTime == nil {
			budget.StartTime = &currentTime
		}
		// ignore budget with start time < currentTime
		if budget.StartTime != nil && budget.StartTime.Before(currentTime) {
			continue
		}

		recipientAddress, err := k.authKeeper.AddressCodec().StringToBytes(budget.RecipientAddress)
		if err != nil {
			return fmt.Errorf("failed to decode recipient address: %w", err)
		}
		if err = k.BudgetProposal.Set(ctx, recipientAddress, *budget); err != nil {
			return fmt.Errorf("failed to set budget for recipient %s: %w", recipientAddress, err)
		}
	}
	return nil
}

func (k Keeper) ExportGenesis(ctx context.Context) (*types.GenesisState, error) {
	var cf []*types.ContinuousFund
	err := k.ContinuousFund.Walk(ctx, nil, func(key sdk.AccAddress, value types.ContinuousFund) (stop bool, err error) {
		cf = append(cf, &types.ContinuousFund{
			Recipient:  key.String(),
			Percentage: value.Percentage,
			Expiry:     value.Expiry,
		})
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	var budget []*types.Budget
	err = k.BudgetProposal.Walk(ctx, nil, func(key sdk.AccAddress, value types.Budget) (stop bool, err error) {
		budget = append(budget, &types.Budget{
			RecipientAddress: key.String(),
			TotalBudget:      value.TotalBudget,
			ClaimedAmount:    value.ClaimedAmount,
			StartTime:        value.StartTime,
			NextClaimFrom:    value.NextClaimFrom,
			Tranches:         value.Tranches,
			TranchesLeft:     value.TranchesLeft,
			Period:           value.Period,
		})
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	return types.NewGenesisState(cf, budget), nil
}
