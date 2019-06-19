package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// AccountKeeper defines the expected account keeper
type AccountKeeper interface {
	NewAccountWithAddress(ctx sdk.Context, addr sdk.AccAddress) authtypes.Account

	GetAccount(ctx sdk.Context, addr sdk.AccAddress) authtypes.Account
	GetAllAccounts(ctx sdk.Context) []authtypes.Account
	SetAccount(ctx sdk.Context, acc authtypes.Account)

	IterateAccounts(ctx sdk.Context, process func(authtypes.Account) bool)
}
