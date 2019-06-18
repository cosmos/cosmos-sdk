package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NodeQuerier is an interface that is satisfied by types that provide the QueryWithData method
type NodeQuerier interface {
	// QueryWithData performs a query to a Tendermint node with the provided path
	// and a data payload. It returns the result and height of the query upon success
	// or an error if the query fails.
	QueryWithData(path string, data []byte) ([]byte, int64, error)
}

type AccountGetter struct {
	querier NodeQuerier
}

func NewAccountGetter(querier NodeQuerier) AccountGetter {
	return AccountGetter{querier: querier}
}

// GetAccount queries for an account given an address and a block height. An
// error is returned if the query or decoding fails.
func (ag AccountGetter) GetAccount(addr sdk.AccAddress) (Account, error) {
	bs, err := ModuleCdc.MarshalJSON(NewQueryAccountParams(addr))
	if err != nil {
		return nil, err
	}

	res, _, err := ag.querier.QueryWithData(fmt.Sprintf("custom/%s/%s", QuerierRoute, QueryAccount), bs)
	if err != nil {
		return nil, err
	}

	var account Account
	if err := ModuleCdc.UnmarshalJSON(res, &account); err != nil {
		return nil, err
	}

	return account, nil
}

func (ag AccountGetter) EnsureExists(addr sdk.AccAddress) error {
	if _, err := ag.GetAccount(addr); err != nil {
		return err
	}
	return nil
}

func (ag AccountGetter) GetAccountNumberSequence(addr sdk.AccAddress) (uint64, uint64, error) {
	acc, err := ag.GetAccount(addr)
	if err != nil {
		return 0, 0, err
	}
	return acc.GetAccountNumber(), acc.GetSequence(), nil
}
