package simapp

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
)

func TestSimAppExportAndBlockedAddrs(t *testing.T) {
	encCfg := MakeTestEncodingConfig()
	db := dbm.NewMemDB()
	app := NewSimApp(log.NewTMLogger(log.NewSyncWriter(os.Stdout)), db, nil, true, map[int64]bool{}, DefaultNodeHome, 0, encCfg, EmptyAppOptions{})

	for acc := range maccPerms {
		require.Equal(t, !allowedReceivingModAcc[acc], app.BankKeeper.BlockedAddr(app.AccountKeeper.GetModuleAddress(acc)),
			"ensure that blocked addresses are properly set in bank keeper")
	}

	genesisState := NewDefaultGenesisState(encCfg.Marshaler)
	stateBytes, err := json.MarshalIndent(genesisState, "", "  ")
	require.NoError(t, err)

	// Initialize the chain
	app.InitChain(
		abci.RequestInitChain{
			Validators:    []abci.ValidatorUpdate{},
			AppStateBytes: stateBytes,
		},
	)
	app.Commit()

	// Making a new app object with the db, so that initchain hasn't been called
	app2 := NewSimApp(log.NewTMLogger(log.NewSyncWriter(os.Stdout)), db, nil, true, map[int64]bool{}, DefaultNodeHome, 0, encCfg, EmptyAppOptions{})
	_, err = app2.ExportAppStateAndValidators(false, []string{})
	require.NoError(t, err, "ExportAppStateAndValidators should not have an error")
}

func TestGetMaccPerms(t *testing.T) {
	dup := GetMaccPerms()
	require.Equal(t, maccPerms, dup, "duplicated module account permissions differed from actual module account permissions")
}

func TestRunMigrations(t *testing.T) {
	db := dbm.NewMemDB()
	app := NewSimApp(log.NewTMLogger(log.NewSyncWriter(os.Stdout)), db, nil, true, map[int64]bool{}, DefaultNodeHome, 0, MakeTestEncodingConfig(), EmptyAppOptions{})

	// Create a new configurator for the purpose of this test.
	app.configurator = module.NewConfigurator(app.MsgServiceRouter(), app.GRPCQueryRouter(), app.keys)

	testCases := []struct {
		name         string
		moduleName   string
		expRegErr    bool // errors while registering migration
		expRegErrMsg string
		expRunErr    bool // errors while running migration
		expRunErrMsg string
		expCalled    int
	}{
		{
			"cannot register migration for non-existant module",
			"foo",
			true, "store key for module foo not found: not found", false, "", 0,
		},
		{
			"throws error on RunMigrations if no migration registered for bank",
			"",
			false, "", true, "no migration found for module bank from version 0 to version 1: not found", 0,
		},
		{
			"can register and run migration handler for x/bank",
			"bank",
			false, "", false, "", 1,
		},
		{
			"cannot register migration handler for same module & fromVersion",
			"bank",
			true, "another migration for module bank and version 0 already exists: internal logic error", false, "", 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var err error

			// Since it's very hard to test actual in-place store migrations in
			// tests (due to the difficulty of maintaing multiple versions of a
			// module), we're just testing here that the migration logic is
			// called.
			called := 0

			if tc.moduleName != "" {
				err = app.configurator.RegisterMigration(tc.moduleName, 0, func(sdk.KVStore) error {
					called++

					return nil
				})
			if tc.expRegErr {
				assert.ErrorEqual(t, err, tc.expRegErrMsg)
			}
			}
			require.NoError(t, err)

			err = app.RunMigrations(
				app.NewContext(true, tmproto.Header{Height: app.LastBlockHeight()}),
				map[string]uint64{"bank": 0},
			)
			if tc.expRunErr {
				require.Error(t, err)
				require.Equal(t, tc.expRunErrMsg, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expCalled, called)
			}
		})
	}
}
