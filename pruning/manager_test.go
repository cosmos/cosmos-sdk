package pruning_test

import (
	"container/list"
	"fmt"

	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	db "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/pruning"
	"github.com/cosmos/cosmos-sdk/pruning/types"
)

func TestNewManager(t *testing.T) {
	manager := pruning.NewManager(log.NewNopLogger())

	require.NotNil(t, manager)
	require.NotNil(t, manager.GetPruningHeights())
	require.Equal(t, types.PruningNothing, manager.GetOptions().GetPruningStrategy())
}

func TestStrategies(t *testing.T) {
	testcases := map[string]struct {
		strategy         *types.PruningOptions
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

	manager := pruning.NewManager(log.NewNopLogger())

	require.NotNil(t, manager)

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
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
			curInterval := curStrategy.Interval

			for curHeight := int64(0); curHeight < 110000; curHeight++ {
				handleHeightActual := manager.HandleHeight(curHeight)
				shouldPruneAtHeightActual := manager.ShouldPruneAtHeight(curHeight)

				curPruningHeihts := manager.GetPruningHeights()

				curHeightStr := fmt.Sprintf("height: %d", curHeight)

				switch curStrategy.GetPruningStrategy() {
				case types.PruningNothing:
					require.Equal(t, int64(0), handleHeightActual, curHeightStr)
					require.False(t, shouldPruneAtHeightActual, curHeightStr)

					require.Equal(t, 0, len(manager.GetPruningHeights()))
				default:
					if curHeight > int64(curKeepRecent) && (tc.snapshotInterval != 0 && (curHeight-int64(curKeepRecent))%int64(tc.snapshotInterval) != 0 || tc.snapshotInterval == 0) {
						expectedHeight := curHeight - int64(curKeepRecent)
						require.Equal(t, curHeight-int64(curKeepRecent), handleHeightActual, curHeightStr)

						require.Contains(t, curPruningHeihts, expectedHeight, curHeightStr)
					} else {
						require.Equal(t, int64(0), handleHeightActual, curHeightStr)

						require.Equal(t, 0, len(manager.GetPruningHeights()))
					}
					require.Equal(t, curHeight%int64(curInterval) == 0, shouldPruneAtHeightActual, curHeightStr)
				}
				manager.ResetPruningHeights()
				require.Equal(t, 0, len(manager.GetPruningHeights()))
			}
		})
	}
}

func TestHandleHeight_Inputs(t *testing.T) {
	var keepRecent int64 = int64(types.NewPruningOptions(types.PruningEverything).KeepRecent)

	testcases := map[string]struct {
		height int64
		expectedResult int64
		strategy types.PruningStrategy
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
			manager := pruning.NewManager(log.NewNopLogger())
			require.NotNil(t, manager)
			manager.SetOptions(types.NewPruningOptions(tc.strategy))

			handleHeightActual := manager.HandleHeight(tc.height)
			require.Equal(t, tc.expectedResult, handleHeightActual)

			require.Equal(t, tc.expectedHeights, manager.GetPruningHeights())
		})
	}
}

func TestFlushLoad(t *testing.T) {
	manager := pruning.NewManager(log.NewNopLogger())
	require.NotNil(t, manager)

	db := db.NewMemDB()

	curStrategy := types.NewCustomPruningOptions(100, 15)

	snapshotInterval := uint64(10)
	manager.SetSnapshotInterval(snapshotInterval)

	manager.SetOptions(curStrategy)
	require.Equal(t, curStrategy, manager.GetOptions())

	keepRecent := curStrategy.KeepRecent

	heightsToPruneMirror := make([]int64, 0)

	for curHeight := int64(0); curHeight < 1000; curHeight++ {
		handleHeightActual := manager.HandleHeight(curHeight)

		curHeightStr := fmt.Sprintf("height: %d", curHeight)

		if curHeight > int64(keepRecent) && (snapshotInterval != 0 && (curHeight-int64(keepRecent))%int64(snapshotInterval) != 0 || snapshotInterval == 0) {
			expectedHandleHeight := curHeight - int64(keepRecent)
			require.Equal(t, expectedHandleHeight, handleHeightActual, curHeightStr)
			heightsToPruneMirror = append(heightsToPruneMirror, expectedHandleHeight)
		} else {
			require.Equal(t, int64(0), handleHeightActual, curHeightStr)
		}

		if manager.ShouldPruneAtHeight(curHeight) {
			manager.ResetPruningHeights()
			heightsToPruneMirror = make([]int64, 0)
		}

		// N.B.: There is no reason behind the choice of 3.
		if curHeight%3 == 0 {
			require.Equal(t, heightsToPruneMirror, manager.GetPruningHeights(), curHeightStr)
			batch := db.NewBatch()
			manager.FlushPruningHeights(batch)
			require.NoError(t, batch.Write())
			require.NoError(t, batch.Close())

			manager.ResetPruningHeights()
			require.Equal(t, make([]int64, 0), manager.GetPruningHeights(), curHeightStr)

			err := manager.LoadPruningHeights(db)
			require.NoError(t, err)
			require.Equal(t, heightsToPruneMirror, manager.GetPruningHeights(), curHeightStr)
		}
	}
}

func TestLoadPruningHeights(t *testing.T) {
	var (
		manager = pruning.NewManager(log.NewNopLogger())
		err     error
	)
	require.NotNil(t, manager)

	// must not be PruningNothing
	manager.SetOptions(types.NewPruningOptions(types.PruningDefault))

	testcases := map[string]struct {
		flushedPruningHeights            []int64
		getFlushedPruningSnapshotHeights func() *list.List
		expectedResult                   error
	}{
		"negative pruningHeight - error": {
			flushedPruningHeights: []int64{10, 0, -1},
			expectedResult:        fmt.Errorf(pruning.ErrNegativeHeightsFmt, -1),
		},
		"negative snapshotPruningHeight - error": {
			getFlushedPruningSnapshotHeights: func() *list.List {
				l := list.New()
				l.PushBack(int64(5))
				l.PushBack(int64(-2))
				l.PushBack(int64(3))
				return l
			},
			expectedResult: fmt.Errorf(pruning.ErrNegativeHeightsFmt, -2),
		},
		"both have negative - pruningHeight error": {
			flushedPruningHeights: []int64{10, 0, -1},
			getFlushedPruningSnapshotHeights: func() *list.List {
				l := list.New()
				l.PushBack(int64(5))
				l.PushBack(int64(-2))
				l.PushBack(int64(3))
				return l
			},
			expectedResult: fmt.Errorf(pruning.ErrNegativeHeightsFmt, -1),
		},
		"both non-negative - success": {
			flushedPruningHeights: []int64{10, 0, 3},
			getFlushedPruningSnapshotHeights: func() *list.List {
				l := list.New()
				l.PushBack(int64(5))
				l.PushBack(int64(0))
				l.PushBack(int64(3))
				return l
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			db := db.NewMemDB()
			if tc.flushedPruningHeights != nil {
				err = db.Set(pruning.PruneHeightsKey, pruning.Int64SliceToBytes(tc.flushedPruningHeights))
				require.NoError(t, err)
			}

			if tc.getFlushedPruningSnapshotHeights != nil {
				err = db.Set(pruning.PruneSnapshotHeightsKey, pruning.ListToBytes(tc.getFlushedPruningSnapshotHeights()))
				require.NoError(t, err)
			}

			err = manager.LoadPruningHeights(db)
			require.Equal(t, tc.expectedResult, err)
		})
	}
}

func TestLoadPruningHeights_PruneNothing(t *testing.T) {
	var manager = pruning.NewManager(log.NewNopLogger())
	require.NotNil(t, manager)

	manager.SetOptions(types.NewPruningOptions(types.PruningNothing))

	require.Nil(t, manager.LoadPruningHeights(db.NewMemDB()))
}

func TestWithSnapshot(t *testing.T) {
	manager := pruning.NewManager(log.NewNopLogger())
	require.NotNil(t, manager)

	curStrategy := types.NewCustomPruningOptions(10, 10)

	snapshotInterval := uint64(15)
	manager.SetSnapshotInterval(snapshotInterval)

	manager.SetOptions(curStrategy)
	require.Equal(t, curStrategy, manager.GetOptions())

	keepRecent := curStrategy.KeepRecent

	heightsToPruneMirror := make([]int64, 0)

	mx := sync.Mutex{}
	snapshotHeightsToPruneMirror := list.New()

	wg := sync.WaitGroup{}
	defer wg.Wait()

	for curHeight := int64(1); curHeight < 100000; curHeight++ {
		mx.Lock()
		handleHeightActual := manager.HandleHeight(curHeight)

		curHeightStr := fmt.Sprintf("height: %d", curHeight)

		if curHeight > int64(keepRecent) && (curHeight-int64(keepRecent))%int64(snapshotInterval) != 0 {
			expectedHandleHeight := curHeight - int64(keepRecent)
			require.Equal(t, expectedHandleHeight, handleHeightActual, curHeightStr)
			heightsToPruneMirror = append(heightsToPruneMirror, expectedHandleHeight)
		} else {
			require.Equal(t, int64(0), handleHeightActual, curHeightStr)
		}

		actualHeightsToPrune := manager.GetPruningHeights()

		var next *list.Element
		for e := snapshotHeightsToPruneMirror.Front(); e != nil; e = next {
			snapshotHeight := e.Value.(int64)
			if snapshotHeight < curHeight-int64(keepRecent) {
				heightsToPruneMirror = append(heightsToPruneMirror, snapshotHeight)

				// We must get next before removing to be able to continue iterating.
				next = e.Next()
				snapshotHeightsToPruneMirror.Remove(e)
			} else {
				next = e.Next()
			}
		}

		require.Equal(t, heightsToPruneMirror, actualHeightsToPrune, curHeightStr)
		mx.Unlock()

		if manager.ShouldPruneAtHeight(curHeight) {
			manager.ResetPruningHeights()
			heightsToPruneMirror = make([]int64, 0)
		}

		// Mimic taking snapshots in the background
		if curHeight%int64(snapshotInterval) == 0 {
			wg.Add(1)
			go func(curHeightCp int64) {
				time.Sleep(time.Nanosecond * 500)

				mx.Lock()
				manager.HandleHeightSnapshot(curHeightCp)
				snapshotHeightsToPruneMirror.PushBack(curHeightCp)
				mx.Unlock()
				wg.Done()
			}(curHeight)
		}
	}
}
