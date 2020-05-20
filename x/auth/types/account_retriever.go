package types

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NodeQuerier is an interface that is satisfied by types that provide the QueryWithData method
type NodeQuerier interface {
	// QueryWithData performs a query to a Tendermint node with the provided path
	// and a data payload. It returns the result and height of the query upon success
	// or an error if the query fails.
	QueryWithData(path string, data []byte) ([]byte, int64, error)
}

// AccountRetriever defines the properties of a type that can be used to
// retrieve accounts.
type AccountRetriever struct {
	codec   codec.Marshaler
	querier NodeQuerier
}

// NewAccountRetriever initialises a new AccountRetriever instance.
func NewAccountRetriever(codec codec.Marshaler, querier NodeQuerier) AccountRetriever {
	return AccountRetriever{codec: codec, querier: querier}
}

// GetAccount queries for an account given an address and a block height. An
// error is returned if the query or decoding fails.
func (ar AccountRetriever) GetAccount(addr sdk.AccAddress) (AccountI, error) {
	account, _, err := ar.GetAccountWithHeight(addr)
	return account, err
}

// GetAccountWithHeight queries for an account given an address. Returns the
// height of the query with the account. An error is returned if the query
// or decoding fails.
func (ar AccountRetriever) GetAccountWithHeight(addr sdk.AccAddress) (AccountI, int64, error) {
	bs, err := ar.codec.MarshalJSON(NewQueryAccountParams(addr))
	if err != nil {
		return nil, 0, err
	}

	bz, height, err := ar.querier.QueryWithData(fmt.Sprintf("custom/%s/%s", QuerierRoute, QueryAccount), bs)
	if err != nil {
		return nil, height, err
	}

	var account AccountI
	if err := ar.codec.UnmarshalJSON(bz, &account); err != nil {
		return nil, height, err
	}

	return account, height, nil
}

// EnsureExists returns an error if no account exists for the given address else nil.
func (ar AccountRetriever) EnsureExists(addr sdk.AccAddress) error {
	if _, err := ar.GetAccount(addr); err != nil {
		return err
	}
	return nil
}

// GetAccountNumberSequence returns sequence and account number for the given address.
// It returns an error if the account couldn't be retrieved from the state.
func (ar AccountRetriever) GetAccountNumberSequence(addr sdk.AccAddress) (uint64, uint64, error) {
	acc, err := ar.GetAccount(addr)
	if err != nil {
		return 0, 0, err
	}
	return acc.GetAccountNumber(), acc.GetSequence(), nil
}
