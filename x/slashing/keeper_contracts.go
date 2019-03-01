package slashing

import sdk "github.com/cosmos/cosmos-sdk/types"

// BankKeeper defines the bank keeper interfact contract the slashing module
// requires. It is needed in order to determine if transfers are enabled.
type BankKeeper interface {
	GetSendEnabled(ctx sdk.Context) bool
}
