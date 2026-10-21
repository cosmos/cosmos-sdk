package msgservice_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	safetytypes "github.com/cosmos/cosmos-sdk/x/safety/types"
)

const (
	sdkModulePath = "github.com/cosmos/cosmos-sdk/"
	msgservicePkg = sdkModulePath + "types/msgservice"
)

// msgServerIface describes the generated MsgServer interface of one Go
// package: each method name mapped to its request type name, and each
// registered proto message type name mapped to its full proto name.
type msgServerIface struct {
	methods   map[string]string
	fullNames map[string]string
}

// guardScan is the source tree scanned for msg servers, relative to the repo
// root. contrib modules are out of scope.
var guardScan = []string{"x", "enterprise"}

// TestEveryMsgServerMethodCallsGuardFirst parses every x/ and enterprise/
// module and requires each method that implements a generated MsgServer
// method to start with
//
//	if err := msgservice.CheckMsgAllowed(ctx, msg); err != nil {
//		return nil, err
//	}
//
// on its own Context and request parameters. Methods whose msg x/safety
// always allows may leave it out. It also requires every MsgServer method of
// those modules to have an implementation the scan recognizes, so a msg
// server written in a shape the scan misses fails instead of passing
// unchecked. Unlike the app sweeps in the SimApp modules, it also covers
// modules that no app wires.
func TestEveryMsgServerMethodCallsGuardFirst(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	require.NoError(t, err)

	ifaces := map[string]*msgServerIface{}
	var sources []string
	for _, dir := range guardScan {
		require.NoError(t, filepath.WalkDir(filepath.Join(root, dir), func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				// generated pulsar APIs, mocks and test apps define no msg
				// servers that a chain runs
				switch d.Name() {
				case "api", "testutil", "simapp", "testdata":
					return filepath.SkipDir
				}
				return nil
			}
			switch {
			case strings.HasSuffix(p, "_test.go") || !strings.HasSuffix(p, ".go"):
			case strings.HasSuffix(p, ".pb.go"):
				scanPB(t, root, p, ifaces)
			default:
				sources = append(sources, p)
			}
			return nil
		}))
	}
	require.NotEmpty(t, ifaces, "no generated MsgServer interfaces found under %v", guardScan)

	implemented := map[string]bool{}
	var guarded, exempt, missing []string
	fset := token.NewFileSet()
	for _, p := range sources {
		f, err := parser.ParseFile(fset, p, nil, parser.SkipObjectResolution)
		require.NoError(t, err)
		imports := importPaths(f)
		fileDir := relDir(root, p)
		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv == nil || fn.Body == nil {
				continue
			}
			ctxName, msgName, reqDir, reqType, ok := msgServerParams(fn, imports, fileDir)
			if !ok {
				continue
			}
			iface := ifaces[reqDir]
			if iface == nil || iface.methods[fn.Name.Name] != reqType {
				continue
			}
			fullName, ok := iface.fullNames[reqType]
			require.True(t, ok, "%s: %s has no registered proto name", fset.Position(fn.Pos()), reqType)
			url := "/" + fullName
			implemented[reqDir+"."+fn.Name.Name] = true
			where := fset.Position(fn.Pos()).String()
			label := url + " " + strings.TrimPrefix(where, root+string(filepath.Separator))

			switch {
			case safetytypes.IsProtected(url):
				exempt = append(exempt, label)
			case startsWithGuard(fn, imports, ctxName, msgName):
				guarded = append(guarded, label)
			default:
				missing = append(missing, label)
			}
		}
	}

	var unrecognized []string
	for dir, iface := range ifaces {
		for method, reqType := range iface.methods {
			if _, registered := iface.fullNames[reqType]; !registered {
				continue
			}
			if !implemented[dir+"."+method] {
				unrecognized = append(unrecognized, dir+" "+method)
			}
		}
	}
	slices.Sort(guarded)
	slices.Sort(exempt)
	slices.Sort(unrecognized)
	for _, l := range guarded {
		t.Logf("guarded %s", l)
	}
	for _, l := range exempt {
		t.Logf("exempt  %s", l)
	}

	require.Empty(t, missing, "msg server methods that do not call msgservice.CheckMsgAllowed first")
	require.Empty(t, unrecognized, "MsgServer methods with no implementation the scan recognizes")
	require.NotEmpty(t, guarded)
}

// TestReadmeNamesUnguardedContribMsgServers requires the "Direct keeper calls"
// section of the x/safety README to name every contrib module that registers
// a msg server, since TestEveryMsgServerMethodCallsGuardFirst does not scan
// contrib and those msg servers do not call the guard.
func TestReadmeNamesUnguardedContribMsgServers(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	require.NoError(t, err)
	readme, err := os.ReadFile(filepath.Join(root, "x", "safety", "README.md"))
	require.NoError(t, err)
	_, gaps, found := strings.Cut(string(readme), "\n### Direct keeper calls\n")
	require.True(t, found, "x/safety README has no Direct keeper calls section")
	gaps, _, _ = strings.Cut(gaps, "\n#")

	contribX := filepath.Join(root, "contrib", "x")
	var files []string
	require.NoError(t, filepath.WalkDir(contribX, func(p string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() && strings.HasSuffix(p, ".go") &&
			!strings.HasSuffix(p, "_test.go") && !strings.HasSuffix(p, ".pb.go") {
			files = append(files, p)
		}
		return err
	}))

	modules := map[string]bool{}
	for _, p := range files {
		src, err := os.ReadFile(p)
		require.NoError(t, err)
		if !strings.Contains(string(src), "RegisterMsgServer(") {
			continue
		}
		rel, err := filepath.Rel(contribX, p)
		require.NoError(t, err)
		modules["contrib/x/"+strings.Split(filepath.ToSlash(rel), "/")[0]] = true
	}
	require.NotEmpty(t, modules)

	for m := range modules {
		require.Contains(t, gaps, "`"+m+"`", "x/safety README known gaps do not name unguarded msg server %s", m)
	}
}

// scanPB records the MsgServer interface and the registered proto names of a
// gogoproto generated file.
func scanPB(t *testing.T, root, p string, ifaces map[string]*msgServerIface) {
	t.Helper()
	f, err := parser.ParseFile(token.NewFileSet(), p, nil, parser.SkipObjectResolution)
	require.NoError(t, err)
	dir := relDir(root, p)
	iface := ifaces[dir]
	if iface == nil {
		iface = &msgServerIface{methods: map[string]string{}, fullNames: map[string]string{}}
		ifaces[dir] = iface
	}
	ast.Inspect(f, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.TypeSpec:
			it, ok := n.Type.(*ast.InterfaceType)
			if !ok || n.Name.Name != "MsgServer" {
				return true
			}
			for _, m := range it.Methods.List {
				ft, ok := m.Type.(*ast.FuncType)
				if !ok || len(m.Names) != 1 || ft.Params.NumFields() != 2 {
					continue
				}
				if star, ok := ft.Params.List[len(ft.Params.List)-1].Type.(*ast.StarExpr); ok {
					if id, ok := star.X.(*ast.Ident); ok {
						iface.methods[m.Names[0].Name] = id.Name
					}
				}
			}
		case *ast.CallExpr:
			// proto.RegisterType((*MsgSend)(nil), "cosmos.bank.v1beta1.MsgSend")
			sel, ok := n.Fun.(*ast.SelectorExpr)
			if !ok || sel.Sel.Name != "RegisterType" || len(n.Args) != 2 {
				return true
			}
			conv, ok := n.Args[0].(*ast.CallExpr)
			if !ok {
				return true
			}
			paren, ok := conv.Fun.(*ast.ParenExpr)
			if !ok {
				return true
			}
			star, ok := paren.X.(*ast.StarExpr)
			if !ok {
				return true
			}
			id, ok := star.X.(*ast.Ident)
			lit, isLit := n.Args[1].(*ast.BasicLit)
			if !ok || !isLit {
				return true
			}
			name, err := strconv.Unquote(lit.Value)
			require.NoError(t, err)
			iface.fullNames[id.Name] = name
		}
		return true
	})
}

// relDir returns the repo-relative slash directory of file p.
func relDir(root, p string) string {
	rel, err := filepath.Rel(root, filepath.Dir(p))
	if err != nil {
		panic(err)
	}
	return filepath.ToSlash(rel)
}

// importPaths maps each import name of f to its import path.
func importPaths(f *ast.File) map[string]string {
	out := map[string]string{}
	for _, imp := range f.Imports {
		ip, err := strconv.Unquote(imp.Path.Value)
		if err != nil {
			continue
		}
		name := path.Base(ip)
		if imp.Name != nil {
			name = imp.Name.Name
		}
		out[name] = ip
	}
	return out
}

// msgServerParams reports whether fn has the msg server shape
// (ctx context.Context, msg *T) (*R, error), returning the parameter names
// and the repo-relative directory and name of T.
func msgServerParams(fn *ast.FuncDecl, imports map[string]string, fileDir string) (ctxName, msgName, reqDir, reqType string, ok bool) {
	var names []string
	var types []ast.Expr
	for _, field := range fn.Type.Params.List {
		for _, n := range field.Names {
			names = append(names, n.Name)
			types = append(types, field.Type)
		}
	}
	if len(names) != 2 || fn.Type.Results.NumFields() != 2 {
		return "", "", "", "", false
	}
	if sel, isSel := types[0].(*ast.SelectorExpr); !isSel || sel.Sel.Name != "Context" {
		return "", "", "", "", false
	}
	star, isStar := types[1].(*ast.StarExpr)
	if !isStar {
		return "", "", "", "", false
	}
	switch x := star.X.(type) {
	case *ast.Ident:
		return names[0], names[1], fileDir, x.Name, true
	case *ast.SelectorExpr:
		pkg, isIdent := x.X.(*ast.Ident)
		if !isIdent {
			return "", "", "", "", false
		}
		ip, found := imports[pkg.Name]
		if !found || !strings.HasPrefix(ip, sdkModulePath) {
			return "", "", "", "", false
		}
		return names[0], names[1], strings.TrimPrefix(ip, sdkModulePath), x.Sel.Name, true
	}
	return "", "", "", "", false
}

// startsWithGuard reports whether the first statement of fn is
//
//	if err := msgservice.CheckMsgAllowed(ctxName, msgName); err != nil {
//		return nil, err
//	}
func startsWithGuard(fn *ast.FuncDecl, imports map[string]string, ctxName, msgName string) bool {
	if len(fn.Body.List) == 0 {
		return false
	}
	ifs, ok := fn.Body.List[0].(*ast.IfStmt)
	if !ok || ifs.Else != nil {
		return false
	}
	init, ok := ifs.Init.(*ast.AssignStmt)
	if !ok || init.Tok != token.DEFINE || len(init.Lhs) != 1 || len(init.Rhs) != 1 {
		return false
	}
	errID, ok := init.Lhs[0].(*ast.Ident)
	if !ok {
		return false
	}
	call, ok := init.Rhs[0].(*ast.CallExpr)
	if !ok || len(call.Args) != 2 {
		return false
	}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "CheckMsgAllowed" {
		return false
	}
	pkg, ok := sel.X.(*ast.Ident)
	if !ok || imports[pkg.Name] != msgservicePkg {
		return false
	}
	if !isIdent(call.Args[0], ctxName) || !isIdent(call.Args[1], msgName) {
		return false
	}
	cond, ok := ifs.Cond.(*ast.BinaryExpr)
	if !ok || cond.Op != token.NEQ || !isIdent(cond.X, errID.Name) || !isIdent(cond.Y, "nil") {
		return false
	}
	if len(ifs.Body.List) != 1 {
		return false
	}
	ret, ok := ifs.Body.List[0].(*ast.ReturnStmt)
	return ok && len(ret.Results) == 2 && isIdent(ret.Results[0], "nil") && isIdent(ret.Results[1], errID.Name)
}

func isIdent(e ast.Expr, name string) bool {
	id, ok := e.(*ast.Ident)
	return ok && id.Name == name
}
