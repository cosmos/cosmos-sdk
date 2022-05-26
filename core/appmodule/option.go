package appmodule

import (
	"cosmossdk.io/core/internal"
	"cosmossdk.io/depinject"
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
// run within the module scope. See cosmossdk.io/depinject v1.0.0-alpha.3 for
// documentation on the dependency injection system.
func Provide(providers ...interface{}) Option {
	return funcOption(func(initializer *internal.ModuleInitializer) error {
		for _, provider := range providers {
			desc, err := depinject.ExtractProviderDescriptor(provider)
			if err != nil {
				return err
			}

			initializer.Providers = append(initializer.Providers, desc)
		}
		return nil
	})
}
