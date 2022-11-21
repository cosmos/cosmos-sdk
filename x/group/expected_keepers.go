package group

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type AccountKeeper interface {
	// Return a new account with the next account number. Does not save the new account to the store.
	NewAccount(sdk.Context, sdk.AccountI) sdk.AccountI

	// Retrieve an account from the store.
	GetAccount(sdk.Context, sdk.AccAddress) sdk.AccountI

	// Set an account in the store.
	SetAccount(sdk.Context, sdk.AccountI)
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
}
