package keeper_test

import (
	"bytes"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
	"sort"
)

// Define a Mock Quarantine Keeper that allows defining of
// function responses and recording of added quarantined funds.

// QuarantinedCoins are the arguments that were provided to AddQuarantinedCoins.
type QuarantinedCoins struct {
	coins     sdk.Coins
	toAddr    sdk.AccAddress
	fromAddrs []sdk.AccAddress
}

type MockQuarantineKeeper struct {
	// FundsHolder is the address to return from GetFundsHolder.
	FundsHolder sdk.AccAddress
	// IsQuarantinedAddrResponses are the responses that IsQuarantinedAddr should return for specific addresses.
	// The keys are the address bytes cast into a string (not the bech32).
	IsQuarantinedAddrResponses map[string]bool
	// IsAutoAcceptResponses are the responses that IsAutoAccept should return for specific addresses.
	// The keys are the concatenation of the to address bytes and each from address cast into a string (not the bech32).
	IsAutoAcceptResponses map[string]bool
	// QueuedAddQuarantinedCoinsErrors are a queue of errors to return from AddQuarantinedCoins.
	// An entry of nil means no error will be returned.
	// If this is empty, no error will be returned.
	// Entries are removed once they're used.
	QueuedAddQuarantinedCoinsErrors []error
	// AddedQuarantinedCoins are the arguments provided to AddQuarantinedCoins on calls that didn't return an error.
	AddedQuarantinedCoins []*QuarantinedCoins
}

var _ types.QuarantineKeeper = &MockQuarantineKeeper{}

func NewMockMockQuarantineKeeper(fundsHolder sdk.AccAddress) *MockQuarantineKeeper {
	return &MockQuarantineKeeper{
		FundsHolder:                     fundsHolder,
		IsQuarantinedAddrResponses:      make(map[string]bool),
		IsAutoAcceptResponses:           make(map[string]bool),
		QueuedAddQuarantinedCoinsErrors: nil,
		AddedQuarantinedCoins:           []*QuarantinedCoins{},
	}
}

func (k *MockQuarantineKeeper) WithIsQuarantinedAddrResponse(toAddr sdk.AccAddress, response bool) *MockQuarantineKeeper {
	k.IsQuarantinedAddrResponses[string(toAddr)] = response
	return k
}

func (k *MockQuarantineKeeper) WithIsAutoAcceptResponse(toAddr sdk.AccAddress, fromAddrs []sdk.AccAddress, response bool) *MockQuarantineKeeper {
	k.IsAutoAcceptResponses[k.makeComboKey(toAddr, fromAddrs)] = response
	return k
}

func (k *MockQuarantineKeeper) WithQueuedAddQuarantinedCoinsErrors(errs ...error) *MockQuarantineKeeper {
	k.QueuedAddQuarantinedCoinsErrors = append(k.QueuedAddQuarantinedCoinsErrors, errs...)
	return k
}

func (k *MockQuarantineKeeper) IsQuarantinedAddr(_ sdk.Context, toAddr sdk.AccAddress) bool {
	return k.IsQuarantinedAddrResponses[string(toAddr)]
}

func (k *MockQuarantineKeeper) IsAutoAccept(_ sdk.Context, toAddr sdk.AccAddress, fromAddrs ...sdk.AccAddress) bool {
	return k.IsAutoAcceptResponses[k.makeComboKey(toAddr, fromAddrs)]
}

func (k *MockQuarantineKeeper) GetFundsHolder() sdk.AccAddress {
	return k.FundsHolder
}

func (k *MockQuarantineKeeper) AddQuarantinedCoins(_ sdk.Context, coins sdk.Coins, toAddr sdk.AccAddress, fromAddrs ...sdk.AccAddress) error {
	if len(k.QueuedAddQuarantinedCoinsErrors) > 0 {
		err := k.QueuedAddQuarantinedCoinsErrors[0]
		k.QueuedAddQuarantinedCoinsErrors = k.QueuedAddQuarantinedCoinsErrors[1:]
		if err != nil {
			return err
		}
	}
	k.AddedQuarantinedCoins = append(k.AddedQuarantinedCoins, &QuarantinedCoins{
		coins:     coins,
		toAddr:    toAddr,
		fromAddrs: fromAddrs,
	})
	return nil
}

func (k *MockQuarantineKeeper) makeComboKey(toAddr sdk.AccAddress, fromAddrs []sdk.AccAddress) string {
	combo := make([][]byte, len(fromAddrs)+1)
	combo[0] = toAddr
	for i, addr := range fromAddrs {
		combo[i+1] = addr
	}
	// sort the from addresses so that their order doesn't matter.
	sort.Slice(combo[1:], func(i, j int) bool {
		return bytes.Compare(combo[i+1], combo[j+1]) < 0
	})
	return string(bytes.Join(combo, []byte{}))
}
