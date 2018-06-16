package types

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"

	simpleGovernance "github.com/cosmos/cosmos-sdk/examples/simpleGov/x/simple_governance"
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
		err = cdc.UnmarshalBinary(accBytes, &acct)
		if err != nil {
			panic(err)
		}
		return acct, err
	}
}

//___________________________________________________________________________________

// State to Unmarshal
type GenesisState struct {
	Accounts         []*GenesisAccount        `json:"accounts"`
	simpleGovGenesis simpleGovernance.Genesis `json:"simple_governance"`
}

// GenesisAccount doesn't need pubkey or sequence
type GenesisAccount struct {
	Name    string      `json:"name"`
	Address sdk.Address `json:"address"`
	Coins   sdk.Coins   `json:"coins"`
}

// NewGenesisAccount creates a GenesisAccount
func NewGenesisAccount(aa *AppAccount) *GenesisAccount {
	return &GenesisAccount{
		Name:    aa.Name,
		Address: aa.Address,
		Coins:   aa.Coins.Sort(),
	}
}

// ValidateBasic validates fields of the genesis account
func (ga *GenesisAccount) ValidateBasic() (err sdk.Error) {
	if !ga.Coins.IsValid() {
		return sdk.ErrInvalidCoins("")
	}
	if len(strings.TrimSpace(ga.Name)) <= 0 {
		return ErrInvalidName("Genesis name can't be blank")
	}
	return nil
}

// ToAppAccount converts a GenesisAccount to an AppAccount
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
