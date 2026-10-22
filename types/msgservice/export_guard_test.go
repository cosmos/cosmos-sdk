package msgservice_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// exportGuardAllowed lists the source paths, relative to the repo root, that
// may export a msg server or a Msg handler shaped method: test fixtures that
// hold no module state, and a test helper that runs msgs through a router.
var exportGuardAllowed = []string{
	// the testpb Msg service fixture baseapp tests register
	"testutil/testdata/",
	// the bank keeper mock whose Msg methods return nil
	"x/authz/testutil/bank_helpers.go",
	// Helper.CreateValidatorWithMsg runs the msg with msgservice.Dispatch
	"x/staking/testutil/helpers.go",
}

// TestNoExportedMsgServer scans every hand-written Go file in the repo and
// fails if a package exports a way to call a msg server outside the
// MsgServiceRouter: an exported func or var that returns or has a MsgServer
// type, a compile-time MsgServer assertion on an exported type, or an
// exported method shaped like a Msg handler on an exported type. Code
// outside a module runs a msg with msgservice.Dispatch.
func TestNoExportedMsgServer(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	require.NoError(t, err)

	var found []string
	scanned := 0
	fset := token.NewFileSet()
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if d.IsDir() {
			if d.Name() == ".git" || d.Name() == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(rel, ".go") || strings.HasSuffix(rel, "_test.go") ||
			strings.HasSuffix(rel, ".pb.go") || strings.HasSuffix(rel, ".pulsar.go") {
			return nil
		}
		for _, allowed := range exportGuardAllowed {
			if strings.HasPrefix(rel, allowed) {
				return nil
			}
		}
		file, err := parser.ParseFile(fset, path, nil, parser.ParseComments|parser.SkipObjectResolution)
		if err != nil {
			return err
		}
		scanned++
		// generated code, such as gomock mocks of an expected keeper that
		// includes a MsgServer, is not an API anyone writes
		if ast.IsGenerated(file) {
			return nil
		}
		for _, decl := range file.Decls {
			for _, reason := range exportedMsgServer(decl) {
				found = append(found, rel+": "+reason)
			}
		}
		return nil
	})
	require.NoError(t, err)
	require.Greater(t, scanned, 1000, "the scan found too few files; is the repo root right?")
	require.Empty(t, found, "these declarations hand out a msg server; register it with a keeper package RegisterMsgServer and keep the type and constructor unexported")
}

// exportedMsgServer returns why decl exports a msg server, if it does.
func exportedMsgServer(decl ast.Decl) []string {
	var reasons []string
	switch d := decl.(type) {
	case *ast.FuncDecl:
		recv := ""
		if d.Recv != nil && len(d.Recv.List) == 1 {
			recv = typeName(d.Recv.List[0].Type)
			if !ast.IsExported(recv) {
				return nil
			}
		}
		if !d.Name.IsExported() {
			return nil
		}
		if d.Type.Results != nil {
			for _, res := range d.Type.Results.List {
				if strings.HasSuffix(typeName(res.Type), "MsgServer") {
					reasons = append(reasons, "func "+d.Name.Name+" returns a MsgServer")
				}
			}
		}
		if recv != "" && isMsgHandler(d.Type) {
			reasons = append(reasons, "method "+recv+"."+d.Name.Name+" is a Msg handler on an exported type")
		}
	case *ast.GenDecl:
		if d.Tok != token.VAR {
			return nil
		}
		for _, spec := range d.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok || vs.Type == nil || !strings.HasSuffix(typeName(vs.Type), "MsgServer") {
				continue
			}
			for i, name := range vs.Names {
				switch {
				case name.IsExported():
					reasons = append(reasons, "var "+name.Name+" has a MsgServer type")
				case name.Name == "_" && i < len(vs.Values) && ast.IsExported(typeName(vs.Values[i])):
					reasons = append(reasons, "exported type "+typeName(vs.Values[i])+" implements "+typeName(vs.Type))
				}
			}
		}
	}
	return reasons
}

// isMsgHandler reports whether ft has the shape of a Msg service method:
// (context.Context, *MsgX) (*MsgXResponse, error).
func isMsgHandler(ft *ast.FuncType) bool {
	if ft.Params == nil || ft.Results == nil || len(ft.Params.List) != 2 || len(ft.Results.List) != 2 {
		return false
	}
	req, ok := ft.Params.List[1].Type.(*ast.StarExpr)
	if !ok {
		return false
	}
	resp, ok := ft.Results.List[0].Type.(*ast.StarExpr)
	if !ok {
		return false
	}
	reqName := typeName(req)
	return strings.HasPrefix(reqName, "Msg") && typeName(resp) == reqName+"Response"
}

// typeName returns the name a type expression or composite value refers to,
// without its package qualifier or pointer: Keeper for Keeper{}, &Keeper{},
// (*Keeper)(nil) and *keeper.Keeper.
func typeName(e ast.Expr) string {
	switch x := e.(type) {
	case *ast.Ident:
		return x.Name
	case *ast.SelectorExpr:
		return x.Sel.Name
	case *ast.StarExpr:
		return typeName(x.X)
	case *ast.ParenExpr:
		return typeName(x.X)
	case *ast.UnaryExpr:
		return typeName(x.X)
	case *ast.CompositeLit:
		return typeName(x.Type)
	case *ast.CallExpr:
		return typeName(x.Fun)
	case *ast.IndexExpr:
		return typeName(x.X)
	}
	return ""
}
