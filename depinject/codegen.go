package depinject

import (
	"fmt"
	"go/ast"
	"go/parser"
	"os"

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
	loc := LocationFromCaller(1)
	return debugOption(func(config *debugConfig) error {
		f, err := parser.ParseFile(config.fset, loc.File(), nil, parser.ParseComments|parser.AllErrors)
		if err != nil {
			return err
		}

		fileGen, err := codegen.NewFileGen(f, loc.PkgPath())
		if err != nil {
			return err
		}

		funcGen := fileGen.PatchFuncDecl(loc.ShortName())
		if funcGen == nil {
			return fmt.Errorf("couldn't resolve function %s in %s", loc.ShortName(), loc.File())
		}

		config.funcGen = funcGen
		config.codegenOut = os.Stdout
		return nil
	})
}

func (c *container) codegenStmt(stmt ast.Stmt) {
	c.funcGen.Func.Body.List = append(c.funcGen.Func.Body.List, stmt)
}
