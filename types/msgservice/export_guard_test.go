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

// exportGuardDeprecated lists the deprecated constructors and keeper Msg
// methods kept so callers compile until the next release removes them. Each
// must carry a Deprecated: doc comment, and the guard fails when one is gone
// but still listed. A "Keeper." prefix names a keeper Msg method, and a
// "type " prefix an alias of the unexported msgServer type.
var exportGuardDeprecated = map[string][]string{
	"contrib/x/circuit/keeper/msg_server.go":        {"NewMsgServerImpl"},
	"contrib/x/crisis/keeper/msg_server.go":         {"Keeper.VerifyInvariant", "Keeper.UpdateParams"},
	"contrib/x/nft/keeper/msg_server.go":            {"Keeper.Send"},
	"enterprise/poa/x/poa/keeper/msg_server.go":     {"type MsgServer", "NewMsgServer"},
	"enterprise/group/x/group/keeper/msg_server.go": {"Keeper.CreateGroup", "Keeper.UpdateGroupMembers", "Keeper.UpdateGroupAdmin", "Keeper.UpdateGroupMetadata", "Keeper.CreateGroupWithPolicy", "Keeper.CreateGroupPolicy", "Keeper.UpdateGroupPolicyAdmin", "Keeper.UpdateGroupPolicyDecisionPolicy", "Keeper.UpdateGroupPolicyMetadata", "Keeper.SubmitProposal", "Keeper.WithdrawProposal", "Keeper.Vote", "Keeper.Exec", "Keeper.LeaveGroup"},
	"testutil/x/counter/keeper/keeper.go":           {"Keeper.IncreaseCount"},
	"x/auth/keeper/msg_server.go":                   {"NewMsgServerImpl"},
	"x/auth/vesting/msg_server.go":                  {"NewMsgServerImpl"},
	"x/authz/keeper/msg_server.go":                  {"Keeper.Grant", "Keeper.Revoke", "Keeper.Exec"},
	"x/bank/keeper/msg_server.go":                   {"NewMsgServerImpl"},
	"x/consensus/keeper/keeper.go":                  {"Keeper.UpdateParams"},
	"x/distribution/keeper/msg_server.go":           {"NewMsgServerImpl"},
	"x/evidence/keeper/msg_server.go":               {"NewMsgServerImpl"},
	"x/feegrant/keeper/msg_server.go":               {"NewMsgServerImpl"},
	"x/gov/keeper/msg_server.go":                    {"NewMsgServerImpl", "NewLegacyMsgServerImpl"},
	"x/mint/keeper/msg_server.go":                   {"NewMsgServerImpl"},
	"x/safety/keeper/msg_server.go":                 {"NewMsgServerImpl"},
	"x/slashing/keeper/msg_server.go":               {"NewMsgServerImpl"},
	"x/staking/keeper/msg_server.go":                {"NewMsgServerImpl"},
	"x/upgrade/keeper/msg_server.go":                {"NewMsgServerImpl"},
}

// deprecatedReasons returns the guard findings exportGuardDeprecated allows.
func deprecatedReasons() map[string]bool {
	allowed := map[string]bool{}
	for path, names := range exportGuardDeprecated {
		for _, name := range names {
			switch {
			case strings.HasPrefix(name, "Keeper."):
				allowed[path+": method "+name+" is a Msg handler on an exported type"] = true
			case strings.HasPrefix(name, "type "):
				allowed[path+": "+name+" is msgServer"] = true
			default:
				allowed[path+": func "+name+" returns a msg server"] = true
			}
		}
	}
	return allowed
}

// isDeprecated reports whether decl has a Deprecated: doc paragraph.
func isDeprecated(decl ast.Decl) bool {
	var doc *ast.CommentGroup
	switch d := decl.(type) {
	case *ast.FuncDecl:
		doc = d.Doc
	case *ast.GenDecl:
		doc = d.Doc
	}
	return doc != nil && strings.Contains(doc.Text(), "\nDeprecated: ")
}

// TestNoExportedMsgServer scans every hand-written Go file in the repo and
// fails if a package exports a way to call a msg server outside the
// MsgServiceRouter: an exported type, func, var or struct field whose type
// or value names a msg server (MsgServer, msgServer, legacyMsgServer,
// newMsgServerImpl), a compile-time MsgServer assertion on an exported
// type, an exported func or var that hands out a Msg handler func, or an
// exported method shaped like a Msg handler on an exported type. Code
// outside a module runs a msg with msgservice.Dispatch. Only the
// declarations in exportGuardDeprecated are exempt.
func TestNoExportedMsgServer(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	require.NoError(t, err)

	var found []string
	deprecated := deprecatedReasons()
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
				key := rel + ": " + reason
				if deprecated[key] && isDeprecated(decl) {
					delete(deprecated, key)
					continue
				}
				found = append(found, key)
			}
		}
		return nil
	})
	require.NoError(t, err)
	require.Greater(t, scanned, 1000, "the scan found too few files; is the repo root right?")
	require.Empty(t, found, "these declarations hand out a msg server; register it with a keeper package RegisterMsgServer and keep the type and constructor unexported")
	require.Empty(t, deprecated, "these exportGuardDeprecated entries match no declaration with a Deprecated: doc; remove them from the list")
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
				if ft, ok := res.Type.(*ast.FuncType); ok && isMsgHandler(ft) {
					reasons = append(reasons, "func "+d.Name.Name+" returns a Msg handler")
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
				reasons = append(reasons, exportedMsgServerType(s)...)
			}
		}
	}
	return reasons
}

// exportedMsgServerVar returns why vs exports a msg server, if it does.
func exportedMsgServerVar(vs *ast.ValueSpec) []string {
	var reasons []string
	hasType := vs.Type != nil && namesMsgServer(typeName(vs.Type))
	if ft, ok := vs.Type.(*ast.FuncType); ok && isMsgHandler(ft) {
		hasType = true
	}
	for i, name := range vs.Names {
		value, mentions := "", false
		if i < len(vs.Values) {
			value = typeName(vs.Values[i])
			mentions = mentionsMsgServer(vs.Values[i])
		}
		switch {
		case name.IsExported() && hasType:
			reasons = append(reasons, "var "+name.Name+" has a msg server type")
		case name.IsExported() && mentions:
			reasons = append(reasons, "var "+name.Name+" holds a msg server or one of its methods")
		case name.Name == "_" && hasType && ast.IsExported(value):
			reasons = append(reasons, "exported type "+value+" implements "+typeName(vs.Type))
		}
	}
	return reasons
}

// exportedMsgServerType returns why ts exports a msg server, if it does: it
// is an exported alias or definition of a msg server type, or an exported
// struct with an exported or embedded msg server field.
func exportedMsgServerType(ts *ast.TypeSpec) []string {
	if !ts.Name.IsExported() {
		return nil
	}
	if namesMsgServer(typeName(ts.Type)) {
		return []string{"type " + ts.Name.Name + " is " + typeName(ts.Type)}
	}
	st, ok := ts.Type.(*ast.StructType)
	if !ok {
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

// mentionsMsgServer reports whether any identifier in e names a msg server,
// as in newMsgServerImpl(k).Send.
func mentionsMsgServer(e ast.Expr) bool {
	found := false
	ast.Inspect(e, func(n ast.Node) bool {
		if id, ok := n.(*ast.Ident); ok && namesMsgServer(id.Name) {
			found = true
		}
		return !found
	})
	return found
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
		{"exported alias of the unexported type", `type Server = msgServer`, true},
		{"exported type defined from the unexported type", `type Server msgServer`, true},
		{"method returns a Msg handler", `func (k BaseKeeper) SendHandler() func(context.Context, *types.MsgSend) (*types.MsgSendResponse, error) { return newMsgServerImpl(k).Send }`, true},
		{"func returns a Msg handler", `func SendHandler(k Keeper) func(context.Context, *types.MsgSend) (*types.MsgSendResponse, error) { return newMsgServerImpl(k).Send }`, true},
		{"var holds a msg server method", `var Send = newMsgServerImpl(Keeper{}).Send`, true},
		{"var has a Msg handler type", `var Send func(context.Context, *types.MsgSend) (*types.MsgSendResponse, error)`, true},
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
