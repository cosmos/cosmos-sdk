package codegen

import (
	"fmt"
	"go/ast"
)

type FuncGen struct {
	*FileGen
	Func   *ast.FuncDecl
	idents map[string]bool
}

func (f *FuncGen) init() {
	if f.idents == nil {
		f.idents = map[string]bool{}
	}
}

func (f *FuncGen) CreateIdent(namePrefix string) *ast.Ident {
	f.init()

	v := namePrefix
	i := 2
	for {
		_, definedFileScope := f.FileGen.idents[v]

		_, definedFuncScope := f.idents[v]
		if !definedFileScope && !definedFuncScope {
			f.FileGen.idents[v] = true
			f.idents[v] = true
			return ast.NewIdent(v)
		}

		v = fmt.Sprintf("%s%d", namePrefix, i)
		i++
	}
}
