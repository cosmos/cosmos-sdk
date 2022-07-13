package codegen

import (
	"fmt"
	"go/ast"
)

func (g *FileGen) CreateIdent(namePrefix string) *ast.Ident {
	return ast.NewIdent(g.doCreateIdent(namePrefix))
}

func (g *FileGen) doCreateIdent(namePrefix string) string {
	v := namePrefix
	i := 2
	for {
		_, ok := g.idents[v]
		if !ok {
			g.idents[v] = true
			return v
		}

		v = fmt.Sprintf("%s%d", namePrefix, i)
		i++
	}
}

func (f *FuncGen) CreateIdent(namePrefix string) *ast.Ident {
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
