package module

import (
	"github.com/cosmos/cosmos-sdk/app/internal"
	"github.com/cosmos/cosmos-sdk/container"
	"google.golang.org/protobuf/proto"
)

type Option interface {
	apply(*internal.ModuleInitializer) error
}

func Provide(providers ...interface{}) Option {
	return option(func(initializer *internal.ModuleInitializer) error {
		for _, provider := range providers {
			desc, err := container.ExtractProviderDescriptor(provider)
			if err != nil {
				return err
			}

			initializer.ProviderFactories = append(initializer.ProviderFactories, func(cfg proto.Message) container.ProviderDescriptor {
				return desc
			})
		}
		return nil
	})
}

func Supply(values ...interface{}) Option {
	return option(func(config *internal.ModuleInitializer) error {
		panic("TODO")
	})
}

type option func(*internal.ModuleInitializer) error

func (o option) apply(opts *internal.ModuleInitializer) error {
	return o(opts)
}
