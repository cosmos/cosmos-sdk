package auth

import (
	sdk "github.com/cosmos/cosmos-sdk"
)

/*

	Usage:

	import "accounts"

	var acc accounts.Account

	accounts.SetAccount(ctx, acc)
	acc2, ok := accounts.GetAccount(ctx)

*/

type contextKey int // local to the auth module

const (
	// A context key of the Account variety
	contextKeyAccount contextKey = iota
)

func SetAccount(ctx sdk.Context, account Account) sdk.Context {
	return ctx.WithValue(contextKeyAccount, account)
}

func GetAccount(ctx sdk.Context) (Account, bool) {
	return ctx.Value(contextKeyAccount).(Account)
}
