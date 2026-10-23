package safety_test

import (
	"context"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/log/v2"
	"cosmossdk.io/simapp"

	circuitkeeper "github.com/cosmos/cosmos-sdk/contrib/x/circuit/keeper"
	circuittypes "github.com/cosmos/cosmos-sdk/contrib/x/circuit/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/store/v2/rootmulti"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/module"
	safetytypes "github.com/cosmos/cosmos-sdk/x/safety/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

// withCircuit wires contrib/x/circuit into SimApp the way a chain still on it
// does: the circuit store is mounted, the circuit keeper is the router's
// circuit breaker, and the module version map records circuit.
func withCircuit(app *simapp.SimApp) {
	key := storetypes.NewKVStoreKey(circuittypes.StoreKey)
	app.MountStores(key)
	k := circuitkeeper.NewKeeper(app.AppCodec(), runtime.NewKVStoreService(key), govAddr, app.AccountKeeper.AddressCodec())
	circuittypes.RegisterInterfaces(app.InterfaceRegistry())
	circuittypes.RegisterMsgServer(app.MsgServiceRouter(), circuitkeeper.NewMsgServerImpl(k))
	app.SetCircuitBreaker(&k)
	app.SetInitChainer(func(ctx sdk.Context, req *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
		res, err := app.InitChainer(ctx, req)
		if err != nil {
			return nil, err
		}
		vm := app.ModuleManager.GetVersionMap()
		vm[circuittypes.ModuleName] = 1
		return res, app.UpgradeKeeper.SetModuleVersionMap(ctx, vm)
	})
}

// committedStores returns the names of the stores in the last commit.
func committedStores(t *testing.T, f *fixture) []string {
	t.Helper()
	rs, ok := f.app.CommitMultiStore().(*rootmulti.Store)
	require.True(t, ok)
	info, err := rs.GetCommitInfo(f.app.LastBlockHeight())
	require.NoError(t, err)
	var names []string
	for _, si := range info.StoreInfos {
		names = append(names, si.Name)
	}
	return names
}

// TestUpgradeFromCircuit runs the upgrade a chain moving off contrib/x/circuit
// makes: the new binary removes the circuit wiring and store key, and its
// handler names the x/safety operator right after RunMigrations. The circuit
// store leaves the commit, the new binary ignores circuit's disabled list, and
// the operator can trip in the upgrade block itself.
//
// SimApp always mounts the x/safety store, so the new binary loads with no
// StoreUpgrades. TestChainUpgrade in the system tests covers
// StoreUpgrades.Added.
func TestUpgradeFromCircuit(t *testing.T) {
	const upgradeName = "circuit-to-safety"
	f := newFixtureWithSetup(t, nil, nil, withCircuit)
	operator := f.operator.String()

	// gov trips MsgSend in x/circuit, and the router refuses it
	f.mustPass(&circuittypes.MsgTripCircuitBreaker{Authority: govAddr, MsgTypeUrls: []string{msgSendURL}})
	res := f.deliver(f.user, send(f.user, f.other))
	require.Equal(t, sdkerrors.ErrCircuitBreakerTripped.ABCICode(), res.Code, res.Log)
	require.Equal(t, sdkerrors.ErrCircuitBreakerTripped.Codespace(), res.Codespace)
	require.Contains(t, committedStores(t, f), circuittypes.StoreKey)

	height := f.app.LastBlockHeight() + 10
	f.mustPass(&upgradetypes.MsgSoftwareUpgrade{
		Authority: govAddr,
		Plan:      upgradetypes.Plan{Name: upgradeName, Height: height},
	})
	for f.app.LastBlockHeight() < height-1 {
		f.finalize()
	}

	// the new binary has no circuit wiring, so SimApp's x/safety keeper is
	// the circuit breaker in the router and the ante handler; the handler
	// names the operator and disables MsgSend again in the upgrade block
	var fromVM module.VersionMap
	app := simapp.NewSimApp(log.NewNopLogger(), f.db, false, simtestutil.NewAppOptionsWithFlagHome(f.home))
	app.UpgradeKeeper.SetUpgradeHandler(upgradeName, func(ctx context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		fromVM = vm
		vm, err := app.ModuleManager.RunMigrations(ctx, app.Configurator(), vm)
		if err != nil {
			return nil, err
		}
		return vm, app.SafetyKeeper.InitGenesis(ctx, safetytypes.NewGenesisState(safetytypes.Params{
			Authority: govAddr,
			Operator:  operator,
			Scope:     []string{safetytypes.Wildcard},
		}, []string{msgSendURL}))
	})
	require.NoError(t, app.LoadLatestVersion())
	f.app = app

	// in the upgrade block, MsgSend stays refused
	res = f.deliver(f.user, send(f.user, f.other))
	require.Equal(t, safetytypes.ErrMsgDisabled.ABCICode(), res.Code, res.Log)
	require.Equal(t, safetytypes.ErrMsgDisabled.Codespace(), res.Codespace)
	require.Equal(t, height, f.app.LastBlockHeight())
	require.NotContains(t, committedStores(t, f), circuittypes.StoreKey)
	require.Contains(t, fromVM, circuittypes.ModuleName)

	// RunMigrations skips circuit's version map entry, and the upgrade
	// keeper leaves it in the store
	vm, err := f.app.UpgradeKeeper.GetModuleVersionMap(f.ctx())
	require.NoError(t, err)
	require.Contains(t, vm, circuittypes.ModuleName)
	require.Equal(t, operator, f.params().Operator)
	require.Equal(t, []string{msgSendURL}, f.disabled())
	f.requireRejected(f.user, safetytypes.ErrMsgDisabled, send(f.user, f.other))
}
