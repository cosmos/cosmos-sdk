package module

import (
	"fmt"
	"reflect"
)

// Container is a low-level dependency injection container which manages dependencies
// based on scopes and security policies. All providers can be run in a scope which
// may provide certain dependencies specifically for that scope or provide/deny access
// to dependencies based on that scope.
type Container struct {
	providers      map[Key]*node
	scopeProviders map[Key]*scopeNode
	nodes          []*node
	scopeNodes     []*scopeNode

	values       map[Key]secureValue
	scopedValues map[Scope]map[Key]reflect.Value
}

func NewContainer() *Container {
	return &Container{
		providers:      map[Key]*node{},
		scopeProviders: map[Key]*scopeNode{},
		nodes:          nil,
		scopeNodes:     nil,
		values:         map[Key]secureValue{},
		scopedValues:   map[Scope]map[Key]reflect.Value{},
	}
}

type Key struct {
	Type reflect.Type
}

type Scope = string

type node struct {
	Provider
	called bool
	values []reflect.Value
	err    error
}

// Provider is a general dependency provider. Its scope parameter is used
// to receive scoped dependencies and gain access to general dependencies within
// its security policy. Access to dependencies provided by this provider can optionally
// be restricted to certain scopes based on SecurityCheckers.
type Provider struct {
	Constructor      func(deps []reflect.Value) ([]reflect.Value, error)
	Needs            []Key
	Provides         []Key
	Scope            Scope
	SecurityCheckers []SecurityChecker
}

type scopeNode struct {
	ScopedProvider
	calledForScope map[Scope]bool
	valuesForScope map[Scope][]reflect.Value
	errsForScope   map[Scope]error
}

// ScopedProvider provides scoped dependencies. Its constructor function will provide
// dependencies specific to the scope parameter. Instead of providing general dependencies
// with restricted access based on security checkers, ScopedProvider provides potentially different
// dependency instances to different scopes. It is assumed that a scoped provider
// can provide a dependency for any valid scope passed to it, although it can return an error
// to deny access.
type ScopedProvider struct {
	Constructor func(scope Scope, deps []reflect.Value) ([]reflect.Value, error)
	Needs       []Key
	Provides    []Key
	Scope       Scope
}

type secureValue struct {
	value           reflect.Value
	securityChecker SecurityChecker
}

type SecurityChecker func(scope Scope) error

func (c *Container) Provide(provider Provider) error {
	n := &node{
		Provider: provider,
		called:   false,
	}

	c.nodes = append(c.nodes, n)

	for _, key := range provider.Provides {
		if c.providers[key] != nil {
			return fmt.Errorf("TODO")
		}

		c.providers[key] = n
	}

	return nil
}

func (c *Container) ProvideScoped(provider ScopedProvider) error {
	n := &scopeNode{
		ScopedProvider: provider,
		calledForScope: map[Scope]bool{},
		valuesForScope: map[Scope][]reflect.Value{},
		errsForScope:   map[Scope]error{},
	}

	c.scopeNodes = append(c.scopeNodes, n)

	for _, key := range provider.Provides {
		if c.scopeProviders[key] != nil {
			return fmt.Errorf("TODO")
		}

		c.scopeProviders[key] = n
	}

	return nil
}

func (c *Container) resolve(scope Scope, key Key, stack map[interface{}]bool) (reflect.Value, error) {
	if scope != "" {
		if val, ok := c.scopedValues[scope][key]; ok {
			return val, nil
		}

		if provider, ok := c.scopeProviders[key]; ok {
			if stack[provider] {
				return reflect.Value{}, fmt.Errorf("fatal: cycle detected")
			}

			if provider.calledForScope[scope] {
				return reflect.Value{}, fmt.Errorf("error: %v", provider.errsForScope[scope])
			}

			var deps []reflect.Value
			for _, need := range provider.Needs {
				stack[provider] = true
				res, err := c.resolve(provider.Scope, need, stack)
				delete(stack, provider)

				if err != nil {
					return reflect.Value{}, err
				}

				deps = append(deps, res)
			}

			res, err := provider.Constructor(scope, deps)
			provider.calledForScope[scope] = true
			if err != nil {
				provider.errsForScope[scope] = err
				return reflect.Value{}, err
			}

			provider.valuesForScope[scope] = res

			for i, val := range res {
				p := provider.Provides[i]
				if _, ok := c.scopedValues[scope][p]; ok {
					return reflect.Value{}, fmt.Errorf("value provided twice")
				}

				if c.scopedValues[scope] == nil {
					c.scopedValues[scope] = map[Key]reflect.Value{}
				}
				c.scopedValues[scope][p] = val
			}

			val, ok := c.scopedValues[scope][key]
			if !ok {
				return reflect.Value{}, fmt.Errorf("internal error: bug")
			}

			return val, nil
		}
	}

	if val, ok := c.values[key]; ok {
		if val.securityChecker != nil {
			if err := val.securityChecker(scope); err != nil {
				return reflect.Value{}, err
			}
		}

		return val.value, nil
	}

	if provider, ok := c.providers[key]; ok {
		if stack[provider] {
			return reflect.Value{}, fmt.Errorf("fatal: cycle detected")
		}

		if provider.called {
			return reflect.Value{}, fmt.Errorf("error: %v", provider.err)
		}

		err := c.execNode(provider, stack)
		if err != nil {
			return reflect.Value{}, err
		}

		val, ok := c.values[key]
		if !ok {
			return reflect.Value{}, fmt.Errorf("internal error: bug")
		}

		if val.securityChecker != nil {
			if err := val.securityChecker(scope); err != nil {
				return reflect.Value{}, err
			}
		}

		return val.value, nil
	}

	return reflect.Value{}, fmt.Errorf("no provider")
}

func (c *Container) execNode(provider *node, stack map[interface{}]bool) error {
	if provider.called {
		return provider.err
	}

	var deps []reflect.Value
	for _, need := range provider.Needs {
		stack[provider] = true
		res, err := c.resolve(provider.Scope, need, stack)
		delete(stack, provider)

		if err != nil {
			return err
		}

		deps = append(deps, res)
	}

	res, err := provider.Constructor(deps)
	provider.called = true
	if err != nil {
		provider.err = err
		return err
	}

	provider.values = res

	for i, val := range res {
		p := provider.Provides[i]
		if _, ok := c.values[p]; ok {
			return fmt.Errorf("value provided twice")
		}

		var secChecker SecurityChecker
		if i < len(provider.SecurityCheckers) {
			secChecker = provider.SecurityCheckers[i]
		}

		c.values[p] = secureValue{
			value:           val,
			securityChecker: secChecker,
		}
	}

	return nil
}

func (c *Container) Resolve(scope Scope, key Key) (reflect.Value, error) {
	val, err := c.resolve(scope, key, map[interface{}]bool{})
	if err != nil {
		return reflect.Value{}, err
	}
	return val, nil
}

// InitializeAll attempts to call all providers instantiating the dependencies they provide
func (c *Container) InitializeAll() error {
	for _, node := range c.nodes {
		err := c.execNode(node, map[interface{}]bool{})
		if err != nil {
			return err
		}
	}
	return nil
}
