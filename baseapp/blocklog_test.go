package baseapp_test

import (
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/log/v2"

	"github.com/cosmos/cosmos-sdk/baseapp"
)

// hookLogger is a logger that also exposes the block log hooks, like
// server/blocklog.Logger does.
type hookLogger struct {
	log.Logger
	calls *[]string
}

func (h hookLogger) BeginBlockLog(height int64) {
	*h.calls = append(*h.calls, "begin:"+string(rune('0'+height)))
}
func (h hookLogger) CommitBlockLog() { *h.calls = append(*h.calls, "commit") }

func TestBaseApp_BlockLogHooksBracketFinalizeBlockAndCommit(t *testing.T) {
	logger := hookLogger{Logger: log.NewTestLogger(t), calls: new([]string)}
	app := baseapp.NewBaseApp(t.Name(), logger, dbm.NewMemDB(), nil)

	_, err := app.InitChain(&abci.RequestInitChain{InitialHeight: 1})
	require.NoError(t, err)
	require.Empty(t, *logger.calls, "nothing outside a block")

	_, err = app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: 1})
	require.NoError(t, err)
	require.Equal(t, []string{"begin:1"}, *logger.calls, "height stays open until Commit")

	_, err = app.Commit()
	require.NoError(t, err)
	require.Equal(t, []string{"begin:1", "commit"}, *logger.calls)

	_, err = app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: 2})
	require.NoError(t, err)
	_, err = app.Commit()
	require.NoError(t, err)
	require.Equal(t, []string{"begin:1", "commit", "begin:2", "commit"}, *logger.calls)
}

func TestBaseApp_PlainLoggerHasNoBlockLogHooks(t *testing.T) {
	app := baseapp.NewBaseApp(t.Name(), log.NewTestLogger(t), dbm.NewMemDB(), nil)
	_, err := app.InitChain(&abci.RequestInitChain{InitialHeight: 1})
	require.NoError(t, err)
	_, err = app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: 1})
	require.NoError(t, err)
	_, err = app.Commit()
	require.NoError(t, err)
}

// With optimistic execution the block body runs during ProcessProposal, so
// the height must open there and stay open through FinalizeBlock; an aborted
// proposal reopens (truncates) the height before re-execution.
func TestBaseApp_BlockLogHooksWithOptimisticExecution(t *testing.T) {
	logger := hookLogger{Logger: log.NewTestLogger(t), calls: new([]string)}
	app := baseapp.NewBaseApp(t.Name(), logger, dbm.NewMemDB(), nil, baseapp.SetOptimisticExecution())

	_, err := app.InitChain(&abci.RequestInitChain{InitialHeight: 1})
	require.NoError(t, err)
	_, err = app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: 1})
	require.NoError(t, err)
	_, err = app.Commit()
	require.NoError(t, err)
	require.Equal(t, []string{"begin:1", "commit"}, *logger.calls, "first block is never executed optimistically")

	// height 2: the accepted proposal is executed optimistically and consumed by FinalizeBlock
	resp, err := app.ProcessProposal(&abci.RequestProcessProposal{Height: 2, Hash: []byte("h2")})
	require.NoError(t, err)
	require.Equal(t, abci.ResponseProcessProposal_ACCEPT, resp.Status)
	require.Equal(t, []string{"begin:1", "commit", "begin:2"}, *logger.calls, "height opens when optimistic execution starts")
	_, err = app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: 2, Hash: []byte("h2")})
	require.NoError(t, err)
	require.Equal(t, []string{"begin:1", "commit", "begin:2"}, *logger.calls, "consuming the optimistic result must not reopen the height")
	_, err = app.Commit()
	require.NoError(t, err)

	// height 3: a different block is finalized, the optimistic run is discarded and the height reopened
	_, err = app.ProcessProposal(&abci.RequestProcessProposal{Height: 3, Hash: []byte("h3-a")})
	require.NoError(t, err)
	_, err = app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: 3, Hash: []byte("h3-b")})
	require.NoError(t, err)
	_, err = app.Commit()
	require.NoError(t, err)
	require.Equal(t, []string{"begin:1", "commit", "begin:2", "commit", "begin:3", "begin:3", "commit"}, *logger.calls)
}
