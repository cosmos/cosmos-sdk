package codegen

import (
	"fmt"
	"go/ast"
	"go/token"
	"strconv"
)

type FileGen struct {
	File           *ast.File
	idents         map[string]bool
	codegenPkgPath string
	pkgImportMap   map[string]*importInfo
}

func NewFileGen(file *ast.File, codegenPkgPath string) (*FileGen, error) {
	g := &FileGen{
		File:           file,
		idents:         map[string]bool{},
		codegenPkgPath: codegenPkgPath,
		pkgImportMap:   map[string]*importInfo{},
	}

	// add all go keywords to reserved idents
	for i := token.Token(0); i <= token.TILDE; i++ {
		name := i.String()
		if token.IsKeyword(name) {
			g.idents[name] = true
		}
	}

	for _, spec := range file.Imports {
		pkgPath, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			return nil, err
		}

		var info *importInfo
		if spec.Name != nil {
			name := spec.Name.Name
			if name == "." {
				return nil, fmt.Errorf(". package imports are not allowed")
			}

			info = &importInfo{importPrefix: name, ImportSpec: spec}
		} else {
			prefix := defaultPkgPrefix(pkgPath)
			info = &importInfo{importPrefix: prefix, ImportSpec: spec}
		}
		g.pkgImportMap[pkgPath] = info
		g.idents[info.importPrefix] = true
	}

	return g, nil
}

func (g *FileGen) PatchFuncDecl(name string) *FuncGen {
	for _, decl := range g.File.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if ok {
			if funcDecl.Name.Name == name {
				return newFuncGen(g, funcDecl)
			}
		}
	}
	return nil
}
