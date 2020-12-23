package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/exported"
)

var _ exported.GenesisBalance = (*Balance)(nil)

// GetAddress returns the account address of the Balance object.
func (b Balance) GetAddress() sdk.AccAddress {
	addr1, _ := sdk.AccAddressFromBech32(b.Address)
	return addr1
}

// GetAddress returns the account coins of the Balance object.
func (b Balance) GetCoins() sdk.Coins {
	return b.Coins
}

// SanitizeGenesisAccounts sorts addresses and coin sets.
func SanitizeGenesisBalances(balances []Balance) []Balance {
	sort.Slice(balances, func(i, j int) bool {
		addr1, _ := sdk.AccAddressFromBech32(balances[i].Address)
		addr2, _ := sdk.AccAddressFromBech32(balances[j].Address)
		return bytes.Compare(addr1.Bytes(), addr2.Bytes()) < 0
	})

	for _, balance := range balances {
		balance.Coins = balance.Coins.Sort()
	}

	return balances
}

// ValidateGenesis performs basic validation of supply genesis data returning an
// error for any failed validation criteria.
func ValidateGenesis(data GenesisState) error {
	if err := data.Params.Validate(); err != nil {
		return err
	}

	totalSupply := NewSupply(data.Supply)
	err := totalSupply.ValidateBasic()
	if err != nil {
		return err
	}

	var accsSupply sdk.Coins
	for _, balance := range data.Balances {
		accsSupply = append(accsSupply, balance.Coins...)
	}

	if !accsSupply.IsEqual(data.Supply) {
		return fmt.Errorf("total supply does not match with accounts balance")
	}
	return nil
}

// NewGenesisState creates a new genesis state.
func NewGenesisState(params Params, balances []Balance, supply sdk.Coins, denomMetaData []Metadata) *GenesisState {
	return &GenesisState{
		Params:        params,
		Balances:      balances,
		Supply:        supply,
		DenomMetadata: denomMetaData,
	}
}

// DefaultGenesisState returns a default bank module genesis state.
func DefaultGenesisState() *GenesisState {
	return NewGenesisState(DefaultParams(), []Balance{}, DefaultSupply().GetTotal(), []Metadata{})
}

// GetGenesisStateFromAppState returns x/bank GenesisState given raw application
// genesis state.
func GetGenesisStateFromAppState(cdc codec.JSONMarshaler, appState map[string]json.RawMessage) *GenesisState {
	var genesisState GenesisState

	if appState[ModuleName] != nil {
		cdc.MustUnmarshalJSON(appState[ModuleName], &genesisState)
	}

	return &genesisState
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
