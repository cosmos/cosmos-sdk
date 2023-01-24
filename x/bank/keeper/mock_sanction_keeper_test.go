package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

// Define a Mock Sanction Keeper that allows defining of
// function responses and recording of calls.

type MockSanctionKeeper struct {
	// IsSanctionedAddrCalls are any addresses provided to IsSanctionedAddr
	IsSanctionedAddrCalls []sdk.AccAddress
	// IsSanctionedAddrResponses are the responses to give for specific addresses.
	// The key is an sdk.AccAddress cast to a string, e.g. string(addr) (not addr.String()).
	IsSanctionedAddrResponses map[string]bool
}

func NewMockSanctionKeeper() *MockSanctionKeeper {
	return &MockSanctionKeeper{
		IsSanctionedAddrCalls:     nil,
		IsSanctionedAddrResponses: make(map[string]bool),
	}
}

var _ types.SanctionKeeper = &MockSanctionKeeper{}

func (k *MockSanctionKeeper) WithSanctionedAddrs(addrs ...sdk.AccAddress) *MockSanctionKeeper {
	for _, addr := range addrs {
		k.IsSanctionedAddrResponses[string(addr)] = true
	}
	return k
}

func (k *MockSanctionKeeper) IsSanctionedAddr(_ sdk.Context, addr sdk.AccAddress) bool {
	k.IsSanctionedAddrCalls = append(k.IsSanctionedAddrCalls, addr)
	return k.IsSanctionedAddrResponses[string(addr)]
}
