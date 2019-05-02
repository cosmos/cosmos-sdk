package testnet

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

// The expected interface for iterating genesis accounts object
type GenesisAccountsIterator interface {
	IterateGenesisAccounts(
		cdc *codec.Codec,
		appState map[string]json.RawMessage,
		iterateFn func(auth.Account) (stop bool),
	)
}
