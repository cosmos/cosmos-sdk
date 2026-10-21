package baseapp_test

import (
	"context"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/baseapp"
	baseapptestutil "github.com/cosmos/cosmos-sdk/baseapp/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// relayServer handles MsgCounter by dispatching a MsgCounter2, the way a
// precompile or another module runs a msg it does not own.
type relayServer struct{ router msgservice.MsgRouter }

func (s relayServer) IncrementCounter(ctx context.Context, msg *baseapptestutil.MsgCounter) (*baseapptestutil.MsgCreateCounterResponse, error) {
	return msgservice.Dispatch[*baseapptestutil.MsgCreateCounterResponse](sdk.UnwrapSDKContext(ctx), s.router, &baseapptestutil.MsgCounter2{Counter: msg.Counter})
}

// TestDispatchChecksCircuitBreakerInEveryMode trips MsgCounter2 and
// dispatches it from a tx in CheckTx, Simulate and FinalizeBlock, from the
// block hooks, from a vote extension and from a query context. The router
// handler checks the circuit breaker on each call whatever the exec mode,
// so every dispatch fails.
func TestDispatchChecksCircuitBreakerInEveryMode(t *testing.T) {
	counter2URL := sdk.MsgTypeURL(&baseapptestutil.MsgCounter2{})
	breaker := &typeBreaker{blocked: map[string]bool{counter2URL: true}}
	tripped := "circuit breaker disables execution of this message: " + counter2URL

	var app *baseapp.BaseApp
	errs := map[string]error{}
	dispatch := func(caller string, ctx sdk.Context) {
		_, errs[caller] = msgservice.Dispatch[*baseapptestutil.MsgCreateCounterResponse](ctx, app.MsgServiceRouter(), &baseapptestutil.MsgCounter2{Counter: 1})
	}
	modeNames := map[sdk.ExecMode]string{
		sdk.ExecModeCheck:    "ante/check",
		sdk.ExecModeSimulate: "ante/simulate",
		sdk.ExecModeFinalize: "ante/finalize",
	}

	suite := NewBaseAppSuite(t, func(bapp *baseapp.BaseApp) {
		bapp.SetCircuitBreaker(breaker)
		bapp.SetPreBlocker(func(ctx sdk.Context, _ *abci.RequestFinalizeBlock) (*sdk.ResponsePreBlock, error) {
			dispatch("preblock", ctx)
			return &sdk.ResponsePreBlock{}, nil
		})
		bapp.SetBeginBlocker(func(ctx sdk.Context) (sdk.BeginBlock, error) {
			dispatch("beginblock", ctx)
			return sdk.BeginBlock{}, nil
		})
		bapp.SetEndBlocker(func(ctx sdk.Context) (sdk.EndBlock, error) {
			dispatch("endblock", ctx)
			return sdk.EndBlock{}, nil
		})
		bapp.SetAnteHandler(func(ctx sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) {
			dispatch(modeNames[ctx.ExecMode()], ctx)
			return ctx, nil
		})
		bapp.SetExtendVoteHandler(func(ctx sdk.Context, _ *abci.RequestExtendVote) (*abci.ResponseExtendVote, error) {
			dispatch("extendvote", ctx)
			return &abci.ResponseExtendVote{}, nil
		})
	})
	app = suite.baseApp
	baseapptestutil.RegisterCounterServer(app.MsgServiceRouter(), relayServer{router: app.MsgServiceRouter()})
	baseapptestutil.RegisterCounter2Server(app.MsgServiceRouter(), Counter2ServerImpl{t, capKey1, []byte("foo")})

	_, err := app.InitChain(&abci.RequestInitChain{
		ConsensusParams: &cmtproto.ConsensusParams{Abci: &cmtproto.ABCIParams{VoteExtensionsEnableHeight: 1}},
	})
	require.NoError(t, err)

	txBytes, err := suite.txConfig.TxEncoder()(newTxCounter(t, suite.txConfig, 0, 1))
	require.NoError(t, err)

	checkRes, err := app.CheckTx(&abci.RequestCheckTx{Tx: txBytes, Type: abci.CheckTxType_New})
	require.NoError(t, err)
	require.Zero(t, checkRes.Code, checkRes.Log)

	_, _, err = app.Simulate(txBytes)
	require.ErrorContains(t, err, tripped)

	res, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: 1, Txs: [][]byte{txBytes}})
	require.NoError(t, err)
	require.NotZero(t, res.TxResults[0].Code)
	require.Contains(t, res.TxResults[0].Log, tripped)
	_, err = app.Commit()
	require.NoError(t, err)

	_, err = app.ExtendVote(context.Background(), &abci.RequestExtendVote{Height: 2})
	require.NoError(t, err)

	queryCtx, err := app.CreateQueryContext(1, false)
	require.NoError(t, err)
	dispatch("query", queryCtx)

	for _, caller := range []string{
		"ante/check", "ante/simulate", "ante/finalize",
		"preblock", "beginblock", "endblock", "extendvote", "query",
	} {
		require.ErrorContains(t, errs[caller], tripped, caller)
	}
}
