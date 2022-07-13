package depinject

import (
	"fmt"
	"go/ast"
	"go/token"
	"reflect"

	"cosmossdk.io/depinject/internal/graphviz"
	"cosmossdk.io/depinject/internal/util"
)

type moduleDepProvider struct {
	provider        *ProviderDescriptor
	calledForModule map[*moduleKey]bool
	valueMap        map[*moduleKey][]reflect.Value
	valueExprs      map[*moduleKey][]ast.Expr
}

type moduleDepResolver struct {
	typ         reflect.Type
	idxInValues int
	node        *moduleDepProvider
	valueMap    map[*moduleKey]reflect.Value
	graphNode   *graphviz.Node
}

func (s moduleDepResolver) getType() reflect.Type {
	return s.typ
}

func (s moduleDepResolver) describeLocation() string {
	return s.node.provider.Location.String()
}

func (s moduleDepResolver) resolve(ctr *container, moduleKey *moduleKey, caller Location) (reflect.Value, ast.Expr, error) {
	// Log
	ctr.logf("Providing %v from %s to %s", s.typ, s.node.provider.Location, caller.Name())

	// Resolve
	if val, ok := s.valueMap[moduleKey]; ok {
		return val, nil, nil
	}

	if !s.node.calledForModule[moduleKey] {
		values, eCall, err := ctr.call(s.node.provider, moduleKey)
		if err != nil {
			return reflect.Value{}, nil, err
		}

		s.node.valueMap[moduleKey] = values
		s.node.calledForModule[moduleKey] = true

		// codegen
		varRefs, valueExprs := s.node.provider.codegenOutputs(ctr, fmt.Sprintf(
			"For%s",
			util.StringFirstUpper(moduleKey.name),
		))
		s.node.valueExprs[moduleKey] = valueExprs
		ctr.codegenStmt(
			&ast.AssignStmt{
				Lhs: varRefs,
				Tok: token.DEFINE,
				Rhs: []ast.Expr{eCall},
			},
		)
		s.node.provider.codegenErrCheck(ctr)
	}

	value := s.node.valueMap[moduleKey][s.idxInValues]
	s.valueMap[moduleKey] = value
	return value, s.node.valueExprs[moduleKey][s.idxInValues], nil
}

func (s moduleDepResolver) addNode(p *simpleProvider, _ int) error {
	return duplicateDefinitionError(s.typ, p.provider.Location, s.node.provider.Location.String())
}

func (s moduleDepResolver) typeGraphNode() *graphviz.Node {
	return s.graphNode
}
