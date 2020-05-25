package types

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AccountRetriever defines the properties of a type that can be used to
// retrieve accounts.
type AccountRetriever struct {
	codec codec.Marshaler
}

// NewAccountRetriever initialises a new AccountRetriever instance.
func NewAccountRetriever(codec codec.Marshaler) AccountRetriever {
	return AccountRetriever{codec: codec}
}

// GetAccount queries for an account given an address and a block height. An
// error is returned if the query or decoding fails.
func (ar AccountRetriever) GetAccount(querier context.NodeQuerier, addr sdk.AccAddress) (AccountI, error) {
	account, _, err := ar.GetAccountWithHeight(querier, addr)
	return account, err
}

// GetAccountWithHeight queries for an account given an address. Returns the
// height of the query with the account. An error is returned if the query
// or decoding fails.
func (ar AccountRetriever) GetAccountWithHeight(querier context.NodeQuerier, addr sdk.AccAddress) (AccountI, int64, error) {
	bs, err := ar.codec.MarshalJSON(NewQueryAccountParams(addr))
	if err != nil {
		return nil, 0, err
	}

	bz, height, err := querier.QueryWithData(fmt.Sprintf("custom/%s/%s", QuerierRoute, QueryAccount), bs)
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
func (ar AccountRetriever) EnsureExists(querier context.NodeQuerier, addr sdk.AccAddress) error {
	if _, err := ar.GetAccount(querier, addr); err != nil {
		return err
	}
	return nil
}

// GetAccountNumberSequence returns sequence and account number for the given address.
// It returns an error if the account couldn't be retrieved from the state.
func (ar AccountRetriever) GetAccountNumberSequence(nodeQuerier context.NodeQuerier, addr sdk.AccAddress) (uint64, uint64, error) {
	acc, err := ar.GetAccount(nodeQuerier, addr)
	if err != nil {
		return 0, 0, err
	}
	return acc.GetAccountNumber(), acc.GetSequence(), nil
}
