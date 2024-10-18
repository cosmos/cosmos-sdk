package moduleaccounts

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Service interface {
	Register(string) error
	Address(name string) []byte
	Account(ctx context.Context, name string) (sdk.ModuleAccountI, error)
	AllAccounts() map[string][]byte

	GetModuleAddress(moduleName string) sdk.AccAddress                // TODO: @facu tmp, so I don't have to modify a bunch of things in sims right now
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI // TODO: same as above
}
