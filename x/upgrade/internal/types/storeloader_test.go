package types

import (
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/store/rootmulti"
	store "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"
	"io/ioutil"
	"os"
	"testing"
)

func useUpgradeLoader(upgrades *store.StoreUpgrades) func(*baseapp.BaseApp) {
	return func(app *baseapp.BaseApp) {
		app.SetStoreLoader(StoreLoaderWithUpgrade(upgrades))
	}
}

func useFileUpgradeLoader(upgradeInfoPath string) func(*baseapp.BaseApp) {
	return func(app *baseapp.BaseApp) {
		app.SetStoreLoader(UpgradeableStoreLoader(upgradeInfoPath))
	}
}

// TODO replace logger
func defaultLogger() log.Logger {
	return log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "sdk/app")
}

func initStore(t *testing.T, db dbm.DB, storeKey string, k, v []byte) {
	rs := rootmulti.NewStore(db)
	rs.SetPruning(store.PruneSyncable)
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
	rs.SetPruning(store.PruneSyncable)
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
	// write a rename to a file
	f, err := ioutil.TempFile("", "upgrade-*.json")
	require.NoError(t, err)
	data := []byte(`{"height": 0, "store_upgrades": {"renamed":[{"old_key": "bnk", "new_key": "banker"}]}}`)
	_, err = f.Write(data)
	require.NoError(t, err)
	configName := f.Name()
	require.NoError(t, f.Close())

	// make sure it exists before running everything
	_, err = os.Stat(configName)
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
			setLoader: useUpgradeLoader(&store.StoreUpgrades{
				Renamed: []store.StoreRename{{
					OldKey: "foo",
					NewKey: "bar",
				}},
			}),
			origStoreKey: "foo",
			loadStoreKey: "bar",
		},
		"file loader with missing file": {
			setLoader:    useFileUpgradeLoader(configName + "randomchars"),
			origStoreKey: "bnk",
			loadStoreKey: "bnk",
		},
		"file loader with existing file": {
			setLoader:    useFileUpgradeLoader(configName),
			origStoreKey: "bnk",
			loadStoreKey: "banker",
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
			opts := []func(*baseapp.BaseApp){baseapp.SetPruning(store.PruneSyncable)}
			if tc.setLoader != nil {
				opts = append(opts, tc.setLoader)
			}
			app := baseapp.NewBaseApp(t.Name(), defaultLogger(), db, nil, opts...)
			capKey := sdk.NewKVStoreKey(baseapp.MainStoreKey)
			app.MountStores(capKey)
			app.MountStores(sdk.NewKVStoreKey(tc.loadStoreKey))
			err := app.LoadLatestVersion(capKey)
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
