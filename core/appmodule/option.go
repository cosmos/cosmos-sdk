package appmodule

import (
	"github.com/cosmos/cosmos-sdk/container"

	"cosmossdk.io/core/internal"
)

// Option is a functional option for implementing modules.
type Option interface {
	apply(*internal.ModuleInitializer) error
}

type funcOption func(initializer *internal.ModuleInitializer) error

func (f funcOption) apply(initializer *internal.ModuleInitializer) error {
	return f(initializer)
}

// Provide registers providers with the dependency injection system that will be
// run within the module scope. See github.com/cosmos/cosmos-sdk/container for
// documentation on the dependency injection system.
func Provide(providers ...interface{}) Option {
	return funcOption(func(initializer *internal.ModuleInitializer) error {
		for _, provider := range providers {
			desc, err := container.ExtractProviderDescriptor(provider)
			if err != nil {
				return err
			}

			initializer.Providers = append(initializer.Providers, desc)
		}
		return nil
	})
}
