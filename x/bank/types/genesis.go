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
	Params   Params    `json:"params" yaml:"params"`
	Balances []Balance `json:"balances" yaml:"balances"`
	Supply   sdk.Coins `json:"supply" yaml:"supply"`
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
func NewGenesisState(params Params, balances []Balance, supply sdk.Coins) GenesisState {
	return GenesisState{
		Params:   params,
		Balances: balances,
		Supply:   supply,
	}
}

// DefaultGenesisState returns a default bank module genesis state.
func DefaultGenesisState() GenesisState {
	return NewGenesisState(DefaultParams(), []Balance{}, DefaultSupply().GetTotal())
}

// GetGenesisStateFromAppState returns x/bank GenesisState given raw application
// genesis state.
func GetGenesisStateFromAppState(cdc codec.JSONMarshaler, appState map[string]json.RawMessage) GenesisState {
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
	cdc codec.JSONMarshaler, appState map[string]json.RawMessage, cb func(exported.GenesisBalance) (stop bool),
) {
	for _, balance := range GetGenesisStateFromAppState(cdc, appState).Balances {
		if cb(balance) {
			break
		}
	}
}
