package exported

import (
authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
)

// ModuleAccountI defines an account interface for modules that hold tokens in an escrow
type ModuleAccountI interface {
	authexported.Account
	GetName() string
	GetPermission() string
}
