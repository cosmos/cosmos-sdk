package appmodule

import (
	"github.com/cosmos/cosmos-sdk/container"

	"cosmossdk.io/core/internal"
)

type Option interface {
	apply(*internal.ModuleInitializer) error
}

type funcOption func(initializer *internal.ModuleInitializer) error

func (f funcOption) apply(initializer *internal.ModuleInitializer) error {
	return f(initializer)
}

// Provide registers providers with the dependency injection system. See
// github.com/cosmos/cosmos-sdk/container for more documentation on the
// dependency injection system.
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
