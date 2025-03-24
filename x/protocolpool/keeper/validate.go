package keeper

import (
	"errors"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/protocolpool/types"
)

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
