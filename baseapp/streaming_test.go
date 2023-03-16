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
	sdk "github.com/cosmos/cosmos-sdk/types"
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

func (m MockABCIListener) ListenBeginBlock(ctx context.Context, req abci.RequestBeginBlock, res abci.ResponseBeginBlock) error {
	return nil
}

func (m MockABCIListener) ListenEndBlock(ctx context.Context, req abci.RequestEndBlock, res abci.ResponseEndBlock) error {
	return nil
}

func (m MockABCIListener) ListenDeliverTx(ctx context.Context, req abci.RequestDeliverTx, res abci.ResponseDeliverTx) error {
	return nil
}

func (m *MockABCIListener) ListenCommit(ctx context.Context, res abci.ResponseCommit, changeSet []*storetypes.StoreKVPair) error {
	m.ChangeSet = changeSet
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

	suite.baseApp.InitChain(abci.RequestInitChain{
		ConsensusParams: &tmproto.ConsensusParams{},
	})

	deliverKey := []byte("deliver-key")
	baseapptestutil.RegisterCounterServer(suite.baseApp.MsgServiceRouter(), CounterServerImpl{t, capKey1, deliverKey})

	nBlocks := 3
	txPerHeight := 5

	for blockN := 0; blockN < nBlocks; blockN++ {
		header := tmproto.Header{Height: int64(blockN) + 1}
		suite.baseApp.BeginBlock(abci.RequestBeginBlock{Header: header})
		var expectedChangeSet []*storetypes.StoreKVPair

		for i := 0; i < txPerHeight; i++ {
			counter := int64(blockN*txPerHeight + i)
			tx := newTxCounter(t, suite.txConfig, counter, counter)

			txBytes, err := suite.txConfig.TxEncoder()(tx)
			require.NoError(t, err)

			sKey := []byte(fmt.Sprintf("distKey%d", i))
			sVal := []byte(fmt.Sprintf("distVal%d", i))
			store := getDeliverStateCtx(suite.baseApp).KVStore(distKey1)
			store.Set(sKey, sVal)

			expectedChangeSet = append(expectedChangeSet, &storetypes.StoreKVPair{
				StoreKey: distKey1.Name(),
				Delete:   false,
				Key:      sKey,
				Value:    sVal,
			})

			res := suite.baseApp.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
			require.True(t, res.IsOK(), fmt.Sprintf("%v", res))

			events := res.GetEvents()
			require.Len(t, events, 3, "should contain ante handler, message type and counter events respectively")
			require.Equal(t, sdk.MarkEventsToIndex(counterEvent("ante_handler", counter).ToABCIEvents(), map[string]struct{}{})[0], events[0], "ante handler event")
			require.Equal(t, sdk.MarkEventsToIndex(counterEvent(sdk.EventTypeMessage, counter).ToABCIEvents(), map[string]struct{}{})[0], events[2], "msg handler update counter event")
		}

		suite.baseApp.EndBlock(abci.RequestEndBlock{})
		suite.baseApp.Commit()

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

	suite.baseApp.InitChain(abci.RequestInitChain{
		ConsensusParams: &tmproto.ConsensusParams{},
	})

	ctx := getDeliverStateCtx(suite.baseApp)
	sm := ctx.StreamingManager()
	require.NotNil(t, sm, fmt.Sprintf("nil StreamingManager: %v", sm))
	require.Equal(t, listeners, sm.ABCIListeners, fmt.Sprintf("should contain same listeners: %v", listeners))
	require.Equal(t, true, sm.StopNodeOnErr, "should contain StopNodeOnErr = true")

	nBlocks := 2

	for blockN := 0; blockN < nBlocks; blockN++ {
		header := tmproto.Header{Height: int64(blockN) + 1}
		suite.baseApp.BeginBlock(abci.RequestBeginBlock{Header: header})

		ctx := getDeliverStateCtx(suite.baseApp)
		sm := ctx.StreamingManager()
		require.NotNil(t, sm, fmt.Sprintf("nil StreamingManager: %v", sm))
		require.Equal(t, listeners, sm.ABCIListeners, fmt.Sprintf("should contain same listeners: %v", listeners))
		require.Equal(t, true, sm.StopNodeOnErr, "should contain StopNodeOnErr = true")

		suite.baseApp.EndBlock(abci.RequestEndBlock{})
		suite.baseApp.Commit()
	}
}
