package authz

import (
	"cosmossdk.io/core/registry"
	"cosmossdk.io/x/accounts/defaults/authz/types"
)

// RegisterInterfaces registers the interfaces types with the interface registry
func RegisterInterfaces(registrar registry.InterfaceRegistrar) {
	registrar.RegisterInterface(
		"cosmos.accounts.defaults.authz.Authorization",
		(*types.Authorization)(nil),
		&types.GenericAuthoriztion{},
	)
}
