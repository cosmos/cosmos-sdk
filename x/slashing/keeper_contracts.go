package slashing

import sdk "github.com/cosmos/cosmos-sdk/types"

// BankKeeper defines the bank keeper interfact contract the slashing module
// requires.
type BankKeeper interface {
	GetSendEnabled(ctx sdk.Context) bool
}
