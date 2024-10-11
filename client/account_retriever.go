package client

import (
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Account defines a read-only version of the auth module's AccountI.
type Account interface {
	GetAddress() sdk.AccAddress
	GetPubKey() cryptotypes.PubKey // can return nil.
	GetAccountNumber() uint64
	GetSequence() uint64
}

type mockAccount struct {
	addr []byte
}

func (m mockAccount) GetAddress() sdk.AccAddress {
	return m.addr
}

func (m mockAccount) GetPubKey() cryptotypes.PubKey {
	return nil
}

func (m mockAccount) GetAccountNumber() uint64 {
	return 0
}

func (m mockAccount) GetSequence() uint64 {
	return 0
}

// AccountRetriever defines the interfaces required by transactions to
// ensure an account exists and to be able to query for account fields necessary
// for signing.
type AccountRetriever interface {
	GetAccount(clientCtx Context, addr sdk.AccAddress) (Account, error)
	GetAccountWithHeight(clientCtx Context, addr sdk.AccAddress) (Account, int64, error)
	EnsureExists(clientCtx Context, addr sdk.AccAddress) error
	GetAccountNumberSequence(clientCtx Context, addr sdk.AccAddress) (accNum, accSeq uint64, err error)
}

var _ AccountRetriever = (*MockAccountRetriever)(nil)

// MockAccountRetriever defines a no-op basic AccountRetriever that can be used
// in mocked contexts. Tests or context that need more sophisticated testing
// state should implement their own mock AccountRetriever.
type MockAccountRetriever struct {
	ReturnAccNum, ReturnAccSeq uint64
}

func (mar MockAccountRetriever) GetAccount(_ Context, address sdk.AccAddress) (Account, error) {
	return mockAccount{addr: address}, nil
}

func (mar MockAccountRetriever) GetAccountWithHeight(_ Context, _ sdk.AccAddress) (Account, int64, error) {
	return nil, 0, nil
}

func (mar MockAccountRetriever) EnsureExists(_ Context, _ sdk.AccAddress) error {
	return nil
}

func (mar MockAccountRetriever) GetAccountNumberSequence(_ Context, _ sdk.AccAddress) (uint64, uint64, error) {
	return mar.ReturnAccNum, mar.ReturnAccSeq, nil
}
