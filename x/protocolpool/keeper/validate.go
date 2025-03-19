package keeper

import (
	"errors"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/protocolpool/types"
)

// validateAndUpdateBudget validates the Budget included in a MsgCreateBudget as follows:
// - BudgetPerTranche must be nonzero
// - the budget amount must be a valid sdk.Coin
// - the startTime must be valid (after current blocktime)
// - - if the startTime was nil, set it to the current blocktime
// - number of tranches must be nonzero
// - period duration must be nonzero
func validateAndUpdateBudget(ctx sdk.Context, bp types.MsgCreateBudget) (types.Budget, error) {
	if bp.BudgetPerTranche.IsZero() {
		return types.Budget{}, errors.New("invalid budget proposal: budget per tranche cannot be zero")
	}

	if err := validateAmount(sdk.Coins{bp.BudgetPerTranche}); err != nil {
		return types.Budget{}, fmt.Errorf("invalid budget proposal: %w", err)
	}

	currentTime := ctx.BlockTime()
	if bp.StartTime == nil || bp.StartTime.IsZero() {
		bp.StartTime = &currentTime
	}

	if currentTime.After(*bp.StartTime) {
		return types.Budget{}, errors.New("invalid budget proposal: start time cannot be less than the current block time")
	}

	if bp.Tranches == 0 {
		return types.Budget{}, errors.New("invalid budget proposal: tranches must be greater than zero")
	}

	if bp.Period == 0 {
		return types.Budget{}, errors.New("invalid budget proposal: period length should be greater than zero")
	}

	// Create and return an updated budget proposal
	updatedBudget := types.Budget{
		RecipientAddress: bp.RecipientAddress,
		BudgetPerTranche: bp.BudgetPerTranche,
		LastClaimedAt:    *bp.StartTime,
		TranchesLeft:     bp.Tranches,
		Period:           bp.Period,
	}

	return updatedBudget, nil
}

// validateContinuousFund validates the fields of the CreateContinuousFund message.
func validateContinuousFund(ctx sdk.Context, msg types.MsgCreateContinuousFund) error {
	// Validate percentage
	if msg.Percentage.IsZero() || msg.Percentage.IsNil() {
		return errors.New("percentage cannot be zero or empty")
	}
	if msg.Percentage.IsNegative() {
		return errors.New("percentage cannot be negative")
	}
	if msg.Percentage.GTE(math.LegacyOneDec()) {
		return errors.New("percentage cannot be greater than or equal to one")
	}

	// Validate expiry
	currentTime := ctx.BlockTime()
	if msg.Expiry != nil && msg.Expiry.Compare(currentTime) == -1 {
		return errors.New("expiry time cannot be less than the current block time")
	}

	return nil
}

func validateAmount(amount sdk.Coins) error {
	if amount == nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, "amount cannot be nil")
	}

	if err := amount.Validate(); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, amount.String())
	}

	return nil
}
