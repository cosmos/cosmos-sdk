package client

import "github.com/cosmos/cosmos-sdk/types"

// AccountRetriever defines the interfaces required by transactions to
// ensure an account exists and to be able to query for account fields necessary
// for signing.
type AccountRetriever interface {
	EnsureExists(clientCtx Context, addr types.AccAddress) error
	GetAccountNumberSequence(clientCtx Context, addr types.AccAddress) (accNum uint64, accSeq uint64, err error)
}
