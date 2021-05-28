package module

import (
	"fmt"
)

type container struct {
	providers      map[Key]*node
	scopeProviders map[Key]*scopeNode
	nodes          []*node
	scopeNodes     []*scopeNode

	values       map[Key]interface{}
	scopedValues map[string]map[Key]interface{}
}

type Key struct {
	Type interface{}
}

type Value struct {
	Key   Key
	Value interface{}
}

type node struct {
	Provider
	called bool
	values []Value
	err    error
}

type Provider struct {
	Constructor func(deps []interface{}) ([]interface{}, error)
	Needs       []Key
	Provides    []Key
	Scope       string
}

type scopeNode struct {
	ScopedProvider
	calledForScope map[string]bool
	valuesForScope map[string][]Value
	errsForScope   map[string]error
}

type ScopedProvider struct {
	Constructor func(scope string, deps []interface{}) ([]interface{}, error)
	Needs       []Key
	Provides    []Key
}

func (c *container) Provide(provider Provider) error {
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

func (c *container) ProvideForScope(provider ScopedProvider) error {
	n := &scopeNode{
		ScopedProvider: provider,
		calledForScope: map[string]bool{},
		valuesForScope: map[string][]Value{},
		errsForScope:   map[string]error{},
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

func (c *container) resolve(scope string, key Key, stack map[interface{}]bool) (interface{}, error) {
	if scope != "" {
		if val, ok := c.scopedValues[scope][key]; ok {
			return val, nil
		}

		if provider, ok := c.scopeProviders[key]; ok {
			if stack[provider] {
				return nil, fmt.Errorf("fatal: cycle detected")
			}

			if provider.calledForScope[scope] {
				return nil, fmt.Errorf("error: %v", provider.errsForScope[scope])
			}

			var deps []interface{}
			for _, need := range provider.Needs {
				stack[provider] = true
				res, err := c.resolve(scope, need, stack)
				delete(stack, provider)

				if err != nil {
					return nil, err
				}

				deps = append(deps, res)
			}

			res, err := provider.Constructor(scope, deps)
			provider.calledForScope[scope] = true
			if err != nil {
				provider.errsForScope[scope] = err
				return nil, err
			}

			for i, val := range res {
				p := provider.Provides[i]
				if _, ok := c.scopedValues[scope][p]; ok {
					return nil, fmt.Errorf("value provided twice")
				}

				if c.scopedValues[scope] == nil {
					c.scopedValues[scope] = map[Key]interface{}{}
				}
				c.scopedValues[scope][p] = val
			}

			val, ok := c.scopedValues[scope][key]
			if !ok {
				return nil, fmt.Errorf("internal error: bug")
			}

			return val, nil
		}
	}

	if val, ok := c.values[key]; ok {
		return val, nil
	}

	if provider, ok := c.providers[key]; ok {
		if stack[provider] {
			return nil, fmt.Errorf("fatal: cycle detected")
		}

		if provider.called {
			return nil, fmt.Errorf("error: %v", provider.err)
		}

		var deps []interface{}
		for _, need := range provider.Needs {
			stack[provider] = true
			res, err := c.resolve(provider.Scope, need, stack)
			delete(stack, provider)

			if err != nil {
				return nil, err
			}

			deps = append(deps, res)
		}

		res, err := provider.Constructor(deps)
		provider.called = true
		if err != nil {
			provider.err = err
			return nil, err
		}

		for i, val := range res {
			p := provider.Provides[i]
			if _, ok := c.values[p]; ok {
				return nil, fmt.Errorf("value provided twice")
			}

			c.values[p] = val
		}

		val, ok := c.values[key]
		if !ok {
			return nil, fmt.Errorf("internal error: bug")
		}

		return val, nil
	}

	return nil, fmt.Errorf("no provider")
}

func (c *container) Resolve(scope string, key Key) (interface{}, error) {
	return c.resolve(scope, key, map[interface{}]bool{})
}

type Container interface {
	Provide(provider Provider) error
	ProvideForScope(provider ScopedProvider) error
	Resolve(scope string, key Key) (interface{}, error)
}

func NewContainer() Container {
	return &container{
		providers:      map[Key]*node{},
		scopeProviders: map[Key]*scopeNode{},
		nodes:          nil,
		scopeNodes:     nil,
		values:         map[Key]interface{}{},
		scopedValues:   map[string]map[Key]interface{}{},
	}
}
