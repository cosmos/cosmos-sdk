package depinject

import (
	"reflect"

	"github.com/pkg/errors"
)

// Config is a functional configuration of a container.
type Config interface {
	apply(*container) error
}

// Provide defines a container configuration which registers the provided dependency
// injection providers. Each provider will be called at most once with the
// exception of module-scoped providers which are called at most once per module
// (see ModuleKey).
func Provide(providers ...interface{}) Config {
	return containerConfig(func(ctr *container) error {
		return provide(ctr, nil, providers)
	})
}

// ProvideInModule defines container configuration which registers the provided dependency
// injection providers that are to be run in the named module. Each provider
// will be called at most once.
func ProvideInModule(moduleName string, providers ...interface{}) Config {
	return containerConfig(func(ctr *container) error {
		if moduleName == "" {
			return errors.Errorf("expected non-empty module name")
		}

		return provide(ctr, ctr.createOrGetModuleKey(moduleName), providers)
	})
}

func provide(ctr *container, key *moduleKey, providers []interface{}) error {
	for _, c := range providers {
		rc, err := ExtractProviderDescriptor(c)
		if err != nil {
			return errors.WithStack(err)
		}
		_, err = ctr.addNode(&rc, key)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func Supply(values ...interface{}) Config {
	loc := LocationFromCaller(1)
	return containerConfig(func(ctr *container) error {
		for _, v := range values {
			err := ctr.supply(reflect.ValueOf(v), loc)
			if err != nil {
				return errors.WithStack(err)
			}
		}
		return nil
	})
}

// Error defines configuration which causes the dependency injection container to
// fail immediately.
func Error(err error) Config {
	return containerConfig(func(*container) error {
		return errors.WithStack(err)
	})
}

// Configs defines a configuration which bundles together multiple Config definitions.
func Configs(opts ...Config) Config {
	return containerConfig(func(ctr *container) error {
		for _, opt := range opts {
			err := opt.apply(ctr)
			if err != nil {
				return errors.WithStack(err)
			}
		}
		return nil
	})
}

type containerConfig func(*container) error

func (c containerConfig) apply(ctr *container) error {
	return c(ctr)
}

var _ Config = (*containerConfig)(nil)
