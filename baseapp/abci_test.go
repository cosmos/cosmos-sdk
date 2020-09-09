package baseapp

import (
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	tmprototypes "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestGetBlockRentionHeight(t *testing.T) {
	logger := defaultLogger()
	db := dbm.NewMemDB()
	name := t.Name()

	testCases := map[string]struct {
		bapp         *BaseApp
		maxAgeBlocks int64
		commitHeight int64
		expected     int64
	}{
		"defaults": {
			bapp:         NewBaseApp(name, logger, db, nil),
			maxAgeBlocks: 0,
			commitHeight: 499000,
			expected:     0,
		},
		"pruning unbonding time only": {
			bapp:         NewBaseApp(name, logger, db, nil, SetMinRetainBlocks(1)),
			maxAgeBlocks: 362880,
			commitHeight: 499000,
			expected:     136120,
		},
		"pruning iavl snapshot only": {
			bapp: NewBaseApp(
				name, logger, db, nil,
				SetPruning(sdk.PruningOptions{Interval: 10000}),
				SetMinRetainBlocks(1),
			),
			maxAgeBlocks: 0,
			commitHeight: 499000,
			expected:     489000,
		},
		"pruning state sync snapshot only": {
			bapp: NewBaseApp(
				name, logger, db, nil,
				SetSnapshotInterval(50000),
				SetSnapshotKeepRecent(3),
				SetMinRetainBlocks(1),
			),
			maxAgeBlocks: 0,
			commitHeight: 499000,
			expected:     349000,
		},
		"pruning min retention only": {
			bapp: NewBaseApp(
				name, logger, db, nil,
				SetMinRetainBlocks(400000),
			),
			maxAgeBlocks: 0,
			commitHeight: 499000,
			expected:     99000,
		},
		"pruning all conditions": {
			bapp: NewBaseApp(
				name, logger, db, nil,
				SetPruning(sdk.PruningOptions{Interval: 10000}),
				SetMinRetainBlocks(400000),
				SetSnapshotInterval(50000), SetSnapshotKeepRecent(3),
			),
			maxAgeBlocks: 362880,
			commitHeight: 499000,
			expected:     99000,
		},
		"disable pruning": {
			bapp: NewBaseApp(
				name, logger, db, nil,
				SetPruning(sdk.PruningOptions{Interval: 10000}),
				SetMinRetainBlocks(0),
				SetSnapshotInterval(50000), SetSnapshotKeepRecent(3),
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
			ConsensusParams: &abci.ConsensusParams{
				Evidence: &tmprototypes.EvidenceParams{
					MaxAgeNumBlocks: tc.maxAgeBlocks,
				},
			},
		})

		t.Run(name, func(t *testing.T) {
			require.Equal(t, tc.expected, tc.bapp.GetBlockRentionHeight(tc.commitHeight))
		})
	}
}
