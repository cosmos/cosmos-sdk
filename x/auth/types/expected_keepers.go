package types

import (
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type AccountI interface {
	GetAddress(sdk.Context) sdk.AccAddress
	SetAddress(sdk.Context, sdk.AccAddress) error // errors if already set.

	GetPubKey(sdk.Context) cryptotypes.PubKey // can return nil.
	SetPubKey(sdk.Context, cryptotypes.PubKey) error

	GetAccountNumber(sdk.Context) uint64
	SetAccountNumber(sdk.Context, uint64) error

	GetSequence(sdk.Context) uint64
	SetSequence(sdk.Context, uint64) error

	// Ensure that account implements stringer
	String() string
}

// BankKeeper defines the contract needed for supply related APIs (noalias)
type BankKeeper interface {
	SendCoins(ctx sdk.Context, from, to sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
}
