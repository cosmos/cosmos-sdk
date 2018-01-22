package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

var _ sdk.Account = (*AppAccount)(nil)

type AppAccount struct {
	auth.BaseAccount

	// Custom extensions for this application.  This is just an example of
	// extending auth.BaseAccount with custom fields.
	//
	// This is compatible with the stock auth.AccountStore, since
	// auth.AccountStore uses the flexible go-wire library.
	Name string
}

func (acc AppAccount) GetName() string {
	return acc.Name
}

func (acc *AppAccount) SetName(name string) {
	acc.Name = name
}
