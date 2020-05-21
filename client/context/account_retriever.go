package context

import "github.com/cosmos/cosmos-sdk/types"

// AccountRetriever defines the interfaces required by transactions to
// ensure an account exists and to be able to query for account fields necessary
// for signing.
type AccountRetriever interface {
	EnsureExists(nodeQuerier NodeQuerier, addr types.AccAddress) error
	GetAccountNumberSequence(nodeQuerier NodeQuerier, addr types.AccAddress) (accNum uint64, accSeq uint64, err error)
}

// NodeQuerier is an interface that is satisfied by types that provide the QueryWithData method
type NodeQuerier interface {
	// QueryWithData performs a query to a Tendermint node with the provided path
	// and a data payload. It returns the result and height of the query upon success
	// or an error if the query fails.
	QueryWithData(path string, data []byte) ([]byte, int64, error)
}

var _ NodeQuerier = CLIContext{}
