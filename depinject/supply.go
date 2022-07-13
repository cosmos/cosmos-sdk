package depinject

import (
	"go/ast"
	"go/token"
	"reflect"

	"github.com/cosmos/cosmos-sdk/depinject/internal/graphviz"
)

type supplyResolver struct {
	typ        reflect.Type
	value      reflect.Value
	loc        Location
	graphNode  *graphviz.Node
	varIdent   *ast.Ident
	codegenDef bool
}

func (s supplyResolver) getType() reflect.Type {
	return s.typ
}

func (s supplyResolver) describeLocation() string {
	return s.loc.String()
}

func (s supplyResolver) addNode(provider *simpleProvider, _ int) error {
	return duplicateDefinitionError(s.typ, provider.provider.Location, s.loc.String())
}

func (s supplyResolver) resolve(c *container, _ *moduleKey, caller Location) (reflect.Value, ast.Expr, error) {
	c.logf("Supplying %v from %s to %s", s.typ, s.loc, caller.Name())
	if !s.codegenDef {
		e, err := c.funcGen.ValueExpr(s.value)
		if err != nil {
			return reflect.Value{}, nil, err
		}

		c.codegenStmt(&ast.AssignStmt{
			Lhs: []ast.Expr{s.varIdent},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{e},
		})
		s.codegenDef = true
	}
	return s.value, s.varIdent, nil
}

func (s supplyResolver) typeGraphNode() *graphviz.Node {
	return s.graphNode
}
