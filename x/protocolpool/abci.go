package protocolpool

import (
	"fmt"

	"cosmossdk.io/x/protocolpool/keeper"
	"cosmossdk.io/x/protocolpool/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func EndBlocker(ctx sdk.Context, k keeper.Keeper, authKeeper types.AccountKeeper) error {
	err := k.BudgetProposal.Walk(ctx, nil, func(key sdk.AccAddress, budget types.BudgetProposal) (stop bool, err error) {
		// check if the distribution is completed
		if budget.RemainingTranches <= 0 {
			// Log the end of the budget
			k.Logger(ctx).Info(fmt.Sprintf("Budget ended for recipient: %s", key.String()))
			return false, nil // Continue iterating
		}

		currentTime := ctx.BlockTime().Unix()

		// Check if the start time is reached
		if currentTime < budget.StartTime {
			return false, fmt.Errorf("distribution has not started yet")
		}

		// Calculate the number of blocks elapsed since the start time
		blocksElapsed := ctx.BlockHeight() - budget.Period

		// Check if its time to distribute funds based on period intervals
		if blocksElapsed > 0 && blocksElapsed%budget.Period == 0 {
			// Calculate the amount to distribute in each tranche
			amountPerTranche := budget.TotalBudget.Amount.QuoRaw(budget.RemainingTranches)
			amount := sdk.NewCoin(budget.TotalBudget.Denom, amountPerTranche)

			recipient, err := authKeeper.AddressCodec().StringToBytes(budget.RecipientAddress)
			if err != nil {
				return false, err
			}

			distributionInfo := types.DistributionInfo{
				Address: recipient,
				Amount:  amount,
			}

			// store funds in the DistributionInfo until a claim tx
			err = k.AppendDistributionInfo(ctx, distributionInfo)
			if err != nil {
				return false, err
			}

			// update the budget's remaining tranches
			budget.RemainingTranches--

			// update the TotalBudget amount
			budget.TotalBudget.Amount.Sub(amountPerTranche)

			k.Logger(ctx).Info(fmt.Sprintf("Processing budget for recipient: %s. Amount: %s", budget.RecipientAddress, amountPerTranche.String()))

			// Save the updated budget in the state
			err = k.BudgetProposal.Set(ctx, recipient, budget)
			if err != nil {
				k.Logger(ctx).Error(fmt.Sprintf("Error while updating the budget for recipient %s", budget.RecipientAddress))
				return false, err
			}

		}
		return false, nil
	})

	return err
}
