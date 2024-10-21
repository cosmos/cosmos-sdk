package moduleaccounts

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Service interface {
	Register(name string, perms []string) error
	Address(name string) []byte // TODO: should we return an empty byte slice if it wasn't registered or should we just register it?
	// AllAccounts() map[string][]byte

	// TODO: @facu, remove these two methods
	GetModuleAddress(moduleName string) sdk.AccAddress                // TODO: @facu tmp, so I don't have to modify a bunch of things in sims right now
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI // TODO: same as above
}

type ServiceWithPerms interface {
	Service

	HasPermission(name string, perm string) bool
	IsModuleAccount(addr []byte) string // Needed in burn coins
}
