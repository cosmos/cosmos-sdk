package container

import (
	"reflect"
)

type simpleProvider struct {
	ctr    *ProviderDescriptor
	called bool
	values []reflect.Value
	scope  Scope
}

type simpleResolver struct {
	node        *simpleProvider
	idxInValues int
	resolved    bool
	typ         reflect.Type
	value       reflect.Value
}

func (s *simpleResolver) describeLocation() string {
	return s.node.ctr.Location.String()
}

func (s *simpleProvider) resolveValues(ctr *container) ([]reflect.Value, error) {
	if !s.called {
		values, err := ctr.call(s.ctr, s.scope)
		if err != nil {
			return nil, err
		}
		s.values = values
		s.called = true
	}

	return s.values, nil
}

func (s *simpleResolver) resolve(c *container, _ Scope, caller Location) (reflect.Value, error) {
	// Log
	c.logf("Providing %v from %s to %s", s.typ, s.node.ctr.Location, caller.Name())

	// Resolve
	if !s.resolved {
		values, err := s.node.resolveValues(c)
		if err != nil {
			return reflect.Value{}, err
		}

		value := values[s.idxInValues]
		s.value = value
		s.resolved = true
	}

	return s.value, nil
}

func (s simpleResolver) addNode(p *simpleProvider, _ int, _ *container) error {
	return duplicateDefinitionError(s.typ, p.ctr.Location, s.node.ctr.Location.String())
}
