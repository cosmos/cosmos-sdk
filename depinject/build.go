package depinject

import (
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"os"
	"reflect"

	"github.com/pkg/errors"
)

func (c *container) build(loc Location, outputs ...interface{}) error {
	moduleKeyContextIdent, _ := c.getOrCreateIdent("moduleKeyContext", c.moduleKeyContext)
	c.codegenStmt(&ast.AssignStmt{
		Lhs: []ast.Expr{moduleKeyContextIdent},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{
			&ast.UnaryExpr{
				Op: token.AND,
				X: &ast.CompositeLit{
					Type: ast.NewIdent("depinject.ModuleKeyContext"),
				},
			},
		},
	})

	var providerIn []ProviderInput
	for _, output := range outputs {
		typ := reflect.TypeOf(output)
		if typ.Kind() != reflect.Pointer {
			return fmt.Errorf("output type must be a pointer, %s is invalid", typ)
		}

		providerIn = append(providerIn, ProviderInput{Type: typ.Elem()})

		// codegen
		astTyp, err := c.typeExpr(typ.Elem())
		if err != nil {
			return err
		}
		c.codegenFunc.Type.Results.List = append(c.codegenFunc.Type.Results.List, &ast.Field{Type: astTyp})

		defaultValue, err := c.valueExpr(reflect.Zero(typ.Elem()))
		if err != nil {
			return err
		}

		c.codegenErrReturn.Results = append(c.codegenErrReturn.Results, defaultValue)
	}
	c.codegenFunc.Type.Results.List = append(c.codegenFunc.Type.Results.List, &ast.Field{Type: ast.NewIdent("error")})
	c.codegenErrReturn.Results = append(c.codegenErrReturn.Results, ast.NewIdent("err"))

	desc := ProviderDescriptor{
		Inputs:  providerIn,
		Outputs: nil,
		Fn: func(values []reflect.Value) ([]reflect.Value, error) {
			if len(values) != len(outputs) {
				return nil, fmt.Errorf("internal error, unexpected number of values")
			}

			for i, output := range outputs {
				val := reflect.ValueOf(output)
				val.Elem().Set(values[i])
			}

			return nil, nil
		},
		Location: loc,
	}
	callerGraphNode := c.locationGraphNode(loc, nil)
	callerGraphNode.SetShape("hexagon")

	desc, err := expandStructArgsProvider(desc)
	if err != nil {
		return err
	}

	c.logf("Registering outputs")
	c.indentLogger()

	node, err := c.addNode(&desc, nil)
	if err != nil {
		return err
	}

	c.dedentLogger()

	sn, ok := node.(*simpleProvider)
	if !ok {
		return errors.Errorf("cannot run module-scoped provider as an invoker")
	}

	c.logf("Building container")
	_, err = sn.resolveValues(c, true)
	if err != nil {
		return err
	}
	c.logf("Done building container")
	c.logf("Calling invokers")
	for _, inv := range c.invokers {
		_, eCall, err := c.call(inv.fn, inv.modKey)
		if err != nil {
			return err
		}

		// codegen
		_, _ = inv.fn.codegenOutputs(c, "")
		c.codegenStmt(&ast.ExprStmt{X: eCall})
		inv.fn.codegenErrCheck(c) // TODO: deal with err already being defined and use token.ASSIGN
	}
	c.logf("Done calling invokers")

	fset := token.NewFileSet()
	//ast.Print(fset, c.codegenBody)
	fmt.Println("Codegen:")
	printer.Fprint(os.Stdout, fset, c.codegenFunc)
	fmt.Println()

	return nil
}
