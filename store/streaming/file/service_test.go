package file

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

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
	testStreamingService         *StreamingService
	testListener1, testListener2 types.WriteListener
	emptyContext                 = sdk.Context{}

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
		ConsensusParamUpdates: &abci.ConsensusParams{},
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

	// file stuff
	testPrefix = "testPrefix"
	testDir    = "./.test"

	// mock state changes
	mockKey1   = []byte{1, 2, 3}
	mockValue1 = []byte{3, 2, 1}
	mockKey2   = []byte{2, 3, 4}
	mockValue2 = []byte{4, 3, 2}
	mockKey3   = []byte{3, 4, 5}
	mockValue3 = []byte{5, 4, 3}
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

func TestFileStreamingService(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping TestFileStreamingService in CI environment")
	}
	err := os.Mkdir(testDir, 0o700)
	require.Nil(t, err)
	defer os.RemoveAll(testDir)

	testKeys := []types.StoreKey{mockStoreKey1, mockStoreKey2}
	testStreamingService, err = NewStreamingService(testDir, testPrefix, testKeys, testMarshaller)
	require.Nil(t, err)
	require.IsType(t, &StreamingService{}, testStreamingService)
	require.Equal(t, testPrefix, testStreamingService.filePrefix)
	require.Equal(t, testDir, testStreamingService.writeDir)
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

func testListenBeginBlock(t *testing.T) {
	expectedBeginBlockReqBytes, err := testMarshaller.Marshal(&testBeginBlockReq)
	require.Nil(t, err)
	expectedBeginBlockResBytes, err := testMarshaller.Marshal(&testBeginBlockRes)
	require.Nil(t, err)

	// write state changes
	testListener1.OnWrite(mockStoreKey1, mockKey1, mockValue1, false)
	testListener2.OnWrite(mockStoreKey2, mockKey2, mockValue2, false)
	testListener1.OnWrite(mockStoreKey1, mockKey3, mockValue3, false)

	// expected KV pairs
	expectedKVPair1, err := testMarshaller.Marshal(&types.StoreKVPair{
		StoreKey: mockStoreKey1.Name(),
		Key:      mockKey1,
		Value:    mockValue1,
		Delete:   false,
	})
	require.Nil(t, err)
	expectedKVPair2, err := testMarshaller.Marshal(&types.StoreKVPair{
		StoreKey: mockStoreKey2.Name(),
		Key:      mockKey2,
		Value:    mockValue2,
		Delete:   false,
	})
	require.Nil(t, err)
	expectedKVPair3, err := testMarshaller.Marshal(&types.StoreKVPair{
		StoreKey: mockStoreKey1.Name(),
		Key:      mockKey3,
		Value:    mockValue3,
		Delete:   false,
	})
	require.Nil(t, err)

	// send the ABCI messages
	err = testStreamingService.ListenBeginBlock(emptyContext, testBeginBlockReq, testBeginBlockRes)
	require.Nil(t, err)

	// load the file, checking that it was created with the expected name
	fileName := fmt.Sprintf("%s-block-%d-begin", testPrefix, testBeginBlockReq.GetHeader().Height)
	fileBytes, err := readInFile(fileName)
	require.Nil(t, err)

	// segment the file into the separate gRPC messages and check the correctness of each
	segments, err := segmentBytes(fileBytes)
	require.Nil(t, err)
	require.Equal(t, 5, len(segments))
	require.Equal(t, expectedBeginBlockReqBytes, segments[0])
	require.Equal(t, expectedKVPair1, segments[1])
	require.Equal(t, expectedKVPair2, segments[2])
	require.Equal(t, expectedKVPair3, segments[3])
	require.Equal(t, expectedBeginBlockResBytes, segments[4])
}

func testListenDeliverTx1(t *testing.T) {
	expectedDeliverTxReq1Bytes, err := testMarshaller.Marshal(&testDeliverTxReq1)
	require.Nil(t, err)
	expectedDeliverTxRes1Bytes, err := testMarshaller.Marshal(&testDeliverTxRes1)
	require.Nil(t, err)

	// write state changes
	testListener1.OnWrite(mockStoreKey1, mockKey1, mockValue1, false)
	testListener2.OnWrite(mockStoreKey2, mockKey2, mockValue2, false)
	testListener1.OnWrite(mockStoreKey2, mockKey3, mockValue3, false)

	// expected KV pairs
	expectedKVPair1, err := testMarshaller.Marshal(&types.StoreKVPair{
		StoreKey: mockStoreKey1.Name(),
		Key:      mockKey1,
		Value:    mockValue1,
		Delete:   false,
	})
	require.Nil(t, err)
	expectedKVPair2, err := testMarshaller.Marshal(&types.StoreKVPair{
		StoreKey: mockStoreKey2.Name(),
		Key:      mockKey2,
		Value:    mockValue2,
		Delete:   false,
	})
	require.Nil(t, err)
	expectedKVPair3, err := testMarshaller.Marshal(&types.StoreKVPair{
		StoreKey: mockStoreKey2.Name(),
		Key:      mockKey3,
		Value:    mockValue3,
		Delete:   false,
	})
	require.Nil(t, err)

	// send the ABCI messages
	err = testStreamingService.ListenDeliverTx(emptyContext, testDeliverTxReq1, testDeliverTxRes1)
	require.Nil(t, err)

	// load the file, checking that it was created with the expected name
	fileName := fmt.Sprintf("%s-block-%d-tx-%d", testPrefix, testBeginBlockReq.GetHeader().Height, 0)
	fileBytes, err := readInFile(fileName)
	require.Nil(t, err)

	// segment the file into the separate gRPC messages and check the correctness of each
	segments, err := segmentBytes(fileBytes)
	require.Nil(t, err)
	require.Equal(t, 5, len(segments))
	require.Equal(t, expectedDeliverTxReq1Bytes, segments[0])
	require.Equal(t, expectedKVPair1, segments[1])
	require.Equal(t, expectedKVPair2, segments[2])
	require.Equal(t, expectedKVPair3, segments[3])
	require.Equal(t, expectedDeliverTxRes1Bytes, segments[4])
}

func testListenDeliverTx2(t *testing.T) {
	expectedDeliverTxReq2Bytes, err := testMarshaller.Marshal(&testDeliverTxReq2)
	require.Nil(t, err)
	expectedDeliverTxRes2Bytes, err := testMarshaller.Marshal(&testDeliverTxRes2)
	require.Nil(t, err)

	// write state changes
	testListener1.OnWrite(mockStoreKey2, mockKey1, mockValue1, false)
	testListener2.OnWrite(mockStoreKey1, mockKey2, mockValue2, false)
	testListener1.OnWrite(mockStoreKey2, mockKey3, mockValue3, false)

	// expected KV pairs
	expectedKVPair1, err := testMarshaller.Marshal(&types.StoreKVPair{
		StoreKey: mockStoreKey2.Name(),
		Key:      mockKey1,
		Value:    mockValue1,
		Delete:   false,
	})
	require.Nil(t, err)
	expectedKVPair2, err := testMarshaller.Marshal(&types.StoreKVPair{
		StoreKey: mockStoreKey1.Name(),
		Key:      mockKey2,
		Value:    mockValue2,
		Delete:   false,
	})
	require.Nil(t, err)
	expectedKVPair3, err := testMarshaller.Marshal(&types.StoreKVPair{
		StoreKey: mockStoreKey2.Name(),
		Key:      mockKey3,
		Value:    mockValue3,
		Delete:   false,
	})
	require.Nil(t, err)

	// send the ABCI messages
	err = testStreamingService.ListenDeliverTx(emptyContext, testDeliverTxReq2, testDeliverTxRes2)
	require.Nil(t, err)

	// load the file, checking that it was created with the expected name
	fileName := fmt.Sprintf("%s-block-%d-tx-%d", testPrefix, testBeginBlockReq.GetHeader().Height, 1)
	fileBytes, err := readInFile(fileName)
	require.Nil(t, err)

	// segment the file into the separate gRPC messages and check the correctness of each
	segments, err := segmentBytes(fileBytes)
	require.Nil(t, err)
	require.Equal(t, 5, len(segments))
	require.Equal(t, expectedDeliverTxReq2Bytes, segments[0])
	require.Equal(t, expectedKVPair1, segments[1])
	require.Equal(t, expectedKVPair2, segments[2])
	require.Equal(t, expectedKVPair3, segments[3])
	require.Equal(t, expectedDeliverTxRes2Bytes, segments[4])
}

func testListenEndBlock(t *testing.T) {
	expectedEndBlockReqBytes, err := testMarshaller.Marshal(&testEndBlockReq)
	require.Nil(t, err)
	expectedEndBlockResBytes, err := testMarshaller.Marshal(&testEndBlockRes)
	require.Nil(t, err)

	// write state changes
	testListener1.OnWrite(mockStoreKey1, mockKey1, mockValue1, false)
	testListener2.OnWrite(mockStoreKey1, mockKey2, mockValue2, false)
	testListener1.OnWrite(mockStoreKey2, mockKey3, mockValue3, false)

	// expected KV pairs
	expectedKVPair1, err := testMarshaller.Marshal(&types.StoreKVPair{
		StoreKey: mockStoreKey1.Name(),
		Key:      mockKey1,
		Value:    mockValue1,
		Delete:   false,
	})
	require.Nil(t, err)
	expectedKVPair2, err := testMarshaller.Marshal(&types.StoreKVPair{
		StoreKey: mockStoreKey1.Name(),
		Key:      mockKey2,
		Value:    mockValue2,
		Delete:   false,
	})
	require.Nil(t, err)
	expectedKVPair3, err := testMarshaller.Marshal(&types.StoreKVPair{
		StoreKey: mockStoreKey2.Name(),
		Key:      mockKey3,
		Value:    mockValue3,
		Delete:   false,
	})
	require.Nil(t, err)

	// send the ABCI messages
	err = testStreamingService.ListenEndBlock(emptyContext, testEndBlockReq, testEndBlockRes)
	require.Nil(t, err)

	// load the file, checking that it was created with the expected name
	fileName := fmt.Sprintf("%s-block-%d-end", testPrefix, testEndBlockReq.Height)
	fileBytes, err := readInFile(fileName)
	require.Nil(t, err)

	// segment the file into the separate gRPC messages and check the correctness of each
	segments, err := segmentBytes(fileBytes)
	require.Nil(t, err)
	require.Equal(t, 5, len(segments))
	require.Equal(t, expectedEndBlockReqBytes, segments[0])
	require.Equal(t, expectedKVPair1, segments[1])
	require.Equal(t, expectedKVPair2, segments[2])
	require.Equal(t, expectedKVPair3, segments[3])
	require.Equal(t, expectedEndBlockResBytes, segments[4])
}

func readInFile(name string) ([]byte, error) {
	path := filepath.Join(testDir, name)
	return os.ReadFile(path)
}

// Returns all of the protobuf messages contained in the byte array as an array of byte arrays
// The messages have their length prefix removed
func segmentBytes(bz []byte) ([][]byte, error) {
	var err error
	segments := make([][]byte, 0)
	for len(bz) > 0 {
		var segment []byte
		segment, bz, err = getHeadSegment(bz)
		if err != nil {
			return nil, err
		}
		segments = append(segments, segment)
	}
	return segments, nil
}

// Returns the bytes for the leading protobuf object in the byte array (removing the length prefix) and returns the remainder of the byte array
func getHeadSegment(bz []byte) ([]byte, []byte, error) {
	size, prefixSize := binary.Uvarint(bz)
	if prefixSize < 0 {
		return nil, nil, fmt.Errorf("invalid number of bytes read from length-prefixed encoding: %d", prefixSize)
	}
	if size > uint64(len(bz)-prefixSize) {
		return nil, nil, fmt.Errorf("not enough bytes to read; want: %v, got: %v", size, len(bz)-prefixSize)
	}
	return bz[prefixSize:(uint64(prefixSize) + size)], bz[uint64(prefixSize)+size:], nil
}
