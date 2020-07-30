package client

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TestAccountRetriever is an AccountRetriever that can be used in unit tests
type TestAccountRetriever struct {
	Accounts map[string]struct {
		Address sdk.AccAddress
		Num     uint64
		Seq     uint64
	}
}

var _ AccountRetriever = TestAccountRetriever{}

// EnsureExists implements AccountRetriever.EnsureExists
func (t TestAccountRetriever) EnsureExists(_ NodeQuerier, addr sdk.AccAddress) error {
	_, ok := t.Accounts[addr.String()]
	if !ok {
		return fmt.Errorf("account %s not found", addr)
	}
	return nil
}

// GetAccountNumberSequence implements AccountRetriever.GetAccountNumberSequence
func (t TestAccountRetriever) GetAccountNumberSequence(_ NodeQuerier, addr sdk.AccAddress) (accNum uint64, accSeq uint64, err error) {
	acc, ok := t.Accounts[addr.String()]
	if !ok {
		return 0, 0, fmt.Errorf("account %s not found", addr)
	}
	return acc.Num, acc.Seq, nil
}
