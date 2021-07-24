package container

import (
	"reflect"

	"github.com/pkg/errors"
)

// Option is a functional option for a container.
type Option interface {
	applyConfig(*config) error
	applyContainer(*container) error
}

// Provide creates a container option which registers the provided dependency
// injection constructors. Each constructor will be called at most once with the
// exception of scoped constructors which are called at most once per scope
// (see Scope).
func Provide(constructors ...interface{}) Option {
	return containerOption(func(ctr *container) error {
		return provide(ctr, nil, constructors)
	})
}

// ProvideWithScope creates a container option which registers the provided dependency
// injection constructors that are to be run in the provided scope. Each constructor
// will be called at most once.
func ProvideWithScope(scope Scope, constructors ...interface{}) Option {
	return containerOption(func(ctr *container) error {
		if scope == nil {
			return errors.Errorf("expected non-empty scope")
		}

		return provide(ctr, scope, constructors)
	})
}

func provide(ctr *container, scope Scope, constructors []interface{}) error {
	for _, c := range constructors {
		rc, err := reflectCtr(c)
		if err != nil {
			return err
		}
		_, err = ctr.addNode(rc, scope, false)
		if err != nil {
			return err
		}
	}
	return nil
}

// AutoGroupTypes creates an option which registers the provided types as types which
// will automatically get grouped together. For a given type T, T and []T can
// be declared as output parameters for constructors as many times within the container
// as desired. All of the provided values for T can be retrieved by declaring an
// []T input parameter.
func AutoGroupTypes(types ...reflect.Type) Option {
	return configOption(func(c *config) error {
		for _, ty := range types {
			if ty.Kind() == reflect.Slice {
				return errors.Errorf("slice type %T cannot be used as an auto-group type", ty)
			}

			if c.onePerScopeTypes[ty] {
				return errors.Errorf("type %v is already registered as a one per scope type, trying to mark as an auto-group type", ty)
			}
			c.logf("Registering auto-group type %v", ty)
			c.autoGroupTypes[ty] = true
		}
		return nil
	})
}

// OnePerScopeTypes creates an option which registers the provided types as types which
// can have up to one value per scope. All of the values for a one-per-scope type T
// and their respective scopes, can be retrieved by declaring an input parameter map[Scope]T.
func OnePerScopeTypes(types ...reflect.Type) Option {
	return configOption(func(c *config) error {
		for _, ty := range types {
			if c.autoGroupTypes[ty] {
				return errors.Errorf("type %v is already registered as an auto-group type, trying to mark as one per scope type", ty)
			}
			c.logf("Registering one-per-sope type %v", ty)
			c.onePerScopeTypes[ty] = true
		}
		return nil
	})
}

func Logger(logger func(string)) Option {
	return configOption(func(c *config) error {
		logger("Initializing logger")
		c.loggers = append(c.loggers, logger)
		return nil
	})
}

// Error creates an option which causes the dependency injection container to
// fail immediately.
func Error(err error) Option {
	return configOption(func(*config) error {
		return errors.WithStack(err)
	})
}

// Options creates an option which bundles together other options.
func Options(opts ...Option) Option {
	return option{
		configOption: func(cfg *config) error {
			for _, opt := range opts {
				err := opt.applyConfig(cfg)
				if err != nil {
					return err
				}
			}
			return nil
		},
		containerOption: func(ctr *container) error {
			for _, opt := range opts {
				err := opt.applyContainer(ctr)
				if err != nil {
					return err
				}
			}
			return nil
		},
	}
}

type configOption func(*config) error

func (c configOption) applyConfig(cfg *config) error {
	return c(cfg)
}

func (c configOption) applyContainer(*container) error {
	return nil
}

type containerOption func(*container) error

func (c containerOption) applyConfig(*config) error {
	return nil
}

func (c containerOption) applyContainer(ctr *container) error {
	return c(ctr)
}

type option struct {
	configOption
	containerOption
}

func (o option) applyConfig(c *config) error {
	return o.configOption(c)
}

func (o option) applyContainer(c *container) error {
	return o.containerOption(c)
}

var _, _, _ Option = (*configOption)(nil), (*containerOption)(nil), option{}
