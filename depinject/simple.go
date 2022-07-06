package depinject

import (
	"reflect"

	"github.com/cosmos/cosmos-sdk/depinject/internal/graphviz"
)

type simpleProvider struct {
	provider  *ProviderDescriptor
	called    bool
	values    []reflect.Value
	valueVars []varRef
	moduleKey *moduleKey
}

type simpleResolver struct {
	node        *simpleProvider
	idxInValues int
	resolved    bool
	typ         reflect.Type
	value       reflect.Value
	graphNode   *graphviz.Node
}

func (s *simpleResolver) getType() reflect.Type {
	return s.typ
}

func (s *simpleResolver) describeLocation() string {
	return s.node.provider.Location.String()
}

func (s *simpleProvider) resolveValues(ctr *container) ([]reflect.Value, error) {
	if !s.called {
		var valueVars []varRef
		for _, output := range s.provider.Outputs {
			v := ctr.createVar(output.Type.Name())
			valueVars = append(valueVars, v)
		}

		values, err := ctr.call(s.provider, s.moduleKey)
		if err != nil {
			return nil, err
		}

		s.valueVars = valueVars
		s.values = values
		s.called = true
		first := true
		for _, valueVar := range valueVars {
			if !first {
				ctr.codegenWrite(", ")
			}
			ctr.codegenWrite(valueVar)
			first = false
		}
		ctr.codegenWrite(" := ")
	}

	return s.values, nil
}

func (s *simpleResolver) resolve(c *container, _ *moduleKey, caller Location) (reflect.Value, expr, error) {
	// Log
	c.logf("Providing %v from %s to %s", s.typ, s.node.provider.Location, caller.Name())

	// Resolve
	if !s.resolved {
		values, err := s.node.resolveValues(c)
		if err != nil {
			return reflect.Value{}, nil, err
		}

		value := values[s.idxInValues]
		s.value = value
		s.resolved = true
	}

	return s.value, nil, nil
}

func (s simpleResolver) addNode(p *simpleProvider, _ int) error {
	return duplicateDefinitionError(s.typ, p.provider.Location, s.node.provider.Location.String())
}

func (s simpleResolver) typeGraphNode() *graphviz.Node {
	return s.graphNode
}
