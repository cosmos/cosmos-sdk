package exported

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/auth/exported"
)

// ModuleAccountI defines an account interface for modules that hold tokens in
// an escrow.
type ModuleAccountI interface {
	exported.Account

	GetName() string
	GetPermissions() []string
	HasPermission(string) bool
}

// SupplyI defines an inflationary supply interface for modules that handle
// token supply.
type SupplyI interface {
	GetTotal() sdk.Coins
	SetTotal(total sdk.Coins)

	Inflate(amount sdk.Coins)
	Deflate(amount sdk.Coins)

	String() string
	ValidateBasic() error
}
