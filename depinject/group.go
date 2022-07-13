package depinject

import (
	"fmt"
	"go/ast"
	"go/token"
	"reflect"

	"github.com/pkg/errors"

	"github.com/cosmos/cosmos-sdk/depinject/internal/graphviz"
	"github.com/cosmos/cosmos-sdk/depinject/internal/util"
)

// ManyPerContainerType marks a type which automatically gets grouped together. For an ManyPerContainerType T,
// T and []T can be declared as output parameters for providers as many times within the container
// as desired. All of the provided values for T can be retrieved by declaring an
// []T input parameter.
type ManyPerContainerType interface {
	// IsManyPerContainerType is a marker function which just indicates that this is a many-per-container type.
	IsManyPerContainerType()
}

var manyPerContainerTypeType = reflect.TypeOf((*ManyPerContainerType)(nil)).Elem()

func isManyPerContainerType(t reflect.Type) bool {
	return t.Implements(manyPerContainerTypeType)
}

func isManyPerContainerSliceType(typ reflect.Type) bool {
	return typ.Kind() == reflect.Slice && isManyPerContainerType(typ.Elem())
}

type groupResolver struct {
	typ          reflect.Type
	sliceType    reflect.Type
	idxsInValues []int
	providers    []*simpleProvider
	resolved     bool
	values       reflect.Value
	graphNode    *graphviz.Node
}

func (g *groupResolver) getType() reflect.Type {
	return g.sliceType
}

type sliceGroupResolver struct {
	*groupResolver
	valueIdent *ast.Ident
}

func (g *groupResolver) describeLocation() string {
	return fmt.Sprintf("many-per-container type %v", g.typ)
}

func (g *sliceGroupResolver) resolve(c *container, _ *moduleKey, caller Location) (reflect.Value, ast.Expr, error) {
	// Log
	c.logf("Providing many-per-container type slice %v to %s from:", g.sliceType, caller.Name())
	c.indentLogger()
	for _, node := range g.providers {
		c.logf(node.provider.Location.String())
	}
	c.dedentLogger()

	// Resolve
	if !g.resolved {
		res := reflect.MakeSlice(g.sliceType, 0, 0)
		var simpleExprs []ast.Expr
		var sliceExprs []ast.Expr
		for i, node := range g.providers {
			values, err := node.resolveValues(c, false)
			if err != nil {
				return reflect.Value{}, nil, err
			}
			idx := g.idxsInValues[i]
			value := values[idx]
			valueExpr := node.valueExprs[idx]
			if value.Kind() == reflect.Slice {
				n := value.Len()
				for j := 0; j < n; j++ {
					res = reflect.Append(res, value.Index(j))
				}
				sliceExprs = append(sliceExprs, valueExpr)
			} else {
				res = reflect.Append(res, value)
				simpleExprs = append(simpleExprs, valueExpr)
			}
		}
		g.values = res
		g.resolved = true

		// codegen
		g.valueIdent = c.funcGen.CreateIdent(util.StringFirstLower(fmt.Sprintf("%sSlice", g.typ.Name())))
		c.codegenStmt(&ast.AssignStmt{
			Lhs: []ast.Expr{g.valueIdent},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{
				&ast.CompositeLit{Type: &ast.ArrayType{
					Elt: ast.NewIdent(g.typ.Name()),
				},
					Elts: simpleExprs,
				},
			},
		})

		for _, expr := range sliceExprs {
			c.codegenStmt(&ast.AssignStmt{
				Lhs: []ast.Expr{g.valueIdent},
				Tok: token.ASSIGN,
				Rhs: []ast.Expr{&ast.CallExpr{
					Fun:      ast.NewIdent("append"),
					Args:     []ast.Expr{g.valueIdent, expr},
					Ellipsis: token.Pos(1),
				}},
			})
		}
	}

	return g.values, g.valueIdent, nil
}

func (g *groupResolver) resolve(_ *container, _ *moduleKey, _ Location) (reflect.Value, ast.Expr, error) {
	return reflect.Value{}, nil, errors.Errorf("%v is an many-per-container type and cannot be used as an input value, instead use %v", g.typ, g.sliceType)
}

func (g *groupResolver) addNode(n *simpleProvider, i int) error {
	g.providers = append(g.providers, n)
	g.idxsInValues = append(g.idxsInValues, i)
	return nil
}

func (g groupResolver) typeGraphNode() *graphviz.Node {
	return g.graphNode
}
