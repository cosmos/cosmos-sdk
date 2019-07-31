package exported

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/auth/exported"
)

// ModuleAccountI defines an account interface for modules that hold tokens in an escrow
type ModuleAccountI interface {
	exported.Account

	GetName() string
	GetPermissions() []string
	HasPermission(string) bool
}

// TODO: ...
type SupplyI interface {
	GetTotal() sdk.Coins
	SetTotal(total sdk.Coins) SupplyI

	Inflate(amount sdk.Coins) SupplyI
	Deflate(amount sdk.Coins) SupplyI

	String() string
	ValidateBasic() error
}
