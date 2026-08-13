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

// TestABCI_GenesisTxEvents_DoNotLeakPastFirstBlock guards the other half of
// the #25984 contract: genesis-tx events must appear exactly once, in the
// first block's FinalizeBlock response, and must not reappear in later
// blocks (which would indicate app.genesisEvents was never cleared).
func TestABCI_GenesisTxEvents_DoNotLeakPastFirstBlock(t *testing.T) {
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

	deliverKey := []byte("genesis-counter-leak")
	baseapptestutil.RegisterCounterServer(
		suite.baseApp.MsgServiceRouter(),
		CounterServerImpl{t, capKey1, deliverKey},
	)

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

	res1, err := suite.baseApp.FinalizeBlock(&abci.RequestFinalizeBlock{Height: 1})
	require.NoError(t, err)
	require.True(t, containsGenesisCounterEvent(res1.Events),
		"genesis event missing from first block: %+v", res1.Events)

	_, err = suite.baseApp.Commit()
	require.NoError(t, err)

	res2, err := suite.baseApp.FinalizeBlock(&abci.RequestFinalizeBlock{Height: 2})
	require.NoError(t, err)
	require.False(t, containsGenesisCounterEvent(res2.Events),
		"genesis event leaked into second block: %+v", res2.Events)
}

func containsGenesisCounterEvent(events []abci.Event) bool {
	for _, ev := range events {
		if ev.Type != sdk.EventTypeMessage {
			continue
		}
		for _, attr := range ev.Attributes {
			if attr.Key == "update_counter" && attr.Value == "0" {
				return true
			}
		}
	}
	return false
}
