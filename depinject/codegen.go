package depinject

import (
	"fmt"
	"go/ast"
	"go/parser"
	"os"
	"path"

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
	loc := locationFromCaller(1)
	return debugOption(func(config *debugConfig) error {
		f, err := parser.ParseFile(config.fset, loc.file, nil, parser.ParseComments|parser.AllErrors)
		if err != nil {
			return err
		}

		fileGen, err := codegen.NewFileGen(f, loc.pkg)
		if err != nil {
			return err
		}

		funcGen := fileGen.PatchFuncDecl(loc.name)
		if funcGen == nil {
			return fmt.Errorf("couldn't resolve function %s in %s", loc.name, loc.file)
		}

		err = config.checkFuncDecl(funcGen.Func)
		if err != nil {
			return err
		}

		// TODO check existing build comments
		fileGen.File.Comments[0] = &ast.CommentGroup{
			[]*ast.Comment{
				{
					Text: "//go:build !depinject\n",
				},
			},
		}
		funcGen.Func.Type.Results.List = nil
		funcGen.Func.Body.List = nil

		config.funcGen = funcGen
		outFilename := loc.file
		ext := path.Ext(outFilename)
		outFilename = outFilename[0:len(outFilename)-len(ext)] + ".depinject.go"
		fmt.Printf("codegen output to %s\n", outFilename)
		outFile, err := os.OpenFile(outFilename, os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0644)
		if err != nil {
			return err
		}

		config.codegenOut = outFile
		config.codegenLoc = loc
		return nil
	})
}

func (c *debugConfig) checkFuncDecl(decl *ast.FuncDecl) error {
	if decl.Type == nil {
		return fmt.Errorf("expected function type")
	}

	if decl.Type.Results == nil || len(decl.Type.Results.List) == 0 {
		return c.astError(decl.Type, "expected non-empty output parameters")
	}

	numOut := len(decl.Type.Results.List)
	if decl.Type.Results.List[numOut-1].Type.(*ast.Ident).Name != "error" {
		return fmt.Errorf("last output parameter must be error")
	}

	if decl.Body == nil || len(decl.Body.List) != 2 {
		return fmt.Errorf("expected exactly 2 statements in function body")
	}

	decl.Pos()
	ret, ok := decl.Body.List[1].(*ast.ReturnStmt)
	if !ok || len(ret.Results) > 0 {
		return fmt.Errorf("expected return (without any arguments) to be the last statement in the function")
	}

	return nil
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
