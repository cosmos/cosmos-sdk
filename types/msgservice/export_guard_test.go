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
// MsgServiceRouter: an exported func, var or struct field whose type or
// value names a msg server (MsgServer, msgServer, legacyMsgServer,
// newMsgServerImpl), a compile-time MsgServer assertion on an exported
// type, or an exported method shaped like a Msg handler on an exported
// type. Code outside a module runs a msg with msgservice.Dispatch.
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
				if namesMsgServer(typeName(res.Type)) {
					reasons = append(reasons, "func "+d.Name.Name+" returns a msg server")
				}
			}
		}
		if recv != "" && isMsgHandler(d.Type) {
			reasons = append(reasons, "method "+recv+"."+d.Name.Name+" is a Msg handler on an exported type")
		}
	case *ast.GenDecl:
		for _, spec := range d.Specs {
			switch s := spec.(type) {
			case *ast.ValueSpec:
				if d.Tok == token.VAR {
					reasons = append(reasons, exportedMsgServerVar(s)...)
				}
			case *ast.TypeSpec:
				reasons = append(reasons, exportedMsgServerField(s)...)
			}
		}
	}
	return reasons
}

// exportedMsgServerVar returns why vs exports a msg server, if it does.
func exportedMsgServerVar(vs *ast.ValueSpec) []string {
	var reasons []string
	hasType := vs.Type != nil && namesMsgServer(typeName(vs.Type))
	for i, name := range vs.Names {
		value := ""
		if i < len(vs.Values) {
			value = typeName(vs.Values[i])
		}
		switch {
		case name.IsExported() && hasType:
			reasons = append(reasons, "var "+name.Name+" has a msg server type")
		case name.IsExported() && namesMsgServer(value):
			reasons = append(reasons, "var "+name.Name+" holds "+value)
		case name.Name == "_" && hasType && ast.IsExported(value):
			reasons = append(reasons, "exported type "+value+" implements "+typeName(vs.Type))
		}
	}
	return reasons
}

// exportedMsgServerField returns why ts, if it is an exported struct, has an
// exported or embedded msg server field.
func exportedMsgServerField(ts *ast.TypeSpec) []string {
	st, ok := ts.Type.(*ast.StructType)
	if !ok || !ts.Name.IsExported() {
		return nil
	}
	var reasons []string
	for _, field := range st.Fields.List {
		if !namesMsgServer(typeName(field.Type)) {
			continue
		}
		if len(field.Names) == 0 {
			reasons = append(reasons, "type "+ts.Name.Name+" embeds "+typeName(field.Type))
		}
		for _, name := range field.Names {
			if name.IsExported() {
				reasons = append(reasons, "field "+ts.Name.Name+"."+name.Name+" has a msg server type")
			}
		}
	}
	return reasons
}

// namesMsgServer reports whether a type or value name refers to a msg
// server: MsgServer, msgServer, legacyMsgServer or newMsgServerImpl.
func namesMsgServer(name string) bool {
	return strings.Contains(strings.ToLower(name), "msgserver")
}

// isMsgHandler reports whether ft has the shape of a Msg service method:
// (context.Context, *MsgX) (*MsgYResponse, error). The response name need
// not match the request: counter's IncreaseCount takes MsgIncreaseCounter
// and returns MsgIncreaseCountResponse.
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
	respName := typeName(resp)
	return strings.HasPrefix(typeName(req), "Msg") && strings.HasPrefix(respName, "Msg") &&
		strings.HasSuffix(respName, "Response")
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

// TestExportedMsgServerShapes checks the guard against declarations that
// hand out a msg server and against ones that do not.
func TestExportedMsgServerShapes(t *testing.T) {
	cases := []struct {
		name string
		src  string
		flag bool
	}{
		{"constructor returns the interface", `func NewMsgServerImpl(k Keeper) types.MsgServer { return msgServer{k} }`, true},
		{"constructor returns the unexported type", `func NewMsgServerImpl(k Keeper) *msgServer { return &msgServer{k} }`, true},
		{"constructor returns an unexported legacy type", `func NewLegacy(k Keeper) legacyMsgServer { return legacyMsgServer{k} }`, true},
		{"var aliases the unexported constructor", `var NewMsgServerImpl = newMsgServerImpl`, true},
		{"var holds an unexported msg server", `var Server = msgServer{}`, true},
		{"var has a MsgServer type", `var Server types.MsgServer`, true},
		{"exported field has a MsgServer type", `type App struct { Bank types.MsgServer }`, true},
		{"exported type embeds a MsgServer", `type App struct { types.MsgServer }`, true},
		{"exported type asserts MsgServer", `var _ types.MsgServer = Keeper{}`, true},
		{"handler with a matching response name", `func (k Keeper) Send(ctx context.Context, msg *types.MsgSend) (*types.MsgSendResponse, error) { return nil, nil }`, true},
		{"handler with a different response name", `func (k *Keeper) IncreaseCount(ctx context.Context, msg *types.MsgIncreaseCounter) (*types.MsgIncreaseCountResponse, error) { return nil, nil }`, true},
		{"unexported constructor", `func newMsgServerImpl(k Keeper) *msgServer { return &msgServer{k} }`, false},
		{"registration", `func RegisterMsgServer(r grpc.ServiceRegistrar, k Keeper) { types.RegisterMsgServer(r, newMsgServerImpl(k)) }`, false},
		{"unexported field", `type App struct { bank types.MsgServer }`, false},
		{"unexported type", `type app struct { Bank types.MsgServer }`, false},
		{"method on the unexported type", `func (m msgServer) Send(ctx context.Context, msg *types.MsgSend) (*types.MsgSendResponse, error) { return nil, nil }`, false},
		{"keeper logic", `func (k Keeper) SendCoins(ctx context.Context, from, to sdk.AccAddress, amt sdk.Coins) error { return nil }`, false},
		{"query handler", `func (k Keeper) Balance(ctx context.Context, req *types.QueryBalanceRequest) (*types.QueryBalanceResponse, error) { return nil, nil }`, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			file, err := parser.ParseFile(token.NewFileSet(), "x.go", "package keeper\n"+tc.src, parser.SkipObjectResolution)
			require.NoError(t, err)
			var reasons []string
			for _, decl := range file.Decls {
				reasons = append(reasons, exportedMsgServer(decl)...)
			}
			if tc.flag {
				require.NotEmpty(t, reasons)
			} else {
				require.Empty(t, reasons)
			}
		})
	}
}
