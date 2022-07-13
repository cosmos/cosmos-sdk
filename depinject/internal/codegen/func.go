package codegen

import (
	"go/ast"
)

type FuncGen struct {
	*FileGen
	Func   *ast.FuncDecl
	idents map[string]bool
}

func newFuncGen(fileGen *FileGen, f *ast.FuncDecl) *FuncGen {
	g := &FuncGen{FileGen: fileGen, Func: f, idents: map[string]bool{}}

	// reserve param idents
	for _, field := range f.Type.Params.List {
		for _, name := range field.Names {
			g.idents[name.Name] = true
		}
	}

	// reserve result
	for _, field := range f.Type.Results.List {
		for _, name := range field.Names {
			g.idents[name.Name] = true
		}
	}

	return g
}
