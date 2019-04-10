package proposal

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Partially implements sdk.Msg
func ValidateMsgBasic(proposer sdk.AccAddress, initialDeposit sdk.Coins) sdk.Error {
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
