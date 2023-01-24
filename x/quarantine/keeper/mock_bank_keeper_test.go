package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/quarantine"
)

// Define a Mock Bank Keeper that defining of SendCoins errors and
// records calls made to SendCoins (but doesn't do anything else).
// Also have it just do a map lookup for GetAllBalances.

// SentCoins are the arguments that were provided to SendCoins.
type SentCoins struct {
	FromAddr sdk.AccAddress
	ToAddr   sdk.AccAddress
	Amt      sdk.Coins
}

var _ quarantine.BankKeeper = &MockBankKeeper{}

type MockBankKeeper struct {
	// SentCoins are the arguments that were provided to SendCoins, one entry for each call to it.
	SentCoins []*SentCoins
	// AllBalances are the balances to return from GetAllBalances.
	AllBalances map[string]sdk.Coins
	// QueuedSendCoinsErrors are any errors queued up to return from SendCoins.
	// An entry of nil means no error will be returned.
	// If this is empty, no error is returned.
	// Entries are removed once they're used.
	QueuedSendCoinsErrors []error
}

func NewMockBankKeeper() *MockBankKeeper {
	return &MockBankKeeper{
		SentCoins:             nil,
		AllBalances:           make(map[string]sdk.Coins),
		QueuedSendCoinsErrors: nil,
	}
}

func (k *MockBankKeeper) SetQuarantineKeeper(_ banktypes.QuarantineKeeper) {
	// do nothing.
}

func (k *MockBankKeeper) GetAllBalances(_ sdk.Context, addr sdk.AccAddress) sdk.Coins {
	return k.AllBalances[string(addr)]
}

func (k *MockBankKeeper) SendCoinsBypassQuarantine(_ sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error {
	if len(k.QueuedSendCoinsErrors) > 0 {
		err := k.QueuedSendCoinsErrors[0]
		k.QueuedSendCoinsErrors = k.QueuedSendCoinsErrors[1:]
		if err != nil {
			return err
		}
	}
	k.SentCoins = append(k.SentCoins, &SentCoins{
		FromAddr: fromAddr,
		ToAddr:   toAddr,
		Amt:      amt,
	})
	return nil
}

func (k *MockBankKeeper) SpendableCoins(_ sdk.Context, addr sdk.AccAddress) sdk.Coins {
	return k.AllBalances[string(addr)]
}
