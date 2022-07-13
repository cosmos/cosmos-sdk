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

// OnePerModuleType marks a type which
// can have up to one value per module. All of the values for a one-per-module type T
// and their respective modules, can be retrieved by declaring an input parameter map[string]T.
type OnePerModuleType interface {
	// IsOnePerModuleType is a marker function just indicates that this is a one-per-module type.
	IsOnePerModuleType()
}

var onePerModuleTypeType = reflect.TypeOf((*OnePerModuleType)(nil)).Elem()

func isOnePerModuleType(t reflect.Type) bool {
	return t.Implements(onePerModuleTypeType)
}

func isOnePerModuleMapType(typ reflect.Type) bool {
	return typ.Kind() == reflect.Map && isOnePerModuleType(typ.Elem()) && typ.Key().Kind() == reflect.String
}

type onePerModuleResolver struct {
	typ        reflect.Type
	mapType    reflect.Type
	providers  map[*moduleKey]*simpleProvider
	idxMap     map[*moduleKey]int
	resolved   bool
	values     reflect.Value
	valueIdent *ast.Ident
	graphNode  *graphviz.Node
}

func (o *onePerModuleResolver) getType() reflect.Type {
	return o.mapType
}

type mapOfOnePerModuleResolver struct {
	*onePerModuleResolver
}

func (o *onePerModuleResolver) resolve(_ *container, _ *moduleKey, _ Location) (reflect.Value, ast.Expr, error) {
	return reflect.Value{}, nil, errors.Errorf("%v is a one-per-module type and thus can't be used as an input parameter, instead use %v", o.typ, o.mapType)
}

func (o *onePerModuleResolver) describeLocation() string {
	return fmt.Sprintf("one-per-module type %v", o.typ)
}

func (o *mapOfOnePerModuleResolver) resolve(c *container, _ *moduleKey, caller Location) (reflect.Value, ast.Expr, error) {
	// Log
	c.logf("Providing one-per-module type map %v to %s from:", o.mapType, caller.Name())
	c.indentLogger()
	for key, node := range o.providers {
		c.logf("%s: %s", key.name, node.provider.Location)
	}
	c.dedentLogger()

	// Resolve
	if !o.resolved {
		res := reflect.MakeMap(o.mapType)
		var elemExprs []ast.Expr
		for key, node := range o.providers {
			values, err := node.resolveValues(c, false)
			if err != nil {
				return reflect.Value{}, nil, err
			}
			idx := o.idxMap[key]
			if len(values) <= idx {
				return reflect.Value{}, nil, errors.Errorf("expected value of type %T at index %d", o.typ, idx)
			}
			value := values[idx]
			res.SetMapIndex(reflect.ValueOf(key.name), value)

			// codegen
			valueExpr := node.valueExprs[idx]
			elemExprs = append(elemExprs, &ast.KeyValueExpr{
				Key:   &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", key.name)},
				Value: valueExpr,
			})
		}

		o.values = res
		o.resolved = true

		// codegen
		o.valueIdent = c.funcGen.CreateIdent(util.StringFirstLower(fmt.Sprintf("%sMap", o.typ.Name())))
		c.codegenStmt(&ast.AssignStmt{
			Lhs: []ast.Expr{o.valueIdent},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{
				&ast.CompositeLit{Type: &ast.MapType{
					Key:   ast.NewIdent("string"),
					Value: ast.NewIdent(o.typ.Name()),
				},
					Elts: elemExprs,
				},
			},
		})
	}

	return o.values, o.valueIdent, nil
}

func (o *onePerModuleResolver) addNode(n *simpleProvider, i int) error {
	if n.moduleKey == nil {
		return errors.Errorf("cannot define a provider with one-per-module dependency %v which isn't provided in a module", o.typ)
	}

	if existing, ok := o.providers[n.moduleKey]; ok {
		return errors.Errorf("duplicate provision for one-per-module type %v in module %s: %s\n\talready provided by %s",
			o.typ, n.moduleKey.name, n.provider.Location, existing.provider.Location)
	}

	o.providers[n.moduleKey] = n
	o.idxMap[n.moduleKey] = i

	return nil
}

func (o *mapOfOnePerModuleResolver) addNode(s *simpleProvider, _ int) error {
	return errors.Errorf("%v is a one-per-module type and thus %v can't be used as an output parameter in %s", o.typ, o.mapType, s.provider.Location)
}

func (o onePerModuleResolver) typeGraphNode() *graphviz.Node {
	return o.graphNode
}
