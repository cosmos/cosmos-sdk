package group

import (
	context "context"

	"cosmossdk.io/core/address"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type AccountKeeper interface {
	address.Codec

	// NewAccount returns a new account with the next account number. Does not save the new account to the store.
	NewAccount(context.Context, sdk.AccountI) (sdk.AccountI,error)

	// GetAccount retrieves an account from the store.
	GetAccount(context.Context, sdk.AccAddress) (sdk.AccountI,error)

	// SetAccount sets an account in the store.
	SetAccount(context.Context, sdk.AccountI)error

	// RemoveAccount Remove an account in the store.
	RemoveAccount(ctx context.Context, acc sdk.AccountI) error
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	SpendableCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins
}
