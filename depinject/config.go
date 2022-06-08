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

// Prefer defines a container configuration for an explicit interface binding of inTypeName to outTypeName
// in global scope, for example,
//
// Prefer(
//	"github.com/cosmos/cosmos-sdk/depinject_test/depinject_test.Duck",
//	"github.com/cosmos/cosmos-sdk/depinject_test/depinject_test.Canvasback")
//
// configures the container to *always* provide a Canvasback instance when an input of interface type Duck is
// requested as an input.
func Prefer(inTypeName string, outTypeName string) Config {
	return containerConfig(func(ctr *container) error {
		return prefer(ctr, inTypeName, outTypeName, "")
	})
}

// PreferInModule defines a container configuration for an explicit interface binding of inTypeName to outTypeName
// in the scope of the module with name moduleName.  For example, given the configuration
//
// PreferInModule(
//  "moduleFoo",
//	"github.com/cosmos/cosmos-sdk/depinject_test/depinject_test.Duck",
//	"github.com/cosmos/cosmos-sdk/depinject_test/depinject_test.Canvasback")
//
// where Duck is an interface and Canvasback implements Duck, the container will attempt to provide a Canvasback
// instance where Duck is requested as an input, but only within the scope of module "moduleFoo".
func PreferInModule(moduleName string, inTypeName string, outTypeName string) Config {
	return containerConfig(func(ctr *container) error {
		return prefer(ctr, inTypeName, outTypeName, moduleName)
	})
}

func prefer(ctr *container, inTypeName string, outTypeName string, moduleName string) error {
	var mk *moduleKey
	if moduleName != "" {
		mk = &moduleKey{name: moduleName}
	}
	ctr.addPreference(preference{
		interfaceName: inTypeName,
		implTypeName:  outTypeName,
		moduleKey:     mk,
	})

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
