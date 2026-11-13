package baseapp_test

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/baseapp"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type queryVersionHookStore struct {
	storetypes.MultiStore
	beforeRead func()
}

func (s queryVersionHookStore) LatestVersion() int64 {
	s.beforeRead()
	return s.MultiStore.LatestVersion()
}

func newQueryCommitApp(t *testing.T) *baseapp.BaseApp {
	t.Helper()
	suite := NewBaseAppSuite(t, func(app *baseapp.BaseApp) {
		app.SetBeginBlocker(func(ctx sdk.Context) (sdk.BeginBlock, error) {
			ctx.KVStore(capKey1).Set([]byte("height"), []byte(fmt.Sprint(ctx.BlockHeight())))
			return sdk.BeginBlock{}, nil
		})
	})
	app := suite.baseApp
	_, err := app.InitChain(&abci.RequestInitChain{
		InitialHeight:   1,
		ConsensusParams: &cmtproto.ConsensusParams{},
	})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, app.Close()) })
	return app
}

func commitQueryBlock(t *testing.T, app *baseapp.BaseApp) {
	t.Helper()
	height := app.LastBlockHeight() + 1
	_, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: height,
		Time:   time.Unix(height, 0).UTC(),
	})
	require.NoError(t, err)
	_, err = app.Commit()
	require.NoError(t, err)
}

func TestCreateQueryContextDuringCommit(t *testing.T) {
	for _, commits := range []int{1, 3} {
		t.Run(fmt.Sprintf("commits=%d", commits), func(t *testing.T) {
			app := newQueryCommitApp(t)
			commitQueryBlock(t, app)
			reads, remaining := 0, commits
			app.SetQueryMultiStore(queryVersionHookStore{
				MultiStore: app.CommitMultiStore(),
				beforeRead: func() {
					reads++
					if reads > 1 && remaining > 0 {
						remaining--
						// Commit after the query has captured the previous states.
						// The check state advances and the finalize state is cleared.
						commitQueryBlock(t, app)
					}
				},
			})

			ctx, err := app.CreateQueryContextWithCheckHeader(0, false, true)
			require.NoError(t, err)
			require.Zero(t, remaining)
			height := int64(commits + 1)
			require.Equal(t, height, ctx.BlockHeight())
			require.Equal(t, time.Unix(height, 0).UTC(), ctx.BlockTime())
			require.Equal(t, []byte(fmt.Sprint(height)), ctx.KVStore(capKey1).Get([]byte("height")))
		})
	}
}

func TestCreateQueryContextConcurrentCommit(t *testing.T) {
	app := newQueryCommitApp(t)
	commitQueryBlock(t, app)
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	t.Cleanup(func() {
		cancel()
		wg.Wait()
	})
	const readers = 8
	ready := make(chan struct{}, readers)
	errors := make(chan error, readers)
	var queries atomic.Uint64
	for range readers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			first := true
			for ctx.Err() == nil {
				query, err := app.CreateQueryContextWithCheckHeader(0, false, true)
				if first {
					ready <- struct{}{}
					first = false
				}
				if err != nil {
					errors <- err
					return
				}
				height := query.BlockHeight()
				value := query.KVStore(capKey1).Get([]byte("height"))
				if query.BlockTime().Unix() != height || string(value) != fmt.Sprint(height) {
					errors <- fmt.Errorf("inconsistent query at height %d: time=%v value=%s", height, query.BlockTime(), value)
					return
				}
				queries.Add(1)
			}
		}()
	}
	for range readers {
		<-ready
	}
	for range 100 {
		commitQueryBlock(t, app)
	}
	cancel()
	wg.Wait()
	close(errors)
	for err := range errors {
		require.NoError(t, err)
	}
	require.Greater(t, queries.Load(), uint64(0))
}
