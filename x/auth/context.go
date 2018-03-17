package auth

import (
	"github.com/cosmos/cosmos-sdk/types"
)

/*

Usage:

var accounts types.AccountMapper

// Fetch all signer accounts.
addrs := tx.GetSigners()
signers := make([]types.Account, len(addrs))
for i, addr := range addrs {
	acc := accounts.GetAccount(ctx)
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
	v := ctx.Value(contextKeySigners)
	if v == nil {
		return []types.Account{}
	}
	return v.([]types.Account)
}
