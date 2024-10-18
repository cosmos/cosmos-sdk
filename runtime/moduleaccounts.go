package runtime

import (
	"context"

	"cosmossdk.io/core/moduleaccounts"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/pkg/errors"
)

var _ moduleaccounts.Service = (*ModuleAccountsService)(nil)

type ModuleAccountsService struct {
	accounts map[string][]byte
	ak       AccountGetter
}

// AllAccounts implements moduleaccounts.Service.
func (m *ModuleAccountsService) AllAccounts() map[string][]byte {
	return m.accounts
}

// GetAccount implements moduleaccounts.Service.
func (m *ModuleAccountsService) GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI {
	return m.ak.GetAccount(ctx, addr)
}

// GetModuleAddress implements moduleaccounts.Service.
func (m *ModuleAccountsService) GetModuleAddress(moduleName string) sdk.AccAddress {
	return sdk.AccAddress(m.accounts[moduleName])
}

type AccountGetter interface {
	GetOrSetModuleAccount(ctx context.Context, moduleName string, addr []byte) sdk.ModuleAccountI
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
}

func (m *ModuleAccountsService) Register(moduleName string) error {
	// check if the module account is already registered
	if _, ok := m.accounts[moduleName]; ok {
		return errors.Errorf("module account %s already registered", moduleName)
	}

	m.accounts[moduleName] = address.Module(moduleName)
	return nil
}

func (m *ModuleAccountsService) Account(ctx context.Context, name string) (sdk.ModuleAccountI, error) {
	addr := m.accounts[name]
	if addr == nil {
		return nil, errors.Errorf("module account %s not registered", name)
	}

	acc := m.ak.GetOrSetModuleAccount(ctx, name, addr)
	return acc, nil
}

// Address implements moduleaccounts.Service.
func (m *ModuleAccountsService) Address(name string) []byte {
	return m.accounts[name]
}

// ModuleAccount is a depinject.AutoGroupType which can be used to pass
// multiple module accounts into the depinject.
type ModuleAccount string

// IsManyPerContainerType indicates that this is a depinject.ManyPerContainerType.
func (m ModuleAccount) IsManyPerContainerType() {}
