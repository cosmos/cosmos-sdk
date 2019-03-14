package proposal

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/errors"
)

// Partially implements sdk.Msg
func ValidateMsgBasic(title, description string, proposer sdk.AccAddress, initialDeposit sdk.Coins) sdk.Error {
	err := IsValidAbstract(errors.DefaultCodespace, title, description)
	if err != nil {
		return err
	}
	if proposer.Empty() {
		return sdk.ErrInvalidAddress(proposer.String())
	}
	if !initialDeposit.IsValid() {
		return sdk.ErrInvalidCoins(initialDeposit.String())
	}
	if initialDeposit.IsAnyNegative() {
		return sdk.ErrInvalidCoins(initialDeposit.String())
	}
	return nil
}
