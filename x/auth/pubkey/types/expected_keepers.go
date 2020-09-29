package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BankKeeper defines the expected interface contract the bank module requires
type BankKeeper interface {
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
}

// StakingKeeper defines the expected interface contract the pubkey module requires
type StakingKeeper interface {
	BondDenom(ctx sdk.Context) string
}
