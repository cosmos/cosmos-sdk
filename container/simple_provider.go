package container

import (
	"fmt"
	"reflect"

	containerreflect "github.com/cosmos/cosmos-sdk/container/reflect"
)

type simpleProvider struct {
	ctr    *containerreflect.Constructor
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

func (s *simpleResolver) resolve(c *container, _ Scope, caller containerreflect.Location) (reflect.Value, error) {
	c.logf("Providing %v from %s to %s", s.typ, s.node.ctr.Location, caller.Name())
	err := c.addGraphEdge(s.node.ctr.Location, caller)
	if err != nil {
		return reflect.Value{}, err
	}

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

func (s simpleResolver) addNode(*simpleProvider, int) error {
	return fmt.Errorf("duplicate constructor for type %v", s.typ)
}
