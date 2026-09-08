package baseapp_test

import (
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/baseapp"
	baseapptestutil "github.com/cosmos/cosmos-sdk/baseapp/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TestABCI_GenesisTxEvents_DoNotDuplicateOnRestart guards against a
// standalone-node restart scenario: with --with-comet=false, CometBFT can
// restart while the app still reports height 0 and before the first
// FinalizeBlock succeeds. The handshake then calls InitChain again, which
// re-runs the genesis txs. Without resetting app.genesisEvents at the start
// of InitChain, the buffer from the first InitChain call survives and the
// second call's ExecuteGenesisTx appends to it, so the genesis event would
// appear twice in the first FinalizeBlock response.
func TestABCI_GenesisTxEvents_DoNotDuplicateOnRestart(t *testing.T) {
	var (
		txBytes []byte
		appRef  *baseapp.BaseApp
	)

	initChainerOpt := func(bapp *baseapp.BaseApp) {
		bapp.SetInitChainer(func(_ sdk.Context, _ *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
			require.NoError(t, appRef.ExecuteGenesisTx(txBytes))
			return &abci.ResponseInitChain{}, nil
		})
	}

	suite := NewBaseAppSuite(t, baseapp.SetChainID("test-chain-id"), initChainerOpt)
	appRef = suite.baseApp

	deliverKey := []byte("genesis-counter-double-initchain")
	baseapptestutil.RegisterCounterServer(
		suite.baseApp.MsgServiceRouter(),
		CounterServerImpl{t, capKey1, deliverKey},
	)

	tx := newTxCounter(t, suite.txConfig, 0, 0)
	var err error
	txBytes, err = suite.txConfig.TxEncoder()(tx)
	require.NoError(t, err)

	initReq := &abci.RequestInitChain{
		ChainId:         "test-chain-id",
		ConsensusParams: &cmtproto.ConsensusParams{},
		AppStateBytes:   []byte("{}"),
	}

	// First InitChain: normal genesis start.
	_, err = suite.baseApp.InitChain(initReq)
	require.NoError(t, err)

	// Simulate a CometBFT restart at height 0, before the first
	// FinalizeBlock ever succeeds: InitChain is called again.
	_, err = suite.baseApp.InitChain(initReq)
	require.NoError(t, err)

	res, err := suite.baseApp.FinalizeBlock(&abci.RequestFinalizeBlock{Height: 1})
	require.NoError(t, err)

	var count int
	for _, ev := range res.Events {
		if ev.Type != sdk.EventTypeMessage {
			continue
		}
		for _, attr := range ev.Attributes {
			if attr.Key == "update_counter" && attr.Value == "0" {
				count++
			}
		}
	}
	require.Equal(t, 1, count,
		"genesis-tx event must appear exactly once even after a repeated InitChain; events=%+v", res.Events)
}
