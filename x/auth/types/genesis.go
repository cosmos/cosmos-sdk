package types

import (
	"fmt"
	"sort"

	"github.com/cosmos/cosmos-sdk/x/auth/exported"
)

// GenesisState - all auth state that must be provided at genesis
type GenesisState struct {
	Params   Params                    `json:"params" yaml:"params"`
	Accounts []exported.GenesisAccount `json:"accounts" yaml:"accounts"`
}

// NewGenesisState - Create a new genesis state
func NewGenesisState(params Params, accounts []exported.GenesisAccount) GenesisState {
	return GenesisState{
		Params:   params,
		Accounts: accounts,
	}
}

// DefaultGenesisState - Return a default genesis state
func DefaultGenesisState() GenesisState {
	return NewGenesisState(DefaultParams(), []exported.GenesisAccount{})
}

// ValidateGenesis performs basic validation of auth genesis data returning an
// error for any failed validation criteria.
func ValidateGenesis(data GenesisState) error {
	if err := data.Params.Validate(); err != nil {
		return err
	}

	return validateGenAccounts(data.Accounts)
}

// Sanitize sorts accounts and coin sets.
func Sanitize(genAccs []exported.GenesisAccount) []exported.GenesisAccount {
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

func validateGenAccounts(accounts []exported.GenesisAccount) error {
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
