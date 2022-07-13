package depinject

import (
	"fmt"
	"go/ast"
	"reflect"

	"github.com/pkg/errors"
)

func (c *container) call(provider *ProviderDescriptor, key *moduleKey) ([]reflect.Value, *ast.CallExpr, error) {
	loc := provider.Location
	graphNode := c.locationGraphNode(loc, key)

	markGraphNodeAsFailed(graphNode)

	if c.callerMap[loc] {
		return nil, nil, errors.Errorf("cyclic dependency: %s -> %s", loc.Name(), loc.Name())
	}

	c.callerMap[loc] = true
	c.callerStack = append(c.callerStack, loc)

	inVals, argExprs, err := c.resolveCallInputs(provider, key)
	if err != nil {
		return nil, nil, err
	}
	c.logf("Calling %s", loc)

	delete(c.callerMap, loc)
	c.callerStack = c.callerStack[0 : len(c.callerStack)-1]

	out, err := provider.Fn(inVals)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "error calling provider %s", loc)
	}

	markGraphNodeAsUsed(graphNode)

	importPrefix := c.funcGen.AddOrGetImport(loc.PkgPath())
	name := loc.ShortName()
	if importPrefix != "" {
		name = fmt.Sprintf("%s.%s", importPrefix, name)
	}
	e := &ast.CallExpr{Fun: ast.NewIdent(name), Args: argExprs}
	return out, e, nil
}

func (c *container) resolveCallInputs(provider *ProviderDescriptor, key *moduleKey) ([]reflect.Value, []ast.Expr, error) {
	loc := provider.Location
	c.logf("Resolving dependencies for %s", loc)
	c.indentLogger()

	var argExprs []ast.Expr
	inVals := make([]reflect.Value, len(provider.Inputs))
	var curStructLit *ast.CompositeLit
	for i, in := range provider.Inputs {
		val, e, err := c.resolve(in, key, loc)
		if err != nil {
			return nil, nil, err
		}
		inVals[i] = val

		if in.structFieldName != "" {
			if in.startStructType != nil {
				typExpr, err := c.funcGen.TypeExpr(in.startStructType)
				if err != nil {
					return nil, nil, err
				}
				curStructLit = &ast.CompositeLit{
					Type: typExpr,
				}
				argExprs = append(argExprs, curStructLit)
			}

			curStructLit.Elts = append(curStructLit.Elts, &ast.KeyValueExpr{
				Key:   ast.NewIdent(in.structFieldName),
				Value: e,
			})
		} else {
			argExprs = append(argExprs, e)
		}
	}

	c.dedentLogger()
	return inVals, argExprs, nil
}
