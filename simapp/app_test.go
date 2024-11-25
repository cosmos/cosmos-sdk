package simapp

import (
	"encoding/json"
	"fmt"
	"testing"

	abci "github.com/cometbft/cometbft/api/cometbft/abci/v1"
	cmtproto "github.com/cometbft/cometbft/api/cometbft/types/v1"
	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"

	"cosmossdk.io/core/appmodule"
	coretesting "cosmossdk.io/core/testing"
	"cosmossdk.io/log"
	"cosmossdk.io/x/accounts"
	authzmodule "cosmossdk.io/x/authz/module"
	"cosmossdk.io/x/bank"
	banktypes "cosmossdk.io/x/bank/types"
	bankv2 "cosmossdk.io/x/bank/v2"
	"cosmossdk.io/x/distribution"
	"cosmossdk.io/x/epochs"
	"cosmossdk.io/x/evidence"
	feegrantmodule "cosmossdk.io/x/feegrant/module"
	"cosmossdk.io/x/gov"
	group "cosmossdk.io/x/group/module"
	"cosmossdk.io/x/mint"
	"cosmossdk.io/x/protocolpool"
	"cosmossdk.io/x/slashing"
	"cosmossdk.io/x/staking"
	"cosmossdk.io/x/upgrade"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/testutil/mock"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	"github.com/cosmos/cosmos-sdk/x/genutil"
)

func TestSimAppExportAndBlockedAddrs(t *testing.T) {
	db := coretesting.NewMemDB()
	logger := log.NewTestLogger(t)
	app := NewSimappWithCustomOptions(t, false, SetupOptions{
		Logger:  logger.With("instance", "first"),
		DB:      db,
		AppOpts: simtestutil.NewAppOptionsWithFlagHome(t.TempDir()),
	})

	// BlockedAddresses returns a map of addresses in app v1 and a map of modules names in app di.
	blockedAddrs, err := BlockedAddresses(app.interfaceRegistry.SigningContext().AddressCodec())
	require.NoError(t, err)
	for acc := range blockedAddrs {
		var addr sdk.AccAddress
		if modAddr, err := app.InterfaceRegistry().SigningContext().AddressCodec().StringToBytes(acc); err == nil {
			addr = modAddr
		} else {
			addr = app.AuthKeeper.GetModuleAddress(acc)
		}

		require.True(
			t,
			app.BankKeeper.BlockedAddr(addr),
			fmt.Sprintf("ensure that blocked addresses are properly set in bank keeper: %s should be blocked", acc),
		)
	}

	// finalize block so we have CheckTx state set
	_, err = app.FinalizeBlock(&abci.FinalizeBlockRequest{
		Height: 1,
	})
	require.NoError(t, err)

	_, err = app.Commit()
	require.NoError(t, err)

	// Making a new app object with the db, so that initchain hasn't been called
	app2 := NewSimApp(logger.With("instance", "second"), db, nil, true, simtestutil.NewAppOptionsWithFlagHome(t.TempDir()))
	_, err = app2.ExportAppStateAndValidators(false, []string{}, []string{})
	require.NoError(t, err, "ExportAppStateAndValidators should not have an error")
}

func TestRunMigrations(t *testing.T) {
	db := coretesting.NewMemDB()
	logger := log.NewTestLogger(t)
	app := NewSimApp(logger.With("instance", "simapp"), db, nil, true, simtestutil.NewAppOptionsWithFlagHome(t.TempDir()))

	// Create a new baseapp and configurator for the purpose of this test.
	bApp := baseapp.NewBaseApp(app.Name(), logger.With("instance", "baseapp"), db, app.TxConfig().TxDecoder())
	bApp.SetCommitMultiStoreTracer(nil)
	bApp.SetInterfaceRegistry(app.InterfaceRegistry())
	app.BaseApp = bApp
	configurator := module.NewConfigurator(app.appCodec, bApp.MsgServiceRouter(), app.GRPCQueryRouter())

	// We register all modules on the Configurator, except x/bank. x/bank will
	// serve as the test subject on which we run the migration tests.
	//
	// The loop below is the same as calling `RegisterServices` on
	// ModuleManager, except that we skip x/bank.
	for name, mod := range app.ModuleManager.Modules {
		if name == banktypes.ModuleName {
			continue
		}

		if mod, ok := mod.(module.HasServices); ok {
			mod.RegisterServices(configurator)
		}

		if mod, ok := mod.(interface {
			RegisterServices(grpc.ServiceRegistrar) error
		}); ok {
			err := mod.RegisterServices(configurator)
			require.NoError(t, err)
		}

		require.NoError(t, configurator.Error())
	}

	// Initialize the chain
	_, err := app.InitChain(&abci.InitChainRequest{})
	require.NoError(t, err)
	_, err = app.Commit()
	require.NoError(t, err)

	testCases := []struct {
		name         string
		moduleName   string
		fromVersion  uint64
		toVersion    uint64
		expRegErr    bool // errors while registering migration
		expRegErrMsg string
		expRunErr    bool // errors while running migration
		expRunErrMsg string
		expCalled    int
	}{
		{
			"cannot register migration for version 0",
			"bank", 0, 1,
			true, "module migration versions should start at 1: invalid version", false, "", 0,
		},
		{
			"throws error on RunMigrations if no migration registered for bank",
			"", 1, 2,
			false, "", true, "no migrations found for module bank: not found", 0,
		},
		{
			"can register 1->2 migration handler for x/bank, cannot run migration",
			"bank", 1, 2,
			false, "", true, "no migration found for module bank from version 2 to version 3: not found", 0,
		},
		{
			"can register 2->3 migration handler for x/bank, can run migration",
			"bank", 2, bank.AppModule{}.ConsensusVersion(),
			false, "", false, "", int(bank.AppModule{}.ConsensusVersion() - 2), // minus 2 because 1-2 is run in the previous test case.
		},
		{
			"cannot register migration handler for same module & fromVersion",
			"bank", 1, 2,
			true, "another migration for module bank and version 1 already exists: internal logic error", false, "", 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(tt *testing.T) {
			var err error

			// Since it's very hard to test actual in-place store migrations in
			// tests (due to the difficulty of maintaining multiple versions of a
			// module), we're just testing here that the migration logic is
			// called.
			called := 0

			if tc.moduleName != "" {
				for i := tc.fromVersion; i < tc.toVersion; i++ {
					// Register migration for module from version `fromVersion` to `fromVersion+1`.
					tt.Logf("Registering migration for %q v%d", tc.moduleName, i)
					err = configurator.RegisterMigration(tc.moduleName, i, func(sdk.Context) error {
						called++

						return nil
					})

					if tc.expRegErr {
						require.EqualError(tt, err, tc.expRegErrMsg)

						return
					}
					require.NoError(tt, err, "registering migration")
				}
			}

			// Run migrations only for bank. That's why we put the initial
			// version for bank as 1, and for all other modules, we put as
			// their latest ConsensusVersion.
			_, err = app.ModuleManager.RunMigrations(
				app.NewContextLegacy(true, cmtproto.Header{Height: app.LastBlockHeight()}), configurator,
				appmodule.VersionMap{
					"accounts":     accounts.AppModule{}.ConsensusVersion(),
					"bank":         1,
					"bankv2":       bankv2.AppModule{}.ConsensusVersion(),
					"auth":         auth.AppModule{}.ConsensusVersion(),
					"authz":        authzmodule.AppModule{}.ConsensusVersion(),
					"staking":      staking.AppModule{}.ConsensusVersion(),
					"mint":         mint.AppModule{}.ConsensusVersion(),
					"distribution": distribution.AppModule{}.ConsensusVersion(),
					"slashing":     slashing.AppModule{}.ConsensusVersion(),
					"gov":          gov.AppModule{}.ConsensusVersion(),
					"group":        group.AppModule{}.ConsensusVersion(),
					"upgrade":      upgrade.AppModule{}.ConsensusVersion(),
					"vesting":      vesting.AppModule{}.ConsensusVersion(),
					"feegrant":     feegrantmodule.AppModule{}.ConsensusVersion(),
					"evidence":     evidence.AppModule{}.ConsensusVersion(),
					"genutil":      genutil.AppModule{}.ConsensusVersion(),
					"protocolpool": protocolpool.AppModule{}.ConsensusVersion(),
					"epochs":       epochs.AppModule{}.ConsensusVersion(),
				},
			)
			if tc.expRunErr {
				require.EqualError(tt, err, tc.expRunErrMsg, "running migration")
			} else {
				require.NoError(tt, err, "running migration")
				// Make sure bank's migration is called.
				require.Equal(tt, tc.expCalled, called)
			}
		})
	}
}

func TestInitGenesisOnMigration(t *testing.T) {
	db := coretesting.NewMemDB()
	app := NewSimApp(log.NewTestLogger(t), db, nil, true, simtestutil.NewAppOptionsWithFlagHome(t.TempDir()))
	ctx := app.NewContextLegacy(true, cmtproto.Header{Height: app.LastBlockHeight()})

	// Create a mock module. This module will serve as the new module we're
	// adding during a migration.
	mockCtrl := gomock.NewController(t)
	t.Cleanup(mockCtrl.Finish)
	mockModule := mock.NewMockAppModuleWithAllExtensions(mockCtrl)
	mockDefaultGenesis := json.RawMessage(`{"key": "value"}`)
	mockModule.EXPECT().DefaultGenesis().Times(1).Return(mockDefaultGenesis)
	mockModule.EXPECT().InitGenesis(gomock.Eq(ctx), gomock.Eq(mockDefaultGenesis)).Times(1)
	mockModule.EXPECT().ConsensusVersion().Times(1).Return(uint64(0))

	app.ModuleManager.Modules["mock"] = mockModule

	// Run migrations only for "mock" module. We exclude it from
	// the VersionMap to simulate upgrading with a new module.
	_, err := app.ModuleManager.RunMigrations(ctx, app.Configurator(),
		appmodule.VersionMap{
			"bank":         bank.AppModule{}.ConsensusVersion(),
			"auth":         auth.AppModule{}.ConsensusVersion(),
			"authz":        authzmodule.AppModule{}.ConsensusVersion(),
			"staking":      staking.AppModule{}.ConsensusVersion(),
			"mint":         mint.AppModule{}.ConsensusVersion(),
			"distribution": distribution.AppModule{}.ConsensusVersion(),
			"slashing":     slashing.AppModule{}.ConsensusVersion(),
			"gov":          gov.AppModule{}.ConsensusVersion(),
			"upgrade":      upgrade.AppModule{}.ConsensusVersion(),
			"vesting":      vesting.AppModule{}.ConsensusVersion(),
			"feegrant":     feegrantmodule.AppModule{}.ConsensusVersion(),
			"evidence":     evidence.AppModule{}.ConsensusVersion(),
			"genutil":      genutil.AppModule{}.ConsensusVersion(),
		},
	)
	require.NoError(t, err)
}

func TestUpgradeStateOnGenesis(t *testing.T) {
	db := coretesting.NewMemDB()
	app := NewSimappWithCustomOptions(t, false, SetupOptions{
		Logger:  log.NewTestLogger(t),
		DB:      db,
		AppOpts: simtestutil.NewAppOptionsWithFlagHome(t.TempDir()),
	})

	// make sure the upgrade keeper has version map in state
	ctx := app.NewContext(false)
	vm, err := app.UpgradeKeeper.GetModuleVersionMap(ctx)
	require.NoError(t, err)
	for v, i := range app.ModuleManager.Modules {
		if i, ok := i.(appmodule.HasConsensusVersion); ok {
			require.Equal(t, vm[v], i.ConsensusVersion())
		}
	}
}

// TestMergedRegistry tests that fetching the gogo/protov2 merged registry
// doesn't fail after loading all file descriptors.
func TestMergedRegistry(t *testing.T) {
	r, err := proto.MergedRegistry()
	require.NoError(t, err)
	require.Greater(t, r.NumFiles(), 0)
}

func TestProtoAnnotations(t *testing.T) {
	r, err := proto.MergedRegistry()
	require.NoError(t, err)
	err = msgservice.ValidateProtoAnnotations(r)
	require.NoError(t, err)
}
