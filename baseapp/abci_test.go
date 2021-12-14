package baseapp_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	tmprototypes "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestGetBlockRentionHeight(t *testing.T) {
	logger := defaultLogger()
	db := dbm.NewMemDB()
	name := t.Name()

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
				baseapp.SetPruning(sdk.PruningOptions{KeepEvery: 10000}),
				baseapp.SetMinRetainBlocks(1),
			),
			maxAgeBlocks: 0,
			commitHeight: 499000,
			expected:     490000,
		},
		"pruning state sync snapshot only": {
			bapp: baseapp.NewBaseApp(
				name, logger, db,
				baseapp.SetSnapshotInterval(50000),
				baseapp.SetSnapshotKeepRecent(3),
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
				baseapp.SetPruning(sdk.PruningOptions{KeepEvery: 10000}),
				baseapp.SetMinRetainBlocks(400000),
				baseapp.SetSnapshotInterval(50000), baseapp.SetSnapshotKeepRecent(3),
			),
			maxAgeBlocks: 362880,
			commitHeight: 499000,
			expected:     99000,
		},
		"no pruning due to no persisted state": {
			bapp: baseapp.NewBaseApp(
				name, logger, db,
				baseapp.SetPruning(sdk.PruningOptions{KeepEvery: 10000}),
				baseapp.SetMinRetainBlocks(400000),
				baseapp.SetSnapshotInterval(50000), baseapp.SetSnapshotKeepRecent(3),
			),
			maxAgeBlocks: 362880,
			commitHeight: 10000,
			expected:     0,
		},
		"disable pruning": {
			bapp: baseapp.NewBaseApp(
				name, logger, db,
				baseapp.SetPruning(sdk.PruningOptions{KeepEvery: 10000}),
				baseapp.SetMinRetainBlocks(0),
				baseapp.SetSnapshotInterval(50000), baseapp.SetSnapshotKeepRecent(3),
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

// Test and ensure that negative heights always cause errors.
// See issue https://github.com/cosmos/cosmos-sdk/issues/7662.
func TestBaseAppCreateQueryContextRejectsNegativeHeights(t *testing.T) {
	t.Parallel()

	logger := defaultLogger()
	db := dbm.NewMemDB()
	name := t.Name()
	app := baseapp.NewBaseApp(name, logger, db)

	proves := []bool{
		false, true,
	}
	for _, prove := range proves {
		t.Run(fmt.Sprintf("prove=%t", prove), func(t *testing.T) {
			sctx, err := app.CreateQueryContext(-10, true)
			require.Error(t, err)
			require.Equal(t, sctx, sdk.Context{})
		})
	}
}
