package pruning

import (
	"errors"
	"fmt"
	"testing"

	db "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"cosmossdk.io/log"
	"cosmossdk.io/store/mock"
	"cosmossdk.io/store/pruning/types"
)

const dbErr = "db error"

func TestNewManager(t *testing.T) {
	manager := NewManager(db.NewMemDB(), log.NewNopLogger())
	require.NotNil(t, manager)
	require.Equal(t, types.PruningNothing, manager.GetOptions().GetPruningStrategy())
}

func TestStrategies(t *testing.T) {
	testcases := map[string]struct {
		strategy         types.PruningOptions
		snapshotInterval uint64
		strategyToAssert types.PruningStrategy
		isValid          bool
	}{
		"prune nothing - no snapshot": {
			strategy:         types.NewPruningOptions(types.PruningNothing),
			strategyToAssert: types.PruningNothing,
		},
		"prune nothing - snapshot": {
			strategy:         types.NewPruningOptions(types.PruningNothing),
			strategyToAssert: types.PruningNothing,
			snapshotInterval: 100,
		},
		"prune default - no snapshot": {
			strategy:         types.NewPruningOptions(types.PruningDefault),
			strategyToAssert: types.PruningDefault,
		},
		"prune default - snapshot": {
			strategy:         types.NewPruningOptions(types.PruningDefault),
			strategyToAssert: types.PruningDefault,
			snapshotInterval: 100,
		},
		"prune everything - no snapshot": {
			strategy:         types.NewPruningOptions(types.PruningEverything),
			strategyToAssert: types.PruningEverything,
		},
		"prune everything - snapshot": {
			strategy:         types.NewPruningOptions(types.PruningEverything),
			strategyToAssert: types.PruningEverything,
			snapshotInterval: 100,
		},
		"custom 100-10-15": {
			strategy:         types.NewCustomPruningOptions(100, 15),
			snapshotInterval: 10,
			strategyToAssert: types.PruningCustom,
		},
		"custom 10-10-15": {
			strategy:         types.NewCustomPruningOptions(10, 15),
			snapshotInterval: 10,
			strategyToAssert: types.PruningCustom,
		},
		"custom 100-0-15": {
			strategy:         types.NewCustomPruningOptions(100, 15),
			snapshotInterval: 0,
			strategyToAssert: types.PruningCustom,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			manager := NewManager(db.NewMemDB(), log.NewNopLogger())
			require.NotNil(t, manager)

			curStrategy := tc.strategy
			manager.SetSnapshotInterval(tc.snapshotInterval)

			pruneStrategy := curStrategy.GetPruningStrategy()
			require.Equal(t, tc.strategyToAssert, pruneStrategy)

			// Validate strategy parameters
			switch pruneStrategy {
			case types.PruningDefault:
				require.Equal(t, uint64(362880), curStrategy.KeepRecent)
				require.Equal(t, uint64(10), curStrategy.Interval)
			case types.PruningNothing:
				require.Equal(t, uint64(0), curStrategy.KeepRecent)
				require.Equal(t, uint64(0), curStrategy.Interval)
			case types.PruningEverything:
				require.Equal(t, uint64(2), curStrategy.KeepRecent)
				require.Equal(t, uint64(10), curStrategy.Interval)
			default:
				//
			}

			manager.SetOptions(curStrategy)
			require.Equal(t, tc.strategy, manager.GetOptions())

			curKeepRecent := curStrategy.KeepRecent
			snHeight := int64(tc.snapshotInterval - 1)
			for curHeight := int64(0); curHeight < 110000; curHeight++ {
				if tc.snapshotInterval != 0 {
					if curHeight > int64(tc.snapshotInterval) && curHeight%int64(tc.snapshotInterval) == int64(tc.snapshotInterval)-1 {
						snapHeight := curHeight - int64(tc.snapshotInterval) + 1
						manager.AnnounceSnapshotHeight(snapHeight)
						manager.HandleSnapshotHeight(snapHeight)
						snHeight = curHeight
					}
				}

				pruningHeightActual := manager.GetPruningHeight(curHeight)
				curHeightStr := fmt.Sprintf("height: %d", curHeight)

				switch curStrategy.GetPruningStrategy() {
				case types.PruningNothing:
					require.Equal(t, int64(0), pruningHeightActual, curHeightStr)
				default:
					if curHeight > int64(curKeepRecent) && curHeight%int64(curStrategy.Interval) == 0 {
						pruningHeightExpected := curHeight - int64(curKeepRecent) - 1
						if tc.snapshotInterval > 0 && snHeight < pruningHeightExpected {
							pruningHeightExpected = snHeight
						}
						require.Equal(t, pruningHeightExpected, pruningHeightActual, curHeightStr)
					} else {
						require.Equal(t, int64(0), pruningHeightActual, curHeightStr)
					}
				}
			}
		})
	}
}

func TestPruningHeight_Inputs(t *testing.T) {
	keepRecent := int64(types.NewPruningOptions(types.PruningEverything).KeepRecent)
	interval := int64(types.NewPruningOptions(types.PruningEverything).Interval)

	testcases := map[string]struct {
		height         int64
		expectedResult int64
		strategy       types.PruningStrategy
	}{
		"currentHeight is negative - prune everything - invalid currentHeight": {
			-1,
			0,
			types.PruningEverything,
		},
		"currentHeight is  zero - prune everything - invalid currentHeight": {
			0,
			0,
			types.PruningEverything,
		},
		"currentHeight is positive but within keep recent- prune everything - not kept": {
			keepRecent,
			0,
			types.PruningEverything,
		},
		"currentHeight is positive and equal to keep recent+1 - no kept": {
			keepRecent + 1,
			0,
			types.PruningEverything,
		},
		"currentHeight is positive and greater than keep recent+1 but not multiple of interval - no kept": {
			keepRecent + 2,
			0,
			types.PruningEverything,
		},
		"currentHeight is positive and greater than keep recent+1 and multiple of interval - kept": {
			interval,
			interval - keepRecent - 1,
			types.PruningEverything,
		},
		"pruning nothing, currentHeight is positive and greater than keep recent - not kept": {
			interval,
			0,
			types.PruningNothing,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			manager := NewManager(db.NewMemDB(), log.NewNopLogger())
			manager.SetOptions(types.NewPruningOptions(tc.strategy))

			pruningHeightActual := manager.GetPruningHeight(tc.height)
			require.Equal(t, tc.expectedResult, pruningHeightActual)
		})
	}
}

func TestGetPruningHeight(t *testing.T) {
	specs := map[string]struct {
		initDBState int64
		opts        types.PruningOptions
		setup       func(manager *Manager)
		exp         map[int64]int64
	}{
		"init from store - no snap": {
			initDBState: 10,
			opts:        types.PruningOptions{KeepRecent: 5, Interval: 10, Strategy: types.PruningCustom},
			setup: func(mgr *Manager) {
				mgr.SetSnapshotInterval(15)
			},
			exp: map[int64]int64{
				20: 9, // initDBState - 1
				30: 9, // initDBState - 1
				45: 0, // not a prune height
			},
		},
		"init from store - snap landed": {
			initDBState: 10,
			opts:        types.PruningOptions{KeepRecent: 5, Interval: 10, Strategy: types.PruningCustom},
			setup: func(mgr *Manager) {
				mgr.SetSnapshotInterval(15)
				mgr.AnnounceSnapshotHeight(15)
				mgr.HandleSnapshotHeight(15)
			},
			exp: map[int64]int64{
				10: 4,  // 10 - 5 (keep) - 1
				15: 0,  // not on prune interval
				20: 14, // 20 - 5 (keep) - 1
				30: 24, // 30 - 5 (keep) - 1
				40: 29, // 15 (last completed snap) + 15 (snap interval) - 1
			},
		},
		"init from store - snap in-flight": {
			initDBState: 10,
			opts:        types.PruningOptions{KeepRecent: 5, Interval: 10, Strategy: types.PruningCustom},
			setup: func(mgr *Manager) {
				mgr.SetSnapshotInterval(15)
				mgr.AnnounceSnapshotHeight(15)
			},
			exp: map[int64]int64{
				10: 4, // 10 - 5 (keep) - 1
				20: 9, // 10 - 5 (keep) - 1
			},
		},
		"init from store - delayed in-flight snap": {
			initDBState: 10,
			opts:        types.PruningOptions{KeepRecent: 5, Interval: 10, Strategy: types.PruningCustom},
			setup: func(mgr *Manager) {
				mgr.SetSnapshotInterval(15)
				mgr.AnnounceSnapshotHeight(15)
				mgr.AnnounceSnapshotHeight(30)
				mgr.HandleSnapshotHeight(30)
			},
			exp: map[int64]int64{
				10: 4,  // 10 - 5 (keep) - 1
				20: 14, // 15 (in-flight) - 1
				30: 14, // 15 (in-flight) - 1
				40: 14, // 15 (in-flight) - 1
			},
		},
		"empty store": {
			opts: types.PruningOptions{KeepRecent: 5, Interval: 10, Strategy: types.PruningCustom},
			setup: func(mgr *Manager) {
				mgr.SetSnapshotInterval(15)
			},
			exp: map[int64]int64{
				10: 4,  // 10 -5 (keep) -1
				20: 14, // 20 -5 (keep) -1
			},
		},
		"empty snap interval set": {
			initDBState: 10,
			opts:        types.PruningOptions{KeepRecent: 5, Interval: 10, Strategy: types.PruningCustom},
			setup:       func(mgr *Manager) {},
			exp: map[int64]int64{
				10: 4,  // 10 -5 (keep) -1
				20: 14, // 20 -5 (keep) -1
			},
		},
		"prune nothing set": {
			initDBState: 10,
			opts:        types.PruningOptions{Strategy: types.PruningNothing, Interval: 10, KeepRecent: 5},
			setup: func(mgr *Manager) {
				mgr.SetSnapshotInterval(15)
			},
			exp: map[int64]int64{
				10: 0, // nothing
				20: 0, // nothing
				30: 0, // nothing
			},
		},
		"empty prune interval": {
			initDBState: 10,
			opts:        types.PruningOptions{Strategy: types.PruningCustom, KeepRecent: 5},
			setup: func(mgr *Manager) {
				mgr.SetSnapshotInterval(15)
			},
			exp: map[int64]int64{
				10: 0, // interval required
				20: 0, // interval required
				30: 0, // interval required
			},
		},
		"height <= keep": {
			initDBState: 10,
			opts:        types.PruningOptions{Strategy: types.PruningCustom, Interval: 1, KeepRecent: 5},
			setup: func(mgr *Manager) {
				mgr.SetSnapshotInterval(15)
			},
			exp: map[int64]int64{
				0: 0, // interval required
				4: 0, // interval required
				5: 0, // interval required
			},
		},
		"height not on prune interval": {
			initDBState: 10,
			opts:        types.PruningOptions{Strategy: types.PruningCustom, Interval: 2},
			setup: func(mgr *Manager) {
				mgr.SetSnapshotInterval(15)
			},
			exp: map[int64]int64{
				0: 0, // excluded
				1: 0, // not on prune interval
				2: 1, // 2 - 1
				3: 0, // not on prune interval
				4: 3, // 2 - 1
			},
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			memDB := db.NewMemDB()
			if spec.initDBState != 0 {
				require.NoError(t, storePruningSnapshotHeight(memDB, spec.initDBState))
			}
			mgr2 := NewManager(memDB, log.NewNopLogger())
			mgr2.SetOptions(spec.opts)
			require.NoError(t, mgr2.LoadSnapshotHeights(memDB))
			spec.setup(mgr2)

			for height, exp := range spec.exp {
				gotHeight := mgr2.GetPruningHeight(height)
				assert.Equal(t, exp, gotHeight, "height: %d", height)
			}
		})
	}
}

func TestHandleSnapshotHeight_DbErr_Panic(t *testing.T) {
	ctrl := gomock.NewController(t)

	// Setup
	dbMock := mock.NewMockDB(ctrl)

	dbMock.EXPECT().SetSync(gomock.Any(), gomock.Any()).Return(errors.New(dbErr)).Times(1)

	manager := NewManager(dbMock, log.NewNopLogger())
	manager.SetSnapshotInterval(1)
	manager.SetOptions(types.NewPruningOptions(types.PruningEverything))
	require.NotNil(t, manager)

	defer func() {
		if r := recover(); r == nil {
			t.Fail()
		}
	}()

	manager.HandleSnapshotHeight(10)
}

func TestHandleSnapshotHeight_LoadFromDisk(t *testing.T) {
	snapshotInterval := uint64(10)

	// Setup
	db := db.NewMemDB()
	manager := NewManager(db, log.NewNopLogger())
	require.NotNil(t, manager)

	manager.SetOptions(types.NewPruningOptions(types.PruningEverything))
	manager.SetSnapshotInterval(snapshotInterval)

	expected := 0
	for snapshotHeight := int64(-1); snapshotHeight < 100; snapshotHeight++ {
		snapshotHeightStr := fmt.Sprintf("snaphost height: %d", snapshotHeight)
		if snapshotHeight > int64(snapshotInterval) && snapshotHeight%int64(snapshotInterval) == 1 {
			// Test flush
			manager.HandleSnapshotHeight(snapshotHeight - 1)
			expected = 1
		}

		loadedSnapshotHeights, err := loadPruningSnapshotHeights(db)
		require.NoError(t, err)
		require.Equal(t, expected, len(loadedSnapshotHeights), snapshotHeightStr)

		// Test load back
		err = manager.LoadSnapshotHeights(db)
		require.NoError(t, err)

		loadedSnapshotHeights, err = loadPruningSnapshotHeights(db)
		require.NoError(t, err)
		require.Equal(t, expected, len(loadedSnapshotHeights), snapshotHeightStr)
	}
}

func TestLoadPruningSnapshotHeights(t *testing.T) {
	var (
		manager = NewManager(db.NewMemDB(), log.NewNopLogger())
		err     error
	)
	require.NotNil(t, manager)

	// must not be PruningNothing
	manager.SetOptions(types.NewPruningOptions(types.PruningDefault))

	testcases := map[string]struct {
		getFlushedPruningSnapshotHeights func() []int64
		expectedResult                   error
	}{
		"negative snapshotPruningHeight - error": {
			getFlushedPruningSnapshotHeights: func() []int64 {
				return []int64{5, -2, 3}
			},
			expectedResult: &NegativeHeightsError{Height: -2},
		},
		"non-negative - success": {
			getFlushedPruningSnapshotHeights: func() []int64 {
				return []int64{5, 0, 3}
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			db := db.NewMemDB()

			if tc.getFlushedPruningSnapshotHeights != nil {
				err = db.Set(pruneSnapshotHeightsKey, int64SliceToBytes(tc.getFlushedPruningSnapshotHeights()...))
				require.NoError(t, err)
			}

			err = manager.LoadSnapshotHeights(db)
			require.Equal(t, tc.expectedResult, err)
		})
	}
}

func TestLoadSnapshotHeights_PruneNothing(t *testing.T) {
	manager := NewManager(db.NewMemDB(), log.NewNopLogger())
	require.NotNil(t, manager)

	manager.SetOptions(types.NewPruningOptions(types.PruningNothing))

	require.Nil(t, manager.LoadSnapshotHeights(db.NewMemDB()))
}
