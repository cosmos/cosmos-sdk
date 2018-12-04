package auth

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

/*

Usage:

var accounts types.AccountKeeper

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

// add the signers to the context
func WithSigners(ctx sdk.Context, accounts []types.Account) sdk.Context {
	return ctx.WithValue(contextKeySigners, accounts)
}

// get the signers from the context
func GetSigners(ctx sdk.Context) []types.Account {
	v := ctx.Value(contextKeySigners)
	if v == nil {
		return []types.Account{}
	}
	return v.([]types.Account)
}
