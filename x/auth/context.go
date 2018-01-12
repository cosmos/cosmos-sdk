package auth

import (
	"github.com/cosmos/cosmos-sdk/types"
)

/*

Usage:

var accountStore types.AccountStore

// Fetch all signer accounts.
addrs := tx.GetSigners()
signers := make([]types.Account, len(addrs))
for i, addr := range addrs {
	acc := accountStore.GetAccount(ctx)
	signers[i] = acc
}
ctx = auth.SetSigners(ctx, signers)

// Get all signer accounts.
signers := auth.GetSigners(ctx)
for i, signer := range signers {
	signer.Address() == tx.GetSigners()[i]
}

*/

type contextKey int // local to the auth module

const (
	contextKeySigners contextKey = iota
)

func WithSigners(ctx types.Context, accounts []types.Account) types.Context {
	return ctx.WithValue(contextKeySigners, accounts)
}

func GetSigners(ctx types.Context) []types.Account {
	return ctx.Value(contextKeySigners).([]types.Account)
}
