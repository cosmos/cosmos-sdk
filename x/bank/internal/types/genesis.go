package types

import (
	"bytes"
	"encoding/json"
	"sort"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/exported"
)

var _ exported.GenesisBalance = (*Balance)(nil)

// GenesisState defines the bank module's genesis state.
type GenesisState struct {
	SendEnabled bool      `json:"send_enabled" yaml:"send_enabled"`
	Balances    []Balance `json:"balances" yaml:"balances"`
}

// Balance defines an account address and balance pair used in the bank module's
// genesis state.
type Balance struct {
	Address sdk.AccAddress `json:"address" yaml:"address"`
	Coins   sdk.Coins      `json:"coins" yaml:"coins"`
}

// GetAddress returns the account address of the Balance object.
func (b Balance) GetAddress() sdk.AccAddress {
	return b.Address
}

// GetAddress returns the account coins of the Balance object.
func (b Balance) GetCoins() sdk.Coins {
	return b.Coins
}

// SanitizeGenesisAccounts sorts addresses and coin sets.
func SanitizeGenesisBalances(balances []Balance) []Balance {
	sort.Slice(balances, func(i, j int) bool {
		return bytes.Compare(balances[i].Address.Bytes(), balances[j].Address.Bytes()) < 0
	})

	for _, balance := range balances {
		balance.Coins = balance.Coins.Sort()
	}

	return balances
}

// NewGenesisState creates a new genesis state.
func NewGenesisState(sendEnabled bool, balances []Balance) GenesisState {
	return GenesisState{SendEnabled: sendEnabled, Balances: balances}
}

// DefaultGenesisState returns a default bank module genesis state.
func DefaultGenesisState() GenesisState { return NewGenesisState(true, []Balance{}) }

// ValidateGenesis performs basic validation of bank genesis data returning an
// error for any failed validation criteria.
func ValidateGenesis(data GenesisState) error { return nil }

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
	cdc *codec.Codec, appState map[string]json.RawMessage, cb func(exported.GenesisBalance) (stop bool),
) {
	for _, balance := range GetGenesisStateFromAppState(cdc, appState).Balances {
		if cb(balance) {
			break
		}
	}
}
