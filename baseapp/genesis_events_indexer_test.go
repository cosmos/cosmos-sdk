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

// TestABCI_GenesisTxEvents_ReachFinalizeBlockResponse asserts the outcome
// issue #25984 asks for: events from genesis txs are visible to indexers.
// CometBFT's indexer service only reads ResponseFinalizeBlock.Events and
// ExecTxResult.Events (state/txindex/indexer_service.go), so the genesis-tx
// event must show up in the first block's FinalizeBlock response — parking
// it on the finalize-state EventManager is not observable by any indexer.
func TestABCI_GenesisTxEvents_ReachFinalizeBlockResponse(t *testing.T) {
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

	deliverKey := []byte("genesis-counter")
	baseapptestutil.RegisterCounterServer(
		suite.baseApp.MsgServiceRouter(),
		CounterServerImpl{t, capKey1, deliverKey},
	)

	// Build a tx whose handler emits a known event (message/update_counter).
	tx := newTxCounter(t, suite.txConfig, 0, 0)
	var err error
	txBytes, err = suite.txConfig.TxEncoder()(tx)
	require.NoError(t, err)

	_, err = suite.baseApp.InitChain(&abci.RequestInitChain{
		ChainId:         "test-chain-id",
		ConsensusParams: &cmtproto.ConsensusParams{},
		AppStateBytes:   []byte("{}"),
	})
	require.NoError(t, err)

	res, err := suite.baseApp.FinalizeBlock(&abci.RequestFinalizeBlock{Height: 1})
	require.NoError(t, err)

	var sawCounterEvent bool
	for _, ev := range res.Events {
		if ev.Type != sdk.EventTypeMessage {
			continue
		}
		for _, attr := range ev.Attributes {
			if attr.Key == "update_counter" && attr.Value == "0" {
				sawCounterEvent = true
			}
		}
	}
	require.True(t, sawCounterEvent,
		"genesis-tx event not in ResponseFinalizeBlock.Events of the first block — invisible to CometBFT indexers; events=%+v", res.Events)
}
