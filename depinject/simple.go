package depinject

import (
	"reflect"

	"github.com/cosmos/cosmos-sdk/depinject/internal/graphviz"
)

type simpleProvider struct {
	provider   *ProviderDescriptor
	called     bool
	values     []reflect.Value
	valueExprs []expr
	moduleKey  *moduleKey
}

type simpleResolver struct {
	node        *simpleProvider
	idxInValues int
	resolved    bool
	typ         reflect.Type
	value       reflect.Value
	valueExpr   expr
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
		values, eCall, err := ctr.call(s.provider, s.moduleKey)
		if err != nil {
			return nil, err
		}
		s.values = values
		s.called = true

		// codegen
		varsDef, valueExprs := s.provider.codegenOutputs(ctr, "")
		s.valueExprs = valueExprs
		ctr.codegenWriteln(varsDef, eCall)
		s.provider.codegenErrCheck(ctr)
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
		s.valueExpr = s.node.valueExprs[s.idxInValues]
		s.resolved = true
	}

	return s.value, s.valueExpr, nil
}

func (s simpleResolver) addNode(p *simpleProvider, _ int) error {
	return duplicateDefinitionError(s.typ, p.provider.Location, s.node.provider.Location.String())
}

func (s simpleResolver) typeGraphNode() *graphviz.Node {
	return s.graphNode
}
