package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/protocolpool/types"
)

// validateContinuousFund validates the fields of the CreateContinuousFund message.
func validateContinuousFund(ctx sdk.Context, msg types.MsgCreateContinuousFund) error {
	fund := types.ContinuousFund{
		Recipient:  msg.Recipient,
		Percentage: msg.Percentage,
		Expiry:     msg.Expiry,
	}
	if err := fund.Validate(); err != nil {
		return fmt.Errorf("invalid continuous fund: %w", err)
	}

	// Validate expiry
	currentTime := ctx.BlockTime()
	if msg.Expiry != nil && msg.Expiry.Compare(currentTime) == -1 {
		return fmt.Errorf("expiry time %s cannot be less than the current block time %s", msg.Expiry, currentTime)
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
