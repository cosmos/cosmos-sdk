package auth

import (
	"github.com/cosmos/cosmos-sdk/types"
)

/*

	Usage:

	import "accounts"

	var acc accounts.Account

	accounts.SetAccount(ctx, acc)
	acc2 := accounts.GetAccount(ctx)

*/

type contextKey int // local to the auth module

const (
	// A context key of the Account variety
	contextKeyAccount contextKey = iota
)

func SetAccount(ctx types.Context, account Account) types.Context {
	return ctx.WithValueSDK(contextKeyAccount, account)
}

func GetAccount(ctx types.Context) Account {
	return ctx.Value(contextKeyAccount).(Account)
}
