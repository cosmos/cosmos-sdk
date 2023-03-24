package pruning_test

import (
	"errors"
	"fmt"
	"testing"

	"cosmossdk.io/log"
	db "github.com/cosmos/cosmos-db"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/store/mock"
	"cosmossdk.io/store/pruning"
	"cosmossdk.io/store/pruning/types"
)

const dbErr = "db error"

func TestNewManager(t *testing.T) {
	manager := pruning.NewManager(db.NewMemDB(), log.NewNopLogger())
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
		tc := tc // Local copy to avoid shadowing.
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			manager := pruning.NewManager(db.NewMemDB(), log.NewNopLogger())
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

			for curHeight := int64(0); curHeight < 110000; curHeight++ {
				if tc.snapshotInterval != 0 {
					if curHeight > int64(tc.snapshotInterval) && curHeight%int64(tc.snapshotInterval) == int64(tc.snapshotInterval)-1 {
						manager.HandleHeightSnapshot(curHeight - int64(tc.snapshotInterval) + 1)
					}
				}

				handleHeightActual := manager.GetPruningHeight(curHeight)
				curHeightStr := fmt.Sprintf("height: %d", curHeight)

				switch curStrategy.GetPruningStrategy() {
				case types.PruningNothing:
					require.Equal(t, int64(0), handleHeightActual, curHeightStr)
				default:
					if curHeight > int64(curKeepRecent) && (tc.snapshotInterval != 0 && (curHeight-int64(curKeepRecent))%int64(tc.snapshotInterval) != 0 || tc.snapshotInterval == 0) {
						require.Equal(t, curHeight-int64(curKeepRecent), handleHeightActual, curHeightStr)
					} else {
						require.Equal(t, int64(0), handleHeightActual, curHeightStr)
					}
				}
			}
		})
	}
}

func TestHandleHeight_Inputs(t *testing.T) {
	keepRecent := int64(types.NewPruningOptions(types.PruningEverything).KeepRecent)

	testcases := map[string]struct {
		height          int64
		expectedResult  int64
		strategy        types.PruningStrategy
		expectedHeights []int64
	}{
		"previousHeight is negative - prune everything - invalid previousHeight": {
			-1,
			0,
			types.PruningEverything,
			[]int64{},
		},
		"previousHeight is  zero - prune everything - invalid previousHeight": {
			0,
			0,
			types.PruningEverything,
			[]int64{},
		},
		"previousHeight is positive but within keep recent- prune everything - not kept": {
			keepRecent,
			0,
			types.PruningEverything,
			[]int64{},
		},
		"previousHeight is positive and greater than keep recent - kept": {
			keepRecent + 1,
			keepRecent + 1 - keepRecent,
			types.PruningEverything,
			[]int64{keepRecent + 1 - keepRecent},
		},
		"pruning nothing, previousHeight is positive and greater than keep recent - not kept": {
			keepRecent + 1,
			0,
			types.PruningNothing,
			[]int64{},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			manager := pruning.NewManager(db.NewMemDB(), log.NewNopLogger())
			require.NotNil(t, manager)
			manager.SetOptions(types.NewPruningOptions(tc.strategy))

			handleHeightActual := manager.GetPruningHeight(tc.height)
			require.Equal(t, tc.expectedResult, handleHeightActual)
		})
	}
}

func TestHandleHeight_FlushLoadFromDisk(t *testing.T) {
	testcases := map[string]struct {
		previousHeight                   int64
		keepRecent                       uint64
		snapshotInterval                 uint64
		movedSnapshotHeights             []int64
		expectedHandleHeightResult       int64
		expectedLoadPruningHeightsResult error
		expectedLoadedHeights            []int64
	}{
		"simple flush occurs": {
			previousHeight:                   11,
			keepRecent:                       10,
			snapshotInterval:                 0,
			movedSnapshotHeights:             []int64{},
			expectedHandleHeightResult:       11 - 10,
			expectedLoadPruningHeightsResult: nil,
			expectedLoadedHeights:            []int64{11 - 10},
		},
		"previous height <= keep recent - no update and no flush": {
			previousHeight:                   9,
			keepRecent:                       10,
			snapshotInterval:                 0,
			movedSnapshotHeights:             []int64{},
			expectedHandleHeightResult:       0,
			expectedLoadPruningHeightsResult: nil,
			expectedLoadedHeights:            []int64{},
		},
		"previous height alligns with snapshot interval - no update and no flush": {
			previousHeight:                   12,
			keepRecent:                       10,
			snapshotInterval:                 2,
			movedSnapshotHeights:             []int64{},
			expectedHandleHeightResult:       0,
			expectedLoadPruningHeightsResult: nil,
			expectedLoadedHeights:            []int64{},
		},
		"previous height does not align with snapshot interval - flush": {
			previousHeight:                   12,
			keepRecent:                       10,
			snapshotInterval:                 3,
			movedSnapshotHeights:             []int64{},
			expectedHandleHeightResult:       2,
			expectedLoadPruningHeightsResult: nil,
			expectedLoadedHeights:            []int64{2},
		},
		"moved snapshot heights - flushed": {
			previousHeight:                   32,
			keepRecent:                       10,
			snapshotInterval:                 5,
			movedSnapshotHeights:             []int64{15, 20, 25},
			expectedHandleHeightResult:       22,
			expectedLoadPruningHeightsResult: nil,
			expectedLoadedHeights:            []int64{15, 20, 22},
		},
		"previous height alligns with snapshot interval - no update but flush snapshot heights": {
			previousHeight:                   30,
			keepRecent:                       10,
			snapshotInterval:                 5,
			movedSnapshotHeights:             []int64{15, 20, 25},
			expectedHandleHeightResult:       0,
			expectedLoadPruningHeightsResult: nil,
			expectedLoadedHeights:            []int64{15},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			// Setup
			db := db.NewMemDB()
			manager := pruning.NewManager(db, log.NewNopLogger())
			require.NotNil(t, manager)

			manager.SetSnapshotInterval(tc.snapshotInterval)
			manager.SetOptions(types.NewCustomPruningOptions(tc.keepRecent, uint64(10)))

			for _, snapshotHeight := range tc.movedSnapshotHeights {
				manager.HandleHeightSnapshot(snapshotHeight)
			}

			// Test HandleHeight and flush
			handleHeightActual := manager.GetPruningHeight(tc.previousHeight)
			require.Equal(t, tc.expectedHandleHeightResult, handleHeightActual)

			loadedPruneHeights, err := pruning.LoadPruningSnapshotHeights(db)
			require.NoError(t, err)
			require.Equal(t, len(loadedPruneHeights), len(tc.movedSnapshotHeights))
		})
	}
}

func TestHandleSnapshotHeight_DbErr_Panic(t *testing.T) {
	ctrl := gomock.NewController(t)

	// Setup
	dbMock := mock.NewMockDB(ctrl)

	dbMock.EXPECT().SetSync(gomock.Any(), gomock.Any()).Return(errors.New(dbErr)).Times(1)

	manager := pruning.NewManager(dbMock, log.NewNopLogger())
	manager.SetOptions(types.NewPruningOptions(types.PruningEverything))
	require.NotNil(t, manager)

	defer func() {
		if r := recover(); r == nil {
			t.Fail()
		}
	}()

	manager.HandleHeightSnapshot(10)
}

func TestHandleHeightSnapshot_FlushLoadFromDisk(t *testing.T) {
	loadedHeightsMirror := []int64{}

	// Setup
	db := db.NewMemDB()
	manager := pruning.NewManager(db, log.NewNopLogger())
	require.NotNil(t, manager)

	manager.SetOptions(types.NewPruningOptions(types.PruningEverything))

	for snapshotHeight := int64(-1); snapshotHeight < 100; snapshotHeight++ {
		// Test flush
		manager.HandleHeightSnapshot(snapshotHeight)

		// Post test
		if snapshotHeight > 0 {
			loadedHeightsMirror = append(loadedHeightsMirror, snapshotHeight)
		}

		loadedSnapshotHeights, err := pruning.LoadPruningSnapshotHeights(db)
		require.NoError(t, err)
		require.Equal(t, len(loadedHeightsMirror), len(loadedSnapshotHeights))

		// Test load back
		err = manager.LoadSnapshotHeights(db)
		require.NoError(t, err)

		loadedSnapshotHeights, err = pruning.LoadPruningSnapshotHeights(db)
		require.NoError(t, err)
		require.Equal(t, len(loadedHeightsMirror), len(loadedSnapshotHeights))
	}
}

func TestHandleHeightSnapshot_DbErr_Panic(t *testing.T) {
	ctrl := gomock.NewController(t)

	// Setup
	dbMock := mock.NewMockDB(ctrl)

	dbMock.EXPECT().SetSync(gomock.Any(), gomock.Any()).Return(errors.New(dbErr)).Times(1)

	manager := pruning.NewManager(dbMock, log.NewNopLogger())
	manager.SetOptions(types.NewPruningOptions(types.PruningEverything))
	require.NotNil(t, manager)

	defer func() {
		if r := recover(); r == nil {
			t.Fail()
		}
	}()

	manager.HandleHeightSnapshot(10)
}

func TestFlushLoad(t *testing.T) {
	db := db.NewMemDB()
	manager := pruning.NewManager(db, log.NewNopLogger())
	require.NotNil(t, manager)

	curStrategy := types.NewCustomPruningOptions(100, 15)

	snapshotInterval := uint64(10)
	manager.SetSnapshotInterval(snapshotInterval)

	manager.SetOptions(curStrategy)
	require.Equal(t, curStrategy, manager.GetOptions())

	keepRecent := curStrategy.KeepRecent

	heightsToPruneMirror := make([]int64, 0)

	for curHeight := int64(0); curHeight < 1000; curHeight++ {
		handleHeightActual := manager.GetPruningHeight(curHeight)

		curHeightStr := fmt.Sprintf("height: %d", curHeight)

		if curHeight > int64(keepRecent) && (snapshotInterval != 0 && (curHeight-int64(keepRecent))%int64(snapshotInterval) != 0 || snapshotInterval == 0) {
			expectedHandleHeight := curHeight - int64(keepRecent)
			require.Equal(t, expectedHandleHeight, handleHeightActual, curHeightStr)
			heightsToPruneMirror = append(heightsToPruneMirror, expectedHandleHeight)
		} else {
			require.Equal(t, int64(0), handleHeightActual, curHeightStr)
		}

		if curHeight > int64(keepRecent) {
			actualHeight := manager.GetPruningHeight(curHeight)
			// require.Equal(t, len(heightsToPruneMirror), len(actualHeights))
			require.Equal(t, heightsToPruneMirror, []int64{actualHeight})

			err := manager.LoadSnapshotHeights(db)
			require.NoError(t, err)

			actualHeight = manager.GetPruningHeight(curHeight)
			// require.Equal(t, len(heightsToPruneMirror), len(actualHeights))
			require.Equal(t, heightsToPruneMirror, []int64{actualHeight})

			heightsToPruneMirror = make([]int64, 0)
		}
	}
}

func TestLoadPruningHeights(t *testing.T) {
	var (
		manager = pruning.NewManager(db.NewMemDB(), log.NewNopLogger())
		err     error
	)
	require.NotNil(t, manager)

	// must not be PruningNothing
	manager.SetOptions(types.NewPruningOptions(types.PruningDefault))

	testcases := map[string]struct {
		flushedPruningHeights            []int64
		getFlushedPruningSnapshotHeights func() []int64
		expectedResult                   error
	}{
		"negative pruningHeight - error": {
			flushedPruningHeights: []int64{10, 0, -1},
			expectedResult:        &pruning.NegativeHeightsError{Height: -1},
		},
		"negative snapshotPruningHeight - error": {
			getFlushedPruningSnapshotHeights: func() []int64 {
				return []int64{5, -2, 3}
			},
			expectedResult: &pruning.NegativeHeightsError{Height: -2},
		},
		"both have negative - pruningHeight error": {
			flushedPruningHeights: []int64{10, 0, -1},
			getFlushedPruningSnapshotHeights: func() []int64 {
				return []int64{5, -2, 3}
			},
			expectedResult: &pruning.NegativeHeightsError{Height: -1},
		},
		"both non-negative - success": {
			flushedPruningHeights: []int64{10, 0, 3},
			getFlushedPruningSnapshotHeights: func() []int64 {
				return []int64{5, 0, 3}
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			db := db.NewMemDB()

			if tc.getFlushedPruningSnapshotHeights != nil {
				err = db.Set(pruning.PruneSnapshotHeightsKey, pruning.Int64SliceToBytes(tc.getFlushedPruningSnapshotHeights()))
				require.NoError(t, err)
			}

			err = manager.LoadSnapshotHeights(db)
			require.Equal(t, tc.expectedResult, err)
		})
	}
}

func TestLoadSnapshotHeights_PruneNothing(t *testing.T) {
	manager := pruning.NewManager(db.NewMemDB(), log.NewNopLogger())
	require.NotNil(t, manager)

	manager.SetOptions(types.NewPruningOptions(types.PruningNothing))

	require.Nil(t, manager.LoadSnapshotHeights(db.NewMemDB()))
}
