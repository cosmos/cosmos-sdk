package types

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
)

// GenesisState - all auth state that must be provided at genesis
type GenesisState struct {
	Params   Params                   `json:"params" yaml:"params"`
	Accounts exported.GenesisAccounts `json:"accounts" yaml:"accounts"`
}

// NewGenesisState - Create a new genesis state
func NewGenesisState(params Params, accounts exported.GenesisAccounts) GenesisState {
	return GenesisState{
		Params:   params,
		Accounts: accounts,
	}
}

// DefaultGenesisState - Return a default genesis state
func DefaultGenesisState() GenesisState {
	return NewGenesisState(DefaultParams(), exported.GenesisAccounts{})
}

// GetGenesisStateFromAppState returns x/auth GenesisState given raw application
// genesis state.
func GetGenesisStateFromAppState(cdc *codec.Codec, appState map[string]json.RawMessage) GenesisState {
	var genesisState GenesisState
	if appState[ModuleName] != nil {
		cdc.MustUnmarshalJSON(appState[ModuleName], &genesisState)
	}

	return genesisState
}

// ValidateGenesis performs basic validation of auth genesis data returning an
// error for any failed validation criteria.
func ValidateGenesis(data GenesisState) error {
	if err := data.Params.Validate(); err != nil {
		return err
	}

	return ValidateGenAccounts(data.Accounts)
}

// SanitizeGenesisAccounts sorts accounts and coin sets.
func SanitizeGenesisAccounts(genAccs exported.GenesisAccounts) exported.GenesisAccounts {
	sort.Slice(genAccs, func(i, j int) bool {
		return genAccs[i].GetAccountNumber() < genAccs[j].GetAccountNumber()
	})

	for _, acc := range genAccs {
		if err := acc.SetCoins(acc.GetCoins().Sort()); err != nil {
			panic(err)
		}
	}

	return genAccs
}

// ValidateGenAccounts validates an array of GenesisAccounts and checks for duplicates
func ValidateGenAccounts(accounts exported.GenesisAccounts) error {
	addrMap := make(map[string]bool, len(accounts))
	for _, acc := range accounts {

		// check for duplicated accounts
		addrStr := acc.GetAddress().String()
		if _, ok := addrMap[addrStr]; ok {
			return fmt.Errorf("duplicate account found in genesis state; address: %s", addrStr)
		}

		addrMap[addrStr] = true

		// check account specific validation
		if err := acc.Validate(); err != nil {
			return fmt.Errorf("invalid account found in genesis state; address: %s, error: %s", addrStr, err.Error())
		}
	}
	return nil
}

// GenesisAccountIterator implements genesis account iteration.
type GenesisAccountIterator struct{}

// IterateGenesisAccounts iterates over all the genesis accounts found in
// appGenesis and invokes a callback on each genesis account. If any call
// returns true, iteration stops.
func (GenesisAccountIterator) IterateGenesisAccounts(
	cdc *codec.Codec, appGenesis map[string]json.RawMessage, cb func(exported.Account) (stop bool),
) {

	for _, genAcc := range GetGenesisStateFromAppState(cdc, appGenesis).Accounts {
		if cb(genAcc) {
			break
		}
	}
}
