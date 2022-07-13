package codegen

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"
)

type FileGen struct {
	File           *ast.File
	idents         map[string]bool
	codegenPkgPath string
	pkgImportMap   map[string]importInfo
}

func NewFileGen(file *ast.File, codegenPkgPath string) (*FileGen, error) {
	g := &FileGen{
		File:           file,
		idents:         map[string]bool{},
		codegenPkgPath: codegenPkgPath,
		pkgImportMap:   map[string]importInfo{},
	}

	for _, spec := range file.Imports {
		if spec.Name != nil {
			name := spec.Name.Name
			if name == "." {
				return nil, fmt.Errorf(". package imports are not allowed")
			}

			g.pkgImportMap[name] = importInfo{importPrefix: name, ImportSpec: spec}
		} else {
			prefix := defaultPkgPrefix(spec.Path.Value)
			g.pkgImportMap[prefix] = importInfo{importPrefix: prefix, ImportSpec: spec}
		}
	}

	return g, nil
}

type importInfo struct {
	*ast.ImportSpec
	importPrefix string
}

func (g *FileGen) AddOrGetImport(pkgPath string) (importPrefix string) {
	if pkgPath == "" || pkgPath == g.codegenPkgPath {
		return ""
	}

	if i, ok := g.pkgImportMap[pkgPath]; ok {
		return i.importPrefix
	}

	imp := &ast.ImportSpec{
		Path: &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", pkgPath)},
	}

	defaultPrefix := defaultPkgPrefix(pkgPath)
	importPrefix = g.doCreateIdent(defaultPrefix)

	if importPrefix != defaultPrefix {
		imp.Name = ast.NewIdent(importPrefix)
	}
	g.File.Imports = append(g.File.Imports, imp)
	g.pkgImportMap[pkgPath] = importInfo{
		ImportSpec:   imp,
		importPrefix: importPrefix,
	}
	g.idents[importPrefix] = true
	return importPrefix
}

func (g *FileGen) CreateIdent(namePrefix string) *ast.Ident {
	return ast.NewIdent(g.doCreateIdent(namePrefix))
}

func (g *FileGen) doCreateIdent(namePrefix string) string {
	// TODO reserved names: keywords, builtin types, imports, err
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

func defaultPkgPrefix(pkgPath string) string {
	pkgParts := strings.Split(pkgPath, "/")
	return pkgParts[len(pkgParts)-1]
}

func (g *FileGen) PatchFuncDecl(name string) *FuncGen {
	for _, decl := range g.File.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if ok {
			if funcDecl.Name.Name == name {
				return &FuncGen{
					FileGen: g,
					Func:    funcDecl,
					idents:  map[string]bool{},
				}
			}
		}
	}
	return nil
}
