package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TODO: AccountKeeper defines the expected account keeper
type AccountKeeper interface {
	NewAccountWithAddress(ctx sdk.Context, addr sdk.AccAddress) Account

	GetAccount(ctx sdk.Context, addr sdk.AccAddress) Account
	GetAllAccounts(ctx sdk.Context) []Account
	SetAccount(ctx sdk.Context, acc Account)

	IterateAccounts(ctx sdk.Context, process func(Account) bool)
}

// TODO: ...
type Account interface {
	GetAddress() sdk.AccAddress
	// SetAddress(sdk.AccAddress) error // errors if already set.
	//
	// GetPubKey() crypto.PubKey // can return nil.
	// SetPubKey(crypto.PubKey) error
	//
	// GetAccountNumber() uint64
	// SetAccountNumber(uint64) error
	//
	// GetSequence() uint64
	// SetSequence(uint64) error

	GetCoins() sdk.Coins
	SetCoins(sdk.Coins) error

	// Calculates the amount of coins that can be sent to other accounts given
	// the current time.
	SpendableCoins(blockTime time.Time) sdk.Coins

	// // Ensure that account implements stringer
	// String() string
}

type VestingAccount interface {
	// Delegation and undelegation accounting that returns the resulting base
	// coins amount.
	TrackDelegation(blockTime time.Time, amount sdk.Coins)
	TrackUndelegation(amount sdk.Coins)
}
