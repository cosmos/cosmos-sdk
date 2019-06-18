package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

// expected account keeper
type AccountKeeper interface {
	NewAccountWithAddress(ctx sdk.Context, addr sdk.AccAddress) auth.Account

	GetAccount(ctx sdk.Context, addr sdk.AccAddress) auth.Account
	GetAllAccounts(ctx sdk.Context) []auth.Account
	SetAccount(ctx sdk.Context, acc auth.Account)

	IterateAccounts(ctx sdk.Context, process func(auth.Account) bool)
}
