package simapp

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	safetytypes "github.com/cosmos/cosmos-sdk/x/safety/types"
)

// x/safety is new in this version, so the upgrade from v0.54 must add its
// store, and every added store must be one the app mounts.
func TestUpgradeStoreUpgrades(t *testing.T) {
	app := Setup(t, false)
	upgrades := upgradeStoreUpgrades()

	require.Contains(t, upgrades.Added, safetytypes.StoreKey)
	for _, name := range upgrades.Added {
		require.NotNil(t, app.GetKey(name), "added store %s is not mounted", name)
	}
	for _, name := range upgrades.Deleted {
		require.Nil(t, app.GetKey(name), "deleted store %s is still mounted", name)
	}
}

// Upgrade handlers call keepers. A msg run with msgservice.Dispatch goes
// through the circuit breaker, so a trip would halt the upgrade. The one
// exception, a chain's circuit migration dispatching the protected MsgTrip,
// does not apply to SimApp.
func TestUpgradeHandlersDoNotDispatch(t *testing.T) {
	const msgservicePath = "github.com/cosmos/cosmos-sdk/types/msgservice"

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "upgrades.go", nil, 0)
	require.NoError(t, err)

	names := map[string]bool{}
	for _, imp := range f.Imports {
		path, err := strconv.Unquote(imp.Path.Value)
		require.NoError(t, err)
		if path != msgservicePath {
			continue
		}
		name := "msgservice"
		if imp.Name != nil {
			name = imp.Name.Name
		}
		require.NotEqual(t, ".", name, "upgrades.go dot-imports msgservice")
		names[name] = true
	}

	ast.Inspect(f, func(n ast.Node) bool {
		sel, ok := n.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if pkg, ok := sel.X.(*ast.Ident); ok && names[pkg.Name] && sel.Sel.Name == "Dispatch" {
			t.Errorf("%s: upgrade handler calls msgservice.Dispatch; call the keeper", fset.Position(sel.Pos()))
		}
		return true
	})
}
