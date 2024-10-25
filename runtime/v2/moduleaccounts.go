package runtime

import (
	"bytes"

	"github.com/pkg/errors"

	"cosmossdk.io/core/moduleaccounts"

	"github.com/cosmos/cosmos-sdk/types/address"
)

var _ moduleaccounts.ServiceWithPerms = (*ModuleAccountsService)(nil)

type addrWithPerms struct {
	addr  []byte
	perms []string
}

type ModuleAccountsService struct {
	accounts map[string]addrWithPerms
}

// HasPermission implements moduleaccounts.ServiceWithPerms.
func (m *ModuleAccountsService) HasPermission(name, perm string) bool {
	for _, v := range m.accounts[name].perms {
		if v == perm {
			return true
		}
	}
	return false
}

// IsModuleAccount implements moduleaccounts.ServiceWithPerms.
func (m *ModuleAccountsService) IsModuleAccount(addr []byte) string {
	for name, v := range m.accounts {
		if bytes.Equal(addr, v.addr) {
			return name
		}
	}
	return ""
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
	accs := map[string][]byte{}
	for k, v := range m.accounts {
		accs[k] = v.addr
	}
	return accs
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

func ProvideModuleAccountsService() (moduleaccounts.Service, moduleaccounts.ServiceWithPerms) {
	svc := NewModuleAccountsService()
	return svc, svc
}

func SetupModuleAccountsService(
	accounts []ModuleAccount,
	service moduleaccounts.Service,
) error {
	for _, acc := range accounts {
		if err := service.Register(acc.Name, acc.Permissions); err != nil {
			return err
		}
	}
	return nil
}
