package bank

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/mock"
)

// GetBenchmarkMockApp initializes a mock application for this module, for purposes of benchmarking
// Any long term API support commitments do not apply to this function.
func GetBenchmarkMockApp() (*mock.App, error) {
	mapp := mock.NewApp()

	RegisterWire(mapp.Cdc)
	coinKeeper := NewKeeper(mapp.AccountMapper)
	mapp.Router().AddRoute("bank", NewHandler(coinKeeper))

	err := mapp.CompleteSetup([]*sdk.KVStoreKey{})
	return mapp, err
}
