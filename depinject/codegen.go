package depinject

import (
	"fmt"
	"go/ast"
	"go/parser"
	"os"
	"path"
	"strings"

	"cosmossdk.io/depinject/internal/codegen"
)

// Codegen tells depinject that the calling function is a template function for generating code. Template
// functions must follow a very specific pattern or there will be an error:
// - function inputs must be supplied dependencies
// - function outputs must be resolved dependencies
// - all functions must have error as the last parameter
// - the function should have two statements: a call to depinject.InjectDebug and return
// - the call to depinject.InjectDebug must have either:
//   - the error assigned to err
//   - a single static variable config parameter, or
//   - a call to depinject.Configs which has
//     - one or more static variable config parameters
//     - zero or one calls to depinject.Supply with all input params passed as args
//
// Ex:
//  func Build(x SomeRuntimeDep) (y SomeResolvedDep, error) {
//    err = depinject.InjectDebug(
//      depinject.Codegen(),
//      depinject.Configs(
//        myStaticConfig,
//        depinject.Supply(x),
//      ),
//      &y
//    )
//    return
//  }
func Codegen() DebugOption {
	loc := LocationFromCaller(1).(*location)
	return debugOption(func(config *debugConfig) error {
		config.codegenLoc = loc

		f, err := parser.ParseFile(config.fset, loc.file, nil, parser.ParseComments|parser.AllErrors)
		if err != nil {
			return err
		}

		err = config.startCodegen(f)
		if err != nil {
			return err
		}

		outFilename := loc.file
		ext := path.Ext(outFilename)
		outFilename = outFilename[0:len(outFilename)-len(ext)] + ".depinject.go"
		fmt.Printf("codegen output to %s\n", outFilename)
		outFile, err := os.OpenFile(outFilename, os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0644)
		if err != nil {
			return err
		}
		config.codegenOut = outFile

		return nil
	})
}

func (c *debugConfig) startCodegen(file *ast.File) error {
	fileGen, err := codegen.NewFileGen(file, c.codegenLoc.pkg)
	if err != nil {
		return err
	}

	funcGen := fileGen.PatchFuncDecl(c.codegenLoc.name)
	if funcGen == nil {
		return fmt.Errorf("couldn't resolve function %s in %s", c.codegenLoc.name, c.codegenLoc.file)
	}

	funcParamNames, err := c.checkAndPatchFuncDecl(funcGen)
	if err != nil {
		return err
	}

	c.funcParamNames = funcParamNames

	if len(fileGen.File.Comments) == 0 ||
		len(fileGen.File.Comments[0].List) == 0 ||
		strings.TrimSpace(fileGen.File.Comments[0].List[0].Text) != "//go:build depinject" {
		return c.astError(fileGen.File, `expected comment: //go:build depinject`)
	}

	fileGen.File.Comments[0] = &ast.CommentGroup{
		List: []*ast.Comment{
			{
				Text: "//go:build !depinject\n",
			},
		},
	}

	c.funcGen = funcGen
	return nil
}

func (c *debugConfig) checkAndPatchFuncDecl(funcGen *codegen.FuncGen) ([]*ast.Ident, error) {
	decl := funcGen.Func
	if decl.Type == nil {
		return nil, fmt.Errorf("expected function type")
	}

	if decl.Type.Results == nil || len(decl.Type.Results.List) == 0 {
		return nil, c.astError(decl.Type, "expected non-empty output parameters")
	}

	numOut := len(decl.Type.Results.List)
	if decl.Type.Results.List[numOut-1].Type.(*ast.Ident).Name != "error" {
		return nil, c.astError(decl.Type.Results.List[numOut-1].Type, "last output parameter must be error")
	}

	if decl.Body == nil || len(decl.Body.List) != 2 {
		return nil, c.astError(decl.Body, "expected exactly 2 statements in function body")
	}

	stmt, ok := decl.Body.List[0].(*ast.AssignStmt)
	if !ok {
		return nil, c.astError(decl.Body.List[0], "expected err = depinject.InjectDebug(...)")
	}

	for _, arg := range stmt.Rhs[0].(*ast.CallExpr).Args[0].(*ast.CallExpr).Args {
		ident := arg.(*ast.Ident)
		for _, d := range funcGen.File.Decls {
			if g, ok := d.(*ast.GenDecl); ok {
				for i, spec := range g.Specs {
					if v, ok := spec.(*ast.ValueSpec); ok {
						for _, name := range v.Names {
							if name.Name == ident.Name {
								g.Specs[i] = nil
							}
						}
					}
				}
			}
		}
	}

	var funcParamNames []*ast.Ident
	for _, field := range decl.Type.Params.List {
		for _, name := range field.Names {
			funcParamNames = append(funcParamNames, name)
		}
	}

	ret, ok := decl.Body.List[1].(*ast.ReturnStmt)
	if !ok || len(ret.Results) > 0 {
		return nil, fmt.Errorf("expected return (without any arguments) to be the last statement in the function")
	}

	decl.Type.Results.List = nil
	decl.Body.List = nil

	return funcParamNames, nil
}

func (c *debugConfig) astError(node ast.Node, format string, args ...any) error {
	tokenFile := c.fset.File(node.Pos())
	position := tokenFile.Position(node.Pos())
	str := fmt.Sprintf("at %s\n  ", position)
	return fmt.Errorf(str+format, args...)
}

func (c *container) codegenStmt(stmt ast.Stmt) {
	c.funcGen.Func.Body.List = append(c.funcGen.Func.Body.List, stmt)
}
