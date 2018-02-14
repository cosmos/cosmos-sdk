package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	crypto "github.com/tendermint/go-crypto"
	cmn "github.com/tendermint/tmlibs/common"
)

var _ sdk.Account = (*AppAccount)(nil)

// Custom extensions for this application.  This is just an example of
// extending auth.BaseAccount with custom fields.
//
// This is compatible with the stock auth.AccountStore, since
// auth.AccountStore uses the flexible go-wire library.
type AppAccount struct {
	auth.BaseAccount
	Name string
}

// nolint
func (acc AppAccount) GetName() string      { return acc.Name }
func (acc *AppAccount) SetName(name string) { acc.Name = name }

//___________________________________________________________________________________

// We use GenesisAccount instead of AppAccount for cleaner json input of PubKey
type GenesisAccount struct {
	Name     string         `json:"name"`
	Address  crypto.Address `json:"address"`
	Coins    sdk.Coins      `json:"coins"`
	PubKey   cmn.HexBytes   `json:"public_key"`
	Sequence int64          `json:"sequence"`
}

func NewGenesisAccount(aa *AppAccount) *GenesisAccount {
	return &GenesisAccount{
		Name:     aa.Name,
		Address:  aa.Address,
		Coins:    aa.Coins,
		PubKey:   aa.PubKey.Bytes(),
		Sequence: aa.Sequence,
	}
}

// convert GenesisAccount to AppAccount
func (ga *GenesisAccount) toAppAccount() (acc *AppAccount, err error) {
	pk, err := crypto.PubKeyFromBytes(ga.PubKey)
	if err != nil {
		return
	}
	baseAcc := auth.BaseAccount{
		Address:  ga.Address,
		Coins:    ga.Coins,
		PubKey:   pk,
		Sequence: ga.Sequence,
	}
	return &AppAccount{
		BaseAccount: baseAcc,
		Name:        ga.Name,
	}, nil
}
