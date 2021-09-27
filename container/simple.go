package container

import (
	"reflect"

	"github.com/goccy/go-graphviz/cgraph"
)

type simpleProvider struct {
	provider *ProviderDescriptor
	called   bool
	values   []reflect.Value
	scope    Scope
}

type simpleResolver struct {
	node        *simpleProvider
	idxInValues int
	resolved    bool
	typ         reflect.Type
	value       reflect.Value
	graphNode   *cgraph.Node
}

func (s *simpleResolver) describeLocation() string {
	return s.node.provider.Location.String()
}

func (s *simpleProvider) resolveValues(ctr *container) ([]reflect.Value, error) {
	if !s.called {
		values, err := ctr.call(s.provider, s.scope)
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
	c.logf("Providing %v from %s to %s", s.typ, s.node.provider.Location, caller.Name())

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

func (s simpleResolver) addNode(p *simpleProvider, _ int) error {
	return duplicateDefinitionError(s.typ, p.provider.Location, s.node.provider.Location.String())
}

func (s simpleResolver) typeGraphNode() *cgraph.Node {
	return s.graphNode
}
