package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

// State to Unmarshal
type GenesisState struct {
	Accounts            []*GenesisAccount `json:"accounts"`
	BondingTransactions [][]byte          `json:"bonding_tx"`
}

// GenesisAccount doesn't need pubkey or sequence
type GenesisAccount struct {
	Name    string      `json:"name"`
	Address sdk.Address `json:"address"`
	Coins   sdk.Coins   `json:"coins"`
}

func NewGenesisAccount(aa *AppAccount) *GenesisAccount {
	return &GenesisAccount{
		Name:    aa.Name,
		Address: aa.Address,
		Coins:   aa.Coins,
	}
}

// convert GenesisAccount to AppAccount
func (ga *GenesisAccount) ToAppAccount() (acc *AppAccount, err error) {
	baseAcc := auth.BaseAccount{
		Address: ga.Address,
		Coins:   ga.Coins,
	}
	return &AppAccount{
		BaseAccount: baseAcc,
		Name:        ga.Name,
	}, nil
}
