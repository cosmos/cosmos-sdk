package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("no input specified")
		os.Exit(-1)
	}

	in := os.Args[1]
	err := filepath.WalkDir(in, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() && filepath.Ext(path) == ".go" {
			src, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}

			if !strings.HasPrefix(string(src), "//go:build depinject") {
				return nil
			}

			return doCodegen(path, src)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
}

var depinjectRegex = regexp.MustCompile(`^//depinject:(.*)`)

func doCodegen(path string, src []byte) error {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, src, parser.ParseComments)
	if err != nil {
		return err
	}

	//ast.Print(fset, f)

	for _, decl := range f.Decls {
		switch decl := decl.(type) {
		case *ast.FuncDecl:
			for _, comment := range decl.Doc.List {
				matches := depinjectRegex.FindStringSubmatch(comment.Text)
				if len(matches) > 1 {
					err := codegenFunc(decl, matches[1])
					if err != nil {
						return err
					}
					break
				}
			}
		}
	}

	return printer.Fprint(os.Stdout, fset, f)
}

func codegenFunc(fn *ast.FuncDecl, config string) error {
	configs := strings.Split(config, ",")
	var configExprs []ast.Expr
	for _, cfg := range configs {
		configExprs = append(configExprs, ast.NewIdent(cfg))
	}

	var supplyExprs []ast.Expr
	for _, field := range fn.Type.Params.List {
		for _, name := range field.Names {
			supplyExprs = append(supplyExprs, name)
		}
	}

	if len(supplyExprs) > 0 {
		configExprs = append(configExprs, &ast.CallExpr{
			Fun:  ast.NewIdent("depinject.Supply"),
			Args: supplyExprs,
		})
	}

	numRet := len(fn.Type.Results.List)
	if numRet < 2 {
		return fmt.Errorf("function must return at least 2 params")
	}

	if fn.Type.Results.List[numRet-1].Type.(*ast.Ident).Name != "error" {
		return fmt.Errorf("last return parameter must be error")
	}

	args := []ast.Expr{
		&ast.CallExpr{
			Fun:  ast.NewIdent("depinject.Configs"),
			Args: configExprs,
		},
	}

	for i := 0; i < numRet-1; i++ {
		ret := fn.Type.Results.List[i]
		if len(ret.Names) == 0 {
			ret.Names = []*ast.Ident{ast.NewIdent(fmt.Sprintf("r%d", i))}
		}

		for _, name := range ret.Names {
			args = append(args, &ast.UnaryExpr{
				Op: token.AND,
				X:  name,
			})
		}
	}

	fn.Body.List = []ast.Stmt{
		&ast.BlockStmt{
			List: []ast.Stmt{
				&ast.AssignStmt{
					Lhs: []ast.Expr{
						ast.NewIdent("err"),
					},
					Tok: token.DEFINE,
					Rhs: []ast.Expr{
						&ast.CallExpr{
							Fun:  ast.NewIdent("depinject.Inject"),
							Args: args,
						},
					},
				},
			},
		},
	}

	return nil
}
