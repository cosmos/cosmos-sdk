package baseapp_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	tmprototypes "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/baseapp"
	pruningtypes "github.com/cosmos/cosmos-sdk/pruning/types"
	"github.com/cosmos/cosmos-sdk/snapshots"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestGetBlockRentionHeight(t *testing.T) {
	logger := defaultLogger()
	db := dbm.NewMemDB()
	name := t.Name()

	snapshotStore, err := snapshots.NewStore(dbm.NewMemDB(), testutil.GetTempDir(t))
	require.NoError(t, err)

	testCases := map[string]struct {
		bapp         *baseapp.BaseApp
		maxAgeBlocks int64
		commitHeight int64
		expected     int64
	}{
		"defaults": {
			bapp:         baseapp.NewBaseApp(name, logger, db),
			maxAgeBlocks: 0,
			commitHeight: 499000,
			expected:     0,
		},
		"pruning unbonding time only": {
			bapp:         baseapp.NewBaseApp(name, logger, db, baseapp.SetMinRetainBlocks(1)),
			maxAgeBlocks: 362880,
			commitHeight: 499000,
			expected:     136120,
		},
		"pruning iavl snapshot only": {
			bapp: baseapp.NewBaseApp(
				name, logger, db,
				baseapp.SetPruning(sdk.NewPruningOptions(pruningtypes.PruningNothing)),
				baseapp.SetMinRetainBlocks(1),
				baseapp.SetSnapshot(snapshotStore, sdk.NewSnapshotOptions(10000, 1)),
			),
			maxAgeBlocks: 0,
			commitHeight: 499000,
			expected:     489000,
		},
		"pruning state sync snapshot only": {
			bapp: baseapp.NewBaseApp(
				name, logger, db,
				baseapp.SetSnapshot(snapshotStore, sdk.NewSnapshotOptions(50000, 3)),
				baseapp.SetMinRetainBlocks(1),
			),
			maxAgeBlocks: 0,
			commitHeight: 499000,
			expected:     349000,
		},
		"pruning min retention only": {
			bapp: baseapp.NewBaseApp(
				name, logger, db,
				baseapp.SetMinRetainBlocks(400000),
			),
			maxAgeBlocks: 0,
			commitHeight: 499000,
			expected:     99000,
		},
		"pruning all conditions": {
			bapp: baseapp.NewBaseApp(
				name, logger, db,
				baseapp.SetPruning(sdk.NewCustomPruningOptions(0, 0)),
				baseapp.SetMinRetainBlocks(400000),
				baseapp.SetSnapshot(snapshotStore, sdk.NewSnapshotOptions(50000, 3)),
			),
			maxAgeBlocks: 362880,
			commitHeight: 499000,
			expected:     99000,
		},
		"no pruning due to no persisted state": {
			bapp: baseapp.NewBaseApp(
				name, logger, db,
				baseapp.SetPruning(sdk.NewCustomPruningOptions(0, 0)),
				baseapp.SetMinRetainBlocks(400000),
				baseapp.SetSnapshot(snapshotStore, sdk.NewSnapshotOptions(50000, 3)),
			),
			maxAgeBlocks: 362880,
			commitHeight: 10000,
			expected:     0,
		},
		"disable pruning": {
			bapp: baseapp.NewBaseApp(
				name, logger, db,
				baseapp.SetPruning(sdk.NewCustomPruningOptions(0, 0)),
				baseapp.SetMinRetainBlocks(0),
				baseapp.SetSnapshot(snapshotStore, sdk.NewSnapshotOptions(50000, 3)),
			),
			maxAgeBlocks: 362880,
			commitHeight: 499000,
			expected:     0,
		},
	}

	for name, tc := range testCases {
		tc := tc

		tc.bapp.SetParamStore(&paramStore{db: dbm.NewMemDB()})
		tc.bapp.InitChain(abci.RequestInitChain{
			ConsensusParams: &tmprototypes.ConsensusParams{
				Evidence: &tmprototypes.EvidenceParams{
					MaxAgeNumBlocks: tc.maxAgeBlocks,
				},
			},
		})

		t.Run(name, func(t *testing.T) {
			require.Equal(t, tc.expected, tc.bapp.GetBlockRetentionHeight(tc.commitHeight))
		})
	}
}

// Test and ensure that invalid block heights always cause errors.
// See issues:
// - https://github.com/cosmos/cosmos-sdk/issues/11220
// - https://github.com/cosmos/cosmos-sdk/issues/7662
func TestBaseAppCreateQueryContext(t *testing.T) {
	t.Parallel()

	logger := defaultLogger()
	db := dbm.NewMemDB()
	name := t.Name()
	app := baseapp.NewBaseApp(name, logger, db)

	app.BeginBlock(abci.RequestBeginBlock{Header: tmprototypes.Header{Height: 1}})
	app.Commit()

	app.BeginBlock(abci.RequestBeginBlock{Header: tmprototypes.Header{Height: 2}})
	app.Commit()

	testCases := []struct {
		name   string
		height int64
		prove  bool
		expErr bool
	}{
		{"valid height", 2, true, false},
		{"future height", 10, true, true},
		{"negative height, prove=true", -1, true, true},
		{"negative height, prove=false", -1, false, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := app.CreateQueryContext(tc.height, tc.prove)
			if tc.expErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
