package container

import (
	"reflect"

	"github.com/pkg/errors"
)

// Option is a functional option for a container.
type Option interface {
	apply(*container) error
}

// Provide creates a container option which registers the provided dependency
// injection providers. Each provider will be called at most once with the
// exception of module-scoped providers which are called at most once per module
// (see ModuleKey).
func Provide(providers ...interface{}) Option {
	return containerOption(func(ctr *container) error {
		return provide(ctr, nil, providers)
	})
}

// ProvideInModule creates a container option which registers the provided dependency
// injection providers that are to be run in the named module. Each provider
// will be called at most once.
func ProvideInModule(moduleName string, providers ...interface{}) Option {
	return containerOption(func(ctr *container) error {
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

func Supply(values ...interface{}) Option {
	loc := LocationFromCaller(1)
	return containerOption(func(ctr *container) error {
		for _, v := range values {
			err := ctr.supply(reflect.ValueOf(v), loc)
			if err != nil {
				return errors.WithStack(err)
			}
		}
		return nil
	})
}

// Error creates an option which causes the dependency injection container to
// fail immediately.
func Error(err error) Option {
	return containerOption(func(*container) error {
		return errors.WithStack(err)
	})
}

// Options creates an option which bundles together other options.
func Options(opts ...Option) Option {
	return containerOption(func(ctr *container) error {
		for _, opt := range opts {
			err := opt.apply(ctr)
			if err != nil {
				return errors.WithStack(err)
			}
		}
		return nil
	})
}

type containerOption func(*container) error

func (c containerOption) apply(ctr *container) error {
	return c(ctr)
}

var _ Option = (*containerOption)(nil)
