package runtime

import (
	"context"

	"cosmossdk.io/core/moduleaccounts"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/pkg/errors"
)

var _ moduleaccounts.Service = (*ModuleAccountsService)(nil)

type addrWithPerms struct {
	addr  []byte
	perms []string
}

type ModuleAccountsService struct {
	accounts map[string]addrWithPerms
	ak       AccountGetter
}

func NewModuleAccountsService(moduleAccounts ...ModuleAccount) *ModuleAccountsService {
	svc := &ModuleAccountsService{
		accounts: make(map[string]addrWithPerms),
	}

	for _, acc := range moduleAccounts {
		// error if there are dups
		if _, ok := svc.accounts[acc.Name]; ok {
			panic(errors.Errorf("module account %s already registered", acc))
		}

		svc.accounts[acc.Name] = addrWithPerms{
			addr:  address.Module(acc.Name),
			perms: acc.Permissions,
		}
	}

	return svc
}

// AllAccounts implements moduleaccounts.Service.
func (m *ModuleAccountsService) AllAccounts() map[string][]byte {
	// return m.accounts
	return map[string][]byte{}
}

// GetAccount implements moduleaccounts.Service.
func (m *ModuleAccountsService) GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI {
	return m.ak.GetAccount(ctx, addr)
}

// GetModuleAddress implements moduleaccounts.Service.
func (m *ModuleAccountsService) GetModuleAddress(moduleName string) sdk.AccAddress {
	return sdk.AccAddress(m.accounts[moduleName].addr)
}

type AccountGetter interface {
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
}

func (m *ModuleAccountsService) Register(moduleName string, perms []string) error {
	// check if the module account is already registered
	if _, ok := m.accounts[moduleName]; ok {
		return errors.Errorf("module account %s already registered", moduleName)
	}

	m.accounts[moduleName] = addrWithPerms{
		addr:  address.Module(moduleName),
		perms: perms,
	}
	return nil
}

// Address implements moduleaccounts.Service.
func (m *ModuleAccountsService) Address(name string) []byte {
	return m.accounts[name].addr
}

// ModuleAccount is a depinject.AutoGroupType which can be used to pass
// multiple module accounts into the depinject.
type ModuleAccount struct {
	Name        string
	Permissions []string
}

func NewModuleAccount(name string, permissions ...string) ModuleAccount {
	return ModuleAccount{
		Name:        name,
		Permissions: permissions,
	}
}

// IsManyPerContainerType indicates that this is a depinject.ManyPerContainerType.
func (m ModuleAccount) IsManyPerContainerType() {}
