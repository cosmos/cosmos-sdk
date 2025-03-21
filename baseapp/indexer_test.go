package baseapp_test

import (
	"cosmossdk.io/schema/appdata"
	storetypes "cosmossdk.io/store/types"
	"fmt"
	abci "github.com/cometbft/cometbft/abci/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	baseapptestutil "github.com/cosmos/cosmos-sdk/baseapp/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"testing"
)

type mockAppDataListener struct {
	appdata.Listener

	startBlockData []appdata.StartBlockData
	txData         []appdata.TxData
	eventData      []appdata.EventData
	kvPairData     []appdata.KVPairData
	commitData     []appdata.CommitData
}

func newMockAppDataListener() *mockAppDataListener {
	listener := &mockAppDataListener{}

	// Initialize the Listener with custom behavior to store data
	listener.Listener = appdata.Listener{
		StartBlock: func(data appdata.StartBlockData) error {
			listener.startBlockData = append(listener.startBlockData, data) // Store StartBlockData
			return nil
		},
		OnTx: func(data appdata.TxData) error {
			listener.txData = append(listener.txData, data) // Store TxData
			return nil
		},
		OnEvent: func(data appdata.EventData) error {
			listener.eventData = append(listener.eventData, data) // Store EventData
			return nil
		},
		OnKVPair: func(data appdata.KVPairData) error {
			listener.kvPairData = append(listener.kvPairData, data) // Store KVPairData
			return nil
		},
		Commit: func(data appdata.CommitData) (func() error, error) {
			listener.commitData = append(listener.commitData, data) // Store CommitData
			return nil, nil
		},
	}

	return listener
}

func TestAppDataListener(t *testing.T) {
	anteKey := []byte("ante-key")
	anteOpt := func(bapp *baseapp.BaseApp) { bapp.SetAnteHandler(anteHandlerTxTest(t, capKey1, anteKey)) }
	distOpt := func(bapp *baseapp.BaseApp) { bapp.MountStores(distKey1) }
	mockListener := newMockAppDataListener()
	streamingManager := storetypes.StreamingManager{ABCIListeners: []storetypes.ABCIListener{baseapp.NewListenerWrapper(mockListener.Listener, nil)}}
	streamingManagerOpt := func(bapp *baseapp.BaseApp) { bapp.SetStreamingManager(streamingManager) }
	addListenerOpt := func(bapp *baseapp.BaseApp) { bapp.CommitMultiStore().AddListeners([]storetypes.StoreKey{distKey1}) }

	// for event tests
	baseappOpts := func(app *baseapp.BaseApp) {
		app.SetPreBlocker(func(ctx sdk.Context, block *abci.RequestFinalizeBlock) (*sdk.ResponsePreBlock, error) {
			ctx.EventManager().EmitEvent(sdk.NewEvent("pre-block"))
			return nil, nil
		})
		app.SetBeginBlocker(func(_ sdk.Context) (sdk.BeginBlock, error) {
			return sdk.BeginBlock{
				Events: []abci.Event{
					{Type: "begin-block"},
				},
			}, nil
		})
		app.SetEndBlocker(func(_ sdk.Context) (sdk.EndBlock, error) {
			return sdk.EndBlock{
				Events: []abci.Event{
					{Type: "end-block"},
				},
			}, nil
		})
	}

	suite := NewBaseAppSuite(t, anteOpt, distOpt, streamingManagerOpt, addListenerOpt, baseappOpts)

	_, err := suite.baseApp.InitChain(
		&abci.RequestInitChain{
			ConsensusParams: &tmproto.ConsensusParams{},
		},
	)
	require.NoError(t, err)
	deliverKey := []byte("deliver-key")
	baseapptestutil.RegisterCounterServer(suite.baseApp.MsgServiceRouter(), CounterServerImpl{t, capKey1, deliverKey})

	txCount := 5
	txs := make([][]byte, txCount)

	for i := 0; i < txCount; i++ {
		tx := newTxCounter(t, suite.txConfig, int64(i), int64(i))

		txBytes, err := suite.txConfig.TxEncoder()(tx)
		require.NoError(t, err)

		sKey := []byte(fmt.Sprintf("distKey%d", i))
		sVal := []byte(fmt.Sprintf("distVal%d", i))
		store := getFinalizeBlockStateCtx(suite.baseApp).KVStore(distKey1)
		store.Set(sKey, sVal)

		txs[i] = txBytes
	}

	_, err = suite.baseApp.FinalizeBlock(&abci.RequestFinalizeBlock{Height: 1, Txs: txs})
	require.NoError(t, err)
	_, err = suite.baseApp.Commit()
	require.NoError(t, err)

	// StartBlockData
	require.Len(t, mockListener.startBlockData, 1)
	require.Equal(t, uint64(1), mockListener.startBlockData[0].Height)
	// TxData
	txData := mockListener.txData
	require.Len(t, txData, len(txs))
	for i := 0; i < txCount; i++ {
		require.Equal(t, int32(i), txData[i].TxIndex)
		txBytes, err := txData[i].Bytes()
		require.NoError(t, err)
		require.Equal(t, txs[i], txBytes)
	}
	// KVPairData
	require.Len(t, mockListener.kvPairData, 1)
	updates := mockListener.kvPairData[0].Updates
	for i := 0; i < txCount; i++ {
		require.Equal(t, []byte(distKey1.Name()), updates[i].Actor)
		require.Len(t, updates[i].StateChanges, 1)
		sKey := []byte(fmt.Sprintf("distKey%d", i))
		sVal := []byte(fmt.Sprintf("distVal%d", i))
		require.Equal(t, sKey, updates[i].StateChanges[0].Key)
		require.Equal(t, sVal, updates[i].StateChanges[0].Value)
	}
	// CommitData
	require.Len(t, mockListener.commitData, 1)
	// EventData
	require.Len(t, mockListener.eventData, 1)
	events := mockListener.eventData[0].Events
	require.Len(t, events, 3+txCount*3)

	for i := 0; i < 3; i++ {
		require.Equal(t, int32(0), events[i].TxIndex)
		require.Equal(t, int32(0), events[i].MsgIndex)
		require.Equal(t, int32(1), events[i].EventIndex)
		attrs, err := events[i].Attributes()
		require.NoError(t, err)
		require.Len(t, attrs, 2)
		switch i {
		case 0:
			require.Equal(t, appdata.PreBlockStage, events[i].BlockStage)
			require.Equal(t, "pre-block", events[i].Type)
		case 1:
			require.Equal(t, appdata.BeginBlockStage, events[i].BlockStage)
			require.Equal(t, "begin-block", events[i].Type)
		case 2:
			require.Equal(t, appdata.EndBlockStage, events[i].BlockStage)
			require.Equal(t, "end-block", events[i].Type)
		}
	}

	for i := 3; i < 3+txCount*3; i++ {
		require.Equal(t, appdata.TxProcessingStage, events[i].BlockStage)
		require.Equal(t, int32(i/3), events[i].TxIndex)
		switch i % 3 {
		case 0:
			require.Equal(t, "ante_handler", events[i].Type)
			require.Equal(t, int32(0), events[i].MsgIndex)
			require.Equal(t, int32(0), events[i].EventIndex)
			attrs, err := events[i].Attributes()
			require.NoError(t, err)
			require.Len(t, attrs, 2)
		case 1:
			require.Equal(t, "message", events[i].Type)
			require.Equal(t, int32(1), events[i].MsgIndex)
			require.Equal(t, int32(1), events[i].EventIndex)
			attrs, err := events[i].Attributes()
			require.NoError(t, err)
			require.Len(t, attrs, 5)
		case 2:
			require.Equal(t, "message", events[i].Type)
			require.Equal(t, int32(1), events[i].MsgIndex)
			require.Equal(t, int32(2), events[i].EventIndex)
			attrs, err := events[i].Attributes()
			require.NoError(t, err)
			require.Len(t, attrs, 4)
		}
	}
}
