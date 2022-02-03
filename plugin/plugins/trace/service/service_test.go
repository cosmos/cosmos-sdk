package service

import (
	"github.com/tendermint/tendermint/libs/log"
	"sync"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	codecTypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	types1 "github.com/tendermint/tendermint/proto/tendermint/types"
)

var (
	interfaceRegistry            = codecTypes.NewInterfaceRegistry()
	testMarshaller               = codec.NewProtoCodec(interfaceRegistry)
	testStreamingService         *TraceStreamingService
	testListener1, testListener2 types.WriteListener
	emptyContext                 = sdk.Context{}
	loggerContext                sdk.Context

	// test abci message types
	mockHash          = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9}
	testBeginBlockReq = abci.RequestBeginBlock{
		Header: types1.Header{
			Height: 1,
		},
		ByzantineValidators: []abci.Evidence{},
		Hash:                mockHash,
		LastCommitInfo: abci.LastCommitInfo{
			Round: 1,
			Votes: []abci.VoteInfo{},
		},
	}
	testBeginBlockRes = abci.ResponseBeginBlock{
		Events: []abci.Event{
			{
				Type: "testEventType1",
			},
			{
				Type: "testEventType2",
			},
		},
	}
	testEndBlockReq = abci.RequestEndBlock{
		Height: 1,
	}
	testEndBlockRes = abci.ResponseEndBlock{
		Events:                []abci.Event{},
		ConsensusParamUpdates: &types1.ConsensusParams{},
		ValidatorUpdates:      []abci.ValidatorUpdate{},
	}
	mockTxBytes1      = []byte{9, 8, 7, 6, 5, 4, 3, 2, 1}
	testDeliverTxReq1 = abci.RequestDeliverTx{
		Tx: mockTxBytes1,
	}
	mockTxBytes2      = []byte{8, 7, 6, 5, 4, 3, 2}
	testDeliverTxReq2 = abci.RequestDeliverTx{
		Tx: mockTxBytes2,
	}
	mockTxResponseData1 = []byte{1, 3, 5, 7, 9}
	testDeliverTxRes1   = abci.ResponseDeliverTx{
		Events:    []abci.Event{},
		Code:      1,
		Codespace: "mockCodeSpace",
		Data:      mockTxResponseData1,
		GasUsed:   2,
		GasWanted: 3,
		Info:      "mockInfo",
		Log:       "mockLog",
	}
	mockTxResponseData2 = []byte{1, 3, 5, 7, 9}
	testDeliverTxRes2   = abci.ResponseDeliverTx{
		Events:    []abci.Event{},
		Code:      1,
		Codespace: "mockCodeSpace",
		Data:      mockTxResponseData2,
		GasUsed:   2,
		GasWanted: 3,
		Info:      "mockInfo",
		Log:       "mockLog",
	}

	// mock store keys
	mockStoreKey1 = sdk.NewKVStoreKey("mockStore1")
	mockStoreKey2 = sdk.NewKVStoreKey("mockStore2")

	// mock state changes
	mockKey1   = []byte{1, 2, 3}
	mockValue1 = []byte{3, 2, 1}
	mockKey2   = []byte{2, 3, 4}
	mockValue2 = []byte{4, 3, 2}
	mockKey3   = []byte{3, 4, 5}
	mockValue3 = []byte{5, 4, 3}

	// maximum amount of time ListenSuccess() will wait receipt
	// that all current block messages were delivered to the service.
	deliverBlockWaitLimit = time.Duration(1000)

	// print event data in stdout
	printDataToStdout = true
)

func TestIntermediateWriter(t *testing.T) {
	outChan := make(chan []byte, 0)
	iw := NewIntermediateWriter(outChan)
	require.IsType(t, &IntermediateWriter{}, iw)
	testBytes := []byte{1, 2, 3, 4, 5}
	var length int
	var err error
	waitChan := make(chan struct{}, 0)
	go func() {
		length, err = iw.Write(testBytes)
		waitChan <- struct{}{}
	}()
	receivedBytes := <-outChan
	<-waitChan
	require.Equal(t, len(testBytes), length)
	require.Equal(t, testBytes, receivedBytes)
	require.Nil(t, err)
}

func TestKafkaStreamingService(t *testing.T) {
	loggerContext = emptyContext.WithLogger(log.TestingLogger())
	testKeys := []types.StoreKey{mockStoreKey1, mockStoreKey2}
	tss, err := NewTraceStreamingService(testKeys, testMarshaller, deliverBlockWaitLimit, printDataToStdout)
	testStreamingService = tss
	require.Nil(t, err)
	require.IsType(t, &TraceStreamingService{}, testStreamingService)
	require.Equal(t, testMarshaller, testStreamingService.codec)
	testListener1 = testStreamingService.listeners[mockStoreKey1][0]
	testListener2 = testStreamingService.listeners[mockStoreKey2][0]
	wg := new(sync.WaitGroup)
	testStreamingService.Stream(wg)
	testListenBeginBlock(t)
	testListenDeliverTx1(t)
	testListenDeliverTx2(t)
	testListenEndBlock(t)
	testStreamingService.Close()
	wg.Wait()
}

func  testListenBeginBlock(t *testing.T) {
	// write state changes
	testListener1.OnWrite(mockStoreKey1, mockKey1, mockValue1, false)
	testListener2.OnWrite(mockStoreKey2, mockKey2, mockValue2, false)
	testListener1.OnWrite(mockStoreKey1, mockKey3, mockValue3, false)

	// send the ABCI messages
	err := testStreamingService.ListenBeginBlock(loggerContext, testBeginBlockReq, testBeginBlockRes)
	require.Nil(t, err)
}

func testListenDeliverTx1(t *testing.T) {
	// write state changes
	testListener1.OnWrite(mockStoreKey1, mockKey1, mockValue1, false)
	testListener2.OnWrite(mockStoreKey2, mockKey2, mockValue2, false)
	testListener1.OnWrite(mockStoreKey2, mockKey3, mockValue3, false)

	// send the ABCI messages
	err := testStreamingService.ListenDeliverTx(loggerContext, testDeliverTxReq1, testDeliverTxRes1)
	require.Nil(t, err)
}

func testListenDeliverTx2(t *testing.T) {
	// write state changes
	testListener1.OnWrite(mockStoreKey2, mockKey1, mockValue1, false)
	testListener2.OnWrite(mockStoreKey1, mockKey2, mockValue2, false)
	testListener1.OnWrite(mockStoreKey2, mockKey3, mockValue3, false)

	// send the ABCI messages
	err := testStreamingService.ListenDeliverTx(loggerContext, testDeliverTxReq2, testDeliverTxRes2)
	require.Nil(t, err)
}

func testListenEndBlock(t *testing.T) {
	// write state changes
	testListener1.OnWrite(mockStoreKey1, mockKey1, mockValue1, false)
	testListener2.OnWrite(mockStoreKey1, mockKey2, mockValue2, false)
	testListener1.OnWrite(mockStoreKey2, mockKey3, mockValue3, false)

	// send the ABCI messages
	err := testStreamingService.ListenEndBlock(loggerContext, testEndBlockReq, testEndBlockRes)
	require.Nil(t, err)
}
