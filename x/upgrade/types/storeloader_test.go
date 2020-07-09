package types

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/store/rootmulti"
	store "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func useUpgradeLoader(height int64, upgrades *store.StoreUpgrades) func(*baseapp.BaseApp) {
	return func(app *baseapp.BaseApp) {
		app.SetStoreLoader(UpgradeStoreLoader(height, upgrades))
	}
}

func defaultLogger() log.Logger {
	return log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "sdk/app")
}

func initStore(t *testing.T, db dbm.DB, storeKey string, k, v []byte) {
	rs := rootmulti.NewStore(db)
	rs.SetPruning(store.PruneNothing)
	key := sdk.NewKVStoreKey(storeKey)
	rs.MountStoreWithDB(key, store.StoreTypeIAVL, nil)
	err := rs.LoadLatestVersion()
	require.Nil(t, err)
	require.Equal(t, int64(0), rs.LastCommitID().Version)

	// write some data in substore
	kv, _ := rs.GetStore(key).(store.KVStore)
	require.NotNil(t, kv)
	kv.Set(k, v)
	commitID := rs.Commit()
	require.Equal(t, int64(1), commitID.Version)
}

func checkStore(t *testing.T, db dbm.DB, ver int64, storeKey string, k, v []byte) {
	rs := rootmulti.NewStore(db)
	rs.SetPruning(store.PruneNothing)
	key := sdk.NewKVStoreKey(storeKey)
	rs.MountStoreWithDB(key, store.StoreTypeIAVL, nil)
	err := rs.LoadLatestVersion()
	require.Nil(t, err)
	require.Equal(t, ver, rs.LastCommitID().Version)

	// query data in substore
	kv, _ := rs.GetStore(key).(store.KVStore)

	require.NotNil(t, kv)
	require.Equal(t, v, kv.Get(k))
}

// Test that we can make commits and then reload old versions.
// Test that LoadLatestVersion actually does.
func TestSetLoader(t *testing.T) {
	// set a temporary home dir
	homeDir, cleanup := testutil.NewTestCaseDir(t)
	t.Cleanup(cleanup)

	upgradeInfoFilePath := filepath.Join(homeDir, "upgrade-info.json")
	upgradeInfo := &store.UpgradeInfo{
		Name: "test", Height: 0,
	}
	data, err := json.Marshal(upgradeInfo)
	require.NoError(t, err)
	err = ioutil.WriteFile(upgradeInfoFilePath, data, 0644)
	require.NoError(t, err)

	// make sure it exists before running everything
	_, err = os.Stat(upgradeInfoFilePath)
	require.NoError(t, err)

	cases := map[string]struct {
		setLoader    func(*baseapp.BaseApp)
		origStoreKey string
		loadStoreKey string
	}{
		"don't set loader": {
			origStoreKey: "foo",
			loadStoreKey: "foo",
		},
		"rename with inline opts": {
			setLoader: useUpgradeLoader(0, &store.StoreUpgrades{
				Renamed: []store.StoreRename{{
					OldKey: "foo",
					NewKey: "bar",
				}},
			}),
			origStoreKey: "foo",
			loadStoreKey: "bar",
		},
	}

	k := []byte("key")
	v := []byte("value")

	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			// prepare a db with some data
			db := dbm.NewMemDB()

			initStore(t, db, tc.origStoreKey, k, v)

			// load the app with the existing db
			opts := []func(*baseapp.BaseApp){baseapp.SetPruning(store.PruneNothing)}
			if tc.setLoader != nil {
				opts = append(opts, tc.setLoader)
			}

			app := baseapp.NewBaseApp(t.Name(), defaultLogger(), db, nil, opts...)
			capKey := sdk.NewKVStoreKey("main")
			app.MountStores(capKey)
			app.MountStores(sdk.NewKVStoreKey(tc.loadStoreKey))
			err := app.LoadLatestVersion()
			require.Nil(t, err)

			// "execute" one block
			app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: 2}})
			res := app.Commit()
			require.NotNil(t, res.Data)

			// check db is properly updated
			checkStore(t, db, 2, tc.loadStoreKey, k, v)
			checkStore(t, db, 2, tc.loadStoreKey, []byte("foo"), nil)
		})
	}
}
