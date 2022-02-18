package types

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	dbm "github.com/cosmos/cosmos-sdk/db"
	"github.com/cosmos/cosmos-sdk/db/memdb"
	"github.com/cosmos/cosmos-sdk/server"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/store/v2/multi"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func defaultLogger() log.Logger {
	writer := zerolog.ConsoleWriter{Out: os.Stderr}
	return server.ZeroLogWrapper{
		Logger: zerolog.New(writer).Level(zerolog.InfoLevel).With().Timestamp().Logger(),
	}
}

func initStore(t *testing.T, db dbm.DBConnection, config multi.StoreParams, key storetypes.StoreKey, k, v []byte) {
	rs, err := multi.NewV1MultiStoreAsV2(db, config)
	require.NoError(t, err)
	rs.SetPruning(storetypes.PruneNothing)
	require.Equal(t, int64(0), rs.LastCommitID().Version)

	// write some data in substore
	kv := rs.GetKVStore(key)
	require.NotNil(t, kv)
	kv.Set(k, v)
	commitID := rs.Commit()
	require.Equal(t, int64(1), commitID.Version)
	require.NoError(t, rs.Close())
}

func checkStore(t *testing.T, db dbm.DBConnection, config multi.StoreParams, ver int64, key storetypes.StoreKey, k, v []byte) {
	rs, err := multi.NewV1MultiStoreAsV2(db, config)
	require.NoError(t, err)
	rs.SetPruning(storetypes.PruneNothing)
	require.Equal(t, ver, rs.LastCommitID().Version)

	if v != nil {
		kv := rs.GetKVStore(key)
		require.NotNil(t, kv)
		require.Equal(t, v, kv.Get(k))
	} else {
		// v == nil indicates the substore was moved and no longer exists
		require.Panics(t, func() { _ = rs.GetKVStore(key) })
	}
	require.NoError(t, rs.Close())
}

// Test that we can make commits and then reload old versions.
// Test that LoadLatestVersion actually does.
func TestSetLoader(t *testing.T) {
	upgradeHeight := int64(5)

	// set a temporary home dir
	homeDir := t.TempDir()
	upgradeInfoFilePath := filepath.Join(homeDir, UpgradeInfoFilename)
	upgradeInfo := &Plan{
		Name: "test", Height: upgradeHeight,
	}

	data, err := json.Marshal(upgradeInfo)
	require.NoError(t, err)

	err = os.WriteFile(upgradeInfoFilePath, data, 0644)
	require.NoError(t, err)

	// make sure it exists before running everything
	_, err = os.Stat(upgradeInfoFilePath)
	require.NoError(t, err)

	fooKey := sdk.NewKVStoreKey("foo")
	barKey := sdk.NewKVStoreKey("bar")
	cases := map[string]struct {
		setLoader    baseapp.AppOption
		origStoreKey storetypes.StoreKey
		loadStoreKey storetypes.StoreKey
	}{
		"don't set loader": {
			setLoader:    nil,
			origStoreKey: fooKey,
			loadStoreKey: fooKey,
		},
		"rename with inline opts": {
			setLoader: UpgradeStoreOption(uint64(upgradeHeight), &storetypes.StoreUpgrades{
				Renamed: []storetypes.StoreRename{{
					OldKey: "foo",
					NewKey: "bar",
				}},
			}),
			origStoreKey: fooKey,
			loadStoreKey: barKey,
		},
	}

	k := []byte("key")
	v := []byte("value")

	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			origConfig := multi.DefaultStoreParams()
			loadConfig := multi.DefaultStoreParams()
			require.NoError(t, origConfig.RegisterSubstore(tc.origStoreKey, storetypes.StoreTypePersistent))
			require.NoError(t, loadConfig.RegisterSubstore(tc.loadStoreKey, storetypes.StoreTypePersistent))

			// prepare a db with some data
			db := memdb.NewDB()
			initStore(t, db, origConfig, tc.origStoreKey, k, v)

			// load the app with the existing db
			opts := []baseapp.AppOption{
				baseapp.SetPruning(storetypes.PruneNothing),
				baseapp.SetSubstores(tc.origStoreKey),
			}
			origapp := baseapp.NewBaseApp(t.Name(), defaultLogger(), db, opts...)
			require.NoError(t, origapp.Init())

			for i := int64(2); i <= upgradeHeight-1; i++ {
				origapp.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: i}})
				res := origapp.Commit()
				require.NotNil(t, res.Data)
			}
			require.NoError(t, origapp.CloseStore())

			// load the new app with the original app db
			opts = []baseapp.AppOption{
				baseapp.SetPruning(storetypes.PruneNothing),
				baseapp.SetSubstores(tc.loadStoreKey),
			}
			if tc.setLoader != nil {
				opts = append(opts, tc.setLoader)
			}
			app := baseapp.NewBaseApp(t.Name(), defaultLogger(), db, opts...)
			require.NoError(t, app.Init())

			// "execute" one block
			app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: upgradeHeight}})
			res := app.Commit()
			require.NotNil(t, res.Data)
			require.NoError(t, app.CloseStore())

			// checking the case of the store being renamed
			if tc.setLoader != nil {
				checkStore(t, db, loadConfig, upgradeHeight, tc.origStoreKey, k, nil)
			}

			// check db is properly updated
			checkStore(t, db, loadConfig, upgradeHeight, tc.loadStoreKey, k, v)
		})
	}
}
