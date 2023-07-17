package baseapp_test

import (
	"context"
	"fmt"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"

	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	baseapptestutil "github.com/cosmos/cosmos-sdk/baseapp/testutil"
)

var _ storetypes.ABCIListener = (*MockABCIListener)(nil)

type MockABCIListener struct {
	name      string
	ChangeSet []*storetypes.StoreKVPair
}

func NewMockABCIListener(name string) MockABCIListener {
	return MockABCIListener{
		name:      name,
		ChangeSet: make([]*storetypes.StoreKVPair, 0),
	}
}

func (m MockABCIListener) ListenFinalizeBlock(_ context.Context, _ abci.RequestFinalizeBlock, _ abci.ResponseFinalizeBlock) error {
	return nil
}

func (m *MockABCIListener) ListenCommit(_ context.Context, _ abci.ResponseCommit, cs []*storetypes.StoreKVPair) error {
	m.ChangeSet = cs
	return nil
}

var distKey1 = storetypes.NewKVStoreKey("distKey1")

func TestABCI_MultiListener_StateChanges(t *testing.T) {
	anteKey := []byte("ante-key")
	anteOpt := func(bapp *baseapp.BaseApp) { bapp.SetAnteHandler(anteHandlerTxTest(t, capKey1, anteKey)) }
	distOpt := func(bapp *baseapp.BaseApp) { bapp.MountStores(distKey1) }
	mockListener1 := NewMockABCIListener("lis_1")
	mockListener2 := NewMockABCIListener("lis_2")
	streamingManager := storetypes.StreamingManager{ABCIListeners: []storetypes.ABCIListener{&mockListener1, &mockListener2}}
	streamingManagerOpt := func(bapp *baseapp.BaseApp) { bapp.SetStreamingManager(streamingManager) }
	addListenerOpt := func(bapp *baseapp.BaseApp) { bapp.CommitMultiStore().AddListeners([]storetypes.StoreKey{distKey1}) }
	suite := NewBaseAppSuite(t, anteOpt, distOpt, streamingManagerOpt, addListenerOpt)

	_, err := suite.baseApp.InitChain(
		&abci.RequestInitChain{
			ConsensusParams: &tmproto.ConsensusParams{},
		},
	)
	require.NoError(t, err)
	deliverKey := []byte("deliver-key")
	baseapptestutil.RegisterCounterServer(suite.baseApp.MsgServiceRouter(), CounterServerImpl{t, capKey1, deliverKey})

	nBlocks := 3
	txPerHeight := 5

	for blockN := 0; blockN < nBlocks; blockN++ {
		txs := [][]byte{}

		var expectedChangeSet []*storetypes.StoreKVPair

		// create final block context state
		_, err := suite.baseApp.FinalizeBlock(&abci.RequestFinalizeBlock{Height: int64(blockN) + 1, Txs: txs})
		require.NoError(t, err)

		for i := 0; i < txPerHeight; i++ {
			counter := int64(blockN*txPerHeight + i)
			tx := newTxCounter(t, suite.txConfig, counter, counter)

			txBytes, err := suite.txConfig.TxEncoder()(tx)
			require.NoError(t, err)

			sKey := []byte(fmt.Sprintf("distKey%d", i))
			sVal := []byte(fmt.Sprintf("distVal%d", i))
			store := getFinalizeBlockStateCtx(suite.baseApp).KVStore(distKey1)
			store.Set(sKey, sVal)

			expectedChangeSet = append(expectedChangeSet, &storetypes.StoreKVPair{
				StoreKey: distKey1.Name(),
				Delete:   false,
				Key:      sKey,
				Value:    sVal,
			})

			txs = append(txs, txBytes)
		}

		res, err := suite.baseApp.FinalizeBlock(&abci.RequestFinalizeBlock{Height: int64(blockN) + 1, Txs: txs})
		require.NoError(t, err)
		for _, tx := range res.TxResults {
			events := tx.GetEvents()
			require.Len(t, events, 3, "should contain ante handler, message type and counter events respectively")
			// require.Equal(t, sdk.MarkEventsToIndex(counterEvent("ante_handler", counter).ToABCIEvents(), map[string]struct{}{})[0], events[0], "ante handler event")
			// require.Equal(t, sdk.MarkEventsToIndex(counterEvent(sdk.EventTypeMessage, counter).ToABCIEvents(), map[string]struct{}{})[0], events[2], "msg handler update counter event")
		}

		_, err = suite.baseApp.Commit()
		require.NoError(t, err)
		require.Equal(t, expectedChangeSet, mockListener1.ChangeSet, "should contain the same changeSet")
		require.Equal(t, expectedChangeSet, mockListener2.ChangeSet, "should contain the same changeSet")
	}
}

func Test_Ctx_with_StreamingManager(t *testing.T) {
	mockListener1 := NewMockABCIListener("lis_1")
	mockListener2 := NewMockABCIListener("lis_2")
	listeners := []storetypes.ABCIListener{&mockListener1, &mockListener2}
	streamingManager := storetypes.StreamingManager{ABCIListeners: listeners, StopNodeOnErr: true}
	streamingManagerOpt := func(bapp *baseapp.BaseApp) { bapp.SetStreamingManager(streamingManager) }
	addListenerOpt := func(bapp *baseapp.BaseApp) { bapp.CommitMultiStore().AddListeners([]storetypes.StoreKey{distKey1}) }
	suite := NewBaseAppSuite(t, streamingManagerOpt, addListenerOpt)

	_, err := suite.baseApp.InitChain(&abci.RequestInitChain{
		ConsensusParams: &tmproto.ConsensusParams{},
	})
	require.NoError(t, err)

	ctx := getFinalizeBlockStateCtx(suite.baseApp)
	sm := ctx.StreamingManager()
	require.NotNil(t, sm, fmt.Sprintf("nil StreamingManager: %v", sm))
	require.Equal(t, listeners, sm.ABCIListeners, fmt.Sprintf("should contain same listeners: %v", listeners))
	require.Equal(t, true, sm.StopNodeOnErr, "should contain StopNodeOnErr = true")

	nBlocks := 2

	for blockN := 0; blockN < nBlocks; blockN++ {

		_, err = suite.baseApp.FinalizeBlock(&abci.RequestFinalizeBlock{Height: int64(blockN) + 1})
		require.NoError(t, err)
		ctx := getFinalizeBlockStateCtx(suite.baseApp)
		sm := ctx.StreamingManager()
		require.NotNil(t, sm, fmt.Sprintf("nil StreamingManager: %v", sm))
		require.Equal(t, listeners, sm.ABCIListeners, fmt.Sprintf("should contain same listeners: %v", listeners))
		require.Equal(t, true, sm.StopNodeOnErr, "should contain StopNodeOnErr = true")

		_, err = suite.baseApp.Commit()
		require.NoError(t, err)
	}
}
