package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"

	"github.com/cosmos/cosmos-sdk/examples/democoin/x/cool"
	"github.com/cosmos/cosmos-sdk/examples/democoin/x/pow"
)

var _ auth.Account = (*AppAccount)(nil)

// Custom extensions for this application.  This is just an example of
// extending auth.BaseAccount with custom fields.
//
// This is compatible with the stock auth.AccountStore, since
// auth.AccountStore uses the flexible go-amino library.
type AppAccount struct {
	auth.BaseAccount
	Name string `json:"name"`
}

// Constructor for AppAccount
func ProtoAppAccount() auth.Account {
	return &AppAccount{}
}

// nolint
func (acc AppAccount) GetName() string      { return acc.Name }
func (acc *AppAccount) SetName(name string) { acc.Name = name }

// Get the AccountDecoder function for the custom AppAccount
func GetAccountDecoder(cdc *wire.Codec) auth.AccountDecoder {
	return func(accBytes []byte) (res auth.Account, err error) {
		if len(accBytes) == 0 {
			return nil, sdk.ErrTxDecode("accBytes are empty")
		}
		acct := new(AppAccount)
		err = cdc.UnmarshalBinaryBare(accBytes, &acct)
		if err != nil {
			panic(err)
		}
		return acct, err
	}
}

//___________________________________________________________________________________

// State to Unmarshal
type GenesisState struct {
	Accounts    []*GenesisAccount `json:"accounts"`
	POWGenesis  pow.Genesis       `json:"pow"`
	CoolGenesis cool.Genesis      `json:"cool"`
}

// GenesisAccount doesn't need pubkey or sequence
type GenesisAccount struct {
	Name    string         `json:"name"`
	Address sdk.AccAddress `json:"address"`
	Coins   sdk.Coins      `json:"coins"`
}

func NewGenesisAccount(aa *AppAccount) *GenesisAccount {
	return &GenesisAccount{
		Name:    aa.Name,
		Address: aa.Address,
		Coins:   aa.Coins.Sort(),
	}
}

// convert GenesisAccount to AppAccount
func (ga *GenesisAccount) ToAppAccount() (acc *AppAccount, err error) {
	baseAcc := auth.BaseAccount{
		Address: ga.Address,
		Coins:   ga.Coins.Sort(),
	}
	return &AppAccount{
		BaseAccount: baseAcc,
		Name:        ga.Name,
	}, nil
}
