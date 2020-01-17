package types

import (
	"bytes"
	"encoding/json"
	"sort"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/exported"
)

// GenesisState is the bank state that must be provided at genesis.
type GenesisState struct {
	SendEnabled     bool             `json:"send_enabled" yaml:"send_enabled"`
	GenesisBalances []GenesisBalance `json:"genesis_balances" yaml:"genesis_balances"`
}

// Genesis Balance is a struct that pairs an sdk.AccAddress with an accounts balance at genesis
type GenesisBalance struct {
	Address sdk.AccAddress
	Coins   sdk.Coins
}

// GetAddress implements exported interface
func (bal GenesisBalance) GetAddress() sdk.AccAddress {
	return bal.Address
}

// GetCoins implements exported interface
func (bal GenesisBalance) GetCoins() sdk.Coins {
	return bal.Coins
}

// SanitizeGenesisAccounts sorts addresses and coin sets.
func SanitizeGenesisBalances(genesisBalances []GenesisBalance) []GenesisBalance {
	sort.Slice(genesisBalances, func(i, j int) bool {
		return bytes.Compare(genesisBalances[i].Address.Bytes(), genesisBalances[j].Address.Bytes()) < 0
	})

	for _, genBalance := range genesisBalances {
		genBalance.Coins = genBalance.Coins.Sort()
	}

	return genesisBalances
}

// NewGenesisState creates a new genesis state.
func NewGenesisState(sendEnabled bool, genesisBalances []GenesisBalance) GenesisState {
	return GenesisState{SendEnabled: sendEnabled, GenesisBalances: genesisBalances}
}

// DefaultGenesisState returns a default genesis state
func DefaultGenesisState() GenesisState { return NewGenesisState(true, []GenesisBalance{}) }

// ValidateGenesis performs basic validation of bank genesis data returning an
// error for any failed validation criteria.
func ValidateGenesis(data GenesisState) error {
	return nil
}

// GetGenesisStateFromAppState returns x/bank GenesisState given raw application
// genesis state.
func GetGenesisStateFromAppState(cdc *codec.Codec, appState map[string]json.RawMessage) GenesisState {
	var genesisState GenesisState
	if appState[ModuleName] != nil {
		cdc.MustUnmarshalJSON(appState[ModuleName], &genesisState)
	}

	return genesisState
}

// GenesisAccountIterator implements genesis account iteration.
type GenesisBalancesIterator struct{}

// IterateGenesisAccounts iterates over all the genesis accounts found in
// appGenesis and invokes a callback on each genesis account. If any call
// returns true, iteration stops.
func (GenesisBalancesIterator) IterateGenesisBalances(
	cdc *codec.Codec, appGenesis map[string]json.RawMessage, cb func(exported.GenesisBalance) (stop bool),
) {
	for _, genBal := range GetGenesisStateFromAppState(cdc, appGenesis).GenesisBalances {
		if cb(genBal) {
			break
		}
	}
}
