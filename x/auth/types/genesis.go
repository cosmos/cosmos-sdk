package types

import (
	"encoding/json"
	"fmt"
	"sort"

	proto "github.com/gogo/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/module"
)

var _ types.UnpackInterfacesMessage = GenesisState{}

// RandomGenesisAccountsFn defines the function required to generate custom account types
type RandomGenesisAccountsFn func(simState *module.SimulationState) GenesisAccounts

// NewGenesisState - Create a new genesis state
func NewGenesisState(params Params, accounts GenesisAccounts) *GenesisState {
	genAccounts, err := PackAccounts(accounts)
	if err != nil {
		panic(err)
	}
	return &GenesisState{
		Params:   params,
		Accounts: genAccounts,
	}
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (g GenesisState) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	for _, any := range g.Accounts {
		var account GenesisAccount
		err := unpacker.UnpackAny(any, &account)
		if err != nil {
			return err
		}
	}
	return nil
}

// DefaultGenesisState - Return a default genesis state
func DefaultGenesisState() *GenesisState {
	return NewGenesisState(DefaultParams(), GenesisAccounts{})
}

// GetGenesisStateFromAppState returns x/auth GenesisState given raw application
// genesis state.
func GetGenesisStateFromAppState(cdc codec.Marshaler, appState map[string]json.RawMessage) GenesisState {
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

	genAccs, err := UnpackAccounts(data.Accounts)
	if err != nil {
		return err
	}

	return ValidateGenAccounts(genAccs)
}

// SanitizeGenesisAccounts sorts accounts and coin sets.
func SanitizeGenesisAccounts(genAccs GenesisAccounts) GenesisAccounts {
	sort.Slice(genAccs, func(i, j int) bool {
		return genAccs[i].GetAccountNumber() < genAccs[j].GetAccountNumber()
	})

	return genAccs
}

// ValidateGenAccounts validates an array of GenesisAccounts and checks for duplicates
func ValidateGenAccounts(accounts GenesisAccounts) error {
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
	cdc codec.Marshaler, appGenesis map[string]json.RawMessage, cb func(AccountI) (stop bool),
) {
	for _, genAcc := range GetGenesisStateFromAppState(cdc, appGenesis).Accounts {
		acc, ok := genAcc.GetCachedValue().(AccountI)
		if !ok {
			panic("expected account")
		}
		if cb(acc) {
			break
		}
	}
}

// PackAccounts converts GenesisAccounts to Any slice
func PackAccounts(accounts GenesisAccounts) ([]*types.Any, error) {
	accountsAny := make([]*types.Any, len(accounts))
	for i, acc := range accounts {
		msg, ok := acc.(proto.Message)
		if !ok {
			return nil, fmt.Errorf("cannot proto marshal %T", acc)
		}
		any, err := types.NewAnyWithValue(msg)
		if err != nil {
			return nil, err
		}
		accountsAny[i] = any
	}

	return accountsAny, nil
}

// UnpackAccounts converts Any slice to GenesisAccounts
func UnpackAccounts(accountsAny []*types.Any) (GenesisAccounts, error) {
	accounts := make(GenesisAccounts, len(accountsAny))
	for i, any := range accountsAny {
		acc, ok := any.GetCachedValue().(GenesisAccount)
		if !ok {
			return nil, fmt.Errorf("expected genesis account")
		}
		accounts[i] = acc
	}

	return accounts, nil
}
