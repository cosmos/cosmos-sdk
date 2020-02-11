package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
)

// AccountKeeper defines the account contract that must be fulfilled when
// creating a x/bank keeper.
type AccountKeeper interface {
	NewAccountWithAddress(ctx sdk.Context, addr sdk.AccAddress) exported.AccountI

	GetAccount(ctx sdk.Context, addr sdk.AccAddress) exported.AccountI
	GetAllAccounts(ctx sdk.Context) []exported.AccountI
	SetAccount(ctx sdk.Context, acc exported.AccountI)

	IterateAccounts(ctx sdk.Context, process func(exported.AccountI) bool)
}
