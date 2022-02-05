package service

import (
	"context"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/tendermint/tendermint/libs/log"
	"os"
	"os/signal"
	"sync"
	"syscall"
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
	testStreamingService         *KafkaStreamingService
	testListener1, testListener2 types.WriteListener
	testingCtx                   sdk.Context

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

	// Kafka stuff
	bootstrapServers = "localhost:9092"
	topicPrefix      = "block"
	flushTimeoutMs   = 15000
	topics           = []string{
		string(BeginBlockReqTopic),
		BeginBlockResTopic,
		DeliverTxReqTopic,
		DeliverTxResTopic,
		EndBlockReqTopic,
		EndBlockResTopic,
		StateChangeTopic,
	}

	producerConfig = kafka.ConfigMap{
		"bootstrap.servers": bootstrapServers,
		"client.id":         "testKafkaStreamService",
		"security.protocol": "PLAINTEXT",
		"enable.idempotence": "true",
		// Best practice for Kafka producer to prevent data loss
		"acks":              "all",
	}

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

// change this to write to in-memory io.Writer (e.g. bytes.Buffer)
func TestKafkaStreamingService(t *testing.T) {
	 testingCtx = sdk.NewContext(nil, types1.Header{}, false, log.TestingLogger())
	testKeys := []types.StoreKey{mockStoreKey1, mockStoreKey2}
	kss, err := NewKafkaStreamingService(producerConfig, topicPrefix, flushTimeoutMs, testKeys, testMarshaller, true)
	testStreamingService = kss
	require.Nil(t, err)
	require.IsType(t, &KafkaStreamingService{}, testStreamingService)
	require.Equal(t, topicPrefix, testStreamingService.topicPrefix)
	require.Equal(t, testMarshaller, testStreamingService.codec)
	deleteTopics(t, topics, bootstrapServers)
	createTopics(t, topics, bootstrapServers)
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
	expectedBeginBlockReqBytes, err := testMarshaller.MarshalJSON(&testBeginBlockReq)
	require.Nil(t, err)
	expectedBeginBlockResBytes, err := testMarshaller.MarshalJSON(&testBeginBlockRes)
	require.Nil(t, err)

	// write state changes
	testListener1.OnWrite(mockStoreKey1, mockKey1, mockValue1, false)
	testListener2.OnWrite(mockStoreKey2, mockKey2, mockValue2, false)
	testListener1.OnWrite(mockStoreKey1, mockKey3, mockValue3, false)

	// expected KV pairs
	expectedKVPair1, err := testMarshaller.MarshalJSON(&types.StoreKVPair{
		StoreKey: mockStoreKey1.Name(),
		Key:      mockKey1,
		Value:    mockValue1,
		Delete:   false,
	})
	require.Nil(t, err)
	expectedKVPair2, err := testMarshaller.MarshalJSON(&types.StoreKVPair{
		StoreKey: mockStoreKey2.Name(),
		Key:      mockKey2,
		Value:    mockValue2,
		Delete:   false,
	})
	require.Nil(t, err)
	expectedKVPair3, err := testMarshaller.MarshalJSON(&types.StoreKVPair{
		StoreKey: mockStoreKey1.Name(),
		Key:      mockKey3,
		Value:    mockValue3,
		Delete:   false,
	})
	require.Nil(t, err)

	// send the ABCI messages
	err = testStreamingService.ListenBeginBlock(testingCtx, testBeginBlockReq, testBeginBlockRes)
	require.Nil(t, err)

	// consume stored messages
	topics := []string{string(BeginBlockReqTopic), BeginBlockResTopic, StateChangeTopic}
	msgs, err := poll(bootstrapServers, topics, 5)
	require.Nil(t, err)

	// validate data stored in Kafka
	require.Equal(t, expectedBeginBlockReqBytes, getMessageValueForTopic(msgs, string(BeginBlockReqTopic), 0))
	require.Equal(t, expectedKVPair1, getMessageValueForTopic(msgs, StateChangeTopic, 0))
	require.Equal(t, expectedKVPair2, getMessageValueForTopic(msgs, StateChangeTopic, 1))
	require.Equal(t, expectedKVPair3, getMessageValueForTopic(msgs, StateChangeTopic, 2))
	require.Equal(t, expectedBeginBlockResBytes, getMessageValueForTopic(msgs, BeginBlockResTopic, 0))
}

func testListenDeliverTx1(t *testing.T) {
	expectedDeliverTxReq1Bytes, err := testMarshaller.MarshalJSON(&testDeliverTxReq1)
	require.Nil(t, err)
	expectedDeliverTxRes1Bytes, err := testMarshaller.MarshalJSON(&testDeliverTxRes1)
	require.Nil(t, err)

	// write state changes
	testListener1.OnWrite(mockStoreKey1, mockKey1, mockValue1, false)
	testListener2.OnWrite(mockStoreKey2, mockKey2, mockValue2, false)
	testListener1.OnWrite(mockStoreKey2, mockKey3, mockValue3, false)

	// expected KV pairs
	expectedKVPair1, err := testMarshaller.MarshalJSON(&types.StoreKVPair{
		StoreKey: mockStoreKey1.Name(),
		Key:      mockKey1,
		Value:    mockValue1,
		Delete:   false,
	})
	require.Nil(t, err)
	expectedKVPair2, err := testMarshaller.MarshalJSON(&types.StoreKVPair{
		StoreKey: mockStoreKey2.Name(),
		Key:      mockKey2,
		Value:    mockValue2,
		Delete:   false,
	})
	require.Nil(t, err)
	expectedKVPair3, err := testMarshaller.MarshalJSON(&types.StoreKVPair{
		StoreKey: mockStoreKey2.Name(),
		Key:      mockKey3,
		Value:    mockValue3,
		Delete:   false,
	})
	require.Nil(t, err)

	// send the ABCI messages
	err = testStreamingService.ListenDeliverTx(testingCtx, testDeliverTxReq1, testDeliverTxRes1)
	require.Nil(t, err)

	// consume stored messages
	topics := []string{DeliverTxReqTopic, DeliverTxResTopic, StateChangeTopic}
	msgs, err := poll(bootstrapServers, topics, 5)
	require.Nil(t, err)

	// validate data stored in Kafka
	require.Equal(t, expectedDeliverTxReq1Bytes, getMessageValueForTopic(msgs, DeliverTxReqTopic, 0))
	require.Equal(t, expectedKVPair1, getMessageValueForTopic(msgs, StateChangeTopic, 3))
	require.Equal(t, expectedKVPair2, getMessageValueForTopic(msgs, StateChangeTopic, 4))
	require.Equal(t, expectedKVPair3, getMessageValueForTopic(msgs, StateChangeTopic, 5))
	require.Equal(t, expectedDeliverTxRes1Bytes, getMessageValueForTopic(msgs, DeliverTxResTopic, 0))
}

func testListenDeliverTx2(t *testing.T) {
	expectedDeliverTxReq2Bytes, err := testMarshaller.MarshalJSON(&testDeliverTxReq2)
	require.Nil(t, err)
	expectedDeliverTxRes2Bytes, err := testMarshaller.MarshalJSON(&testDeliverTxRes2)
	require.Nil(t, err)

	// write state changes
	testListener1.OnWrite(mockStoreKey2, mockKey1, mockValue1, false)
	testListener2.OnWrite(mockStoreKey1, mockKey2, mockValue2, false)
	testListener1.OnWrite(mockStoreKey2, mockKey3, mockValue3, false)

	// expected KV pairs
	expectedKVPair1, err := testMarshaller.MarshalJSON(&types.StoreKVPair{
		StoreKey: mockStoreKey2.Name(),
		Key:      mockKey1,
		Value:    mockValue1,
		Delete:   false,
	})
	require.Nil(t, err)
	expectedKVPair2, err := testMarshaller.MarshalJSON(&types.StoreKVPair{
		StoreKey: mockStoreKey1.Name(),
		Key:      mockKey2,
		Value:    mockValue2,
		Delete:   false,
	})
	require.Nil(t, err)
	expectedKVPair3, err := testMarshaller.MarshalJSON(&types.StoreKVPair{
		StoreKey: mockStoreKey2.Name(),
		Key:      mockKey3,
		Value:    mockValue3,
		Delete:   false,
	})
	require.Nil(t, err)

	// send the ABCI messages
	err = testStreamingService.ListenDeliverTx(testingCtx, testDeliverTxReq2, testDeliverTxRes2)
	require.Nil(t, err)

	// consume stored messages
	topics := []string{DeliverTxReqTopic, DeliverTxResTopic, StateChangeTopic}
	msgs, err := poll(bootstrapServers, topics, 5)
	require.Nil(t, err)

	// validate data stored in Kafka
	require.Equal(t, expectedDeliverTxReq2Bytes, getMessageValueForTopic(msgs, DeliverTxReqTopic, 1))
	require.Equal(t, expectedKVPair1, getMessageValueForTopic(msgs, StateChangeTopic, 6))
	require.Equal(t, expectedKVPair2, getMessageValueForTopic(msgs, StateChangeTopic, 7))
	require.Equal(t, expectedKVPair3, getMessageValueForTopic(msgs, StateChangeTopic, 8))
	require.Equal(t, expectedDeliverTxRes2Bytes, getMessageValueForTopic(msgs, DeliverTxResTopic, 1))
}

func testListenEndBlock(t *testing.T) {
	expectedEndBlockReqBytes, err := testMarshaller.MarshalJSON(&testEndBlockReq)
	require.Nil(t, err)
	expectedEndBlockResBytes, err := testMarshaller.MarshalJSON(&testEndBlockRes)
	require.Nil(t, err)

	// write state changes
	testListener1.OnWrite(mockStoreKey1, mockKey1, mockValue1, false)
	testListener2.OnWrite(mockStoreKey1, mockKey2, mockValue2, false)
	testListener1.OnWrite(mockStoreKey2, mockKey3, mockValue3, false)

	// expected KV pairs
	expectedKVPair1, err := testMarshaller.MarshalJSON(&types.StoreKVPair{
		StoreKey: mockStoreKey1.Name(),
		Key:      mockKey1,
		Value:    mockValue1,
		Delete:   false,
	})
	require.Nil(t, err)
	expectedKVPair2, err := testMarshaller.MarshalJSON(&types.StoreKVPair{
		StoreKey: mockStoreKey1.Name(),
		Key:      mockKey2,
		Value:    mockValue2,
		Delete:   false,
	})
	require.Nil(t, err)
	expectedKVPair3, err := testMarshaller.MarshalJSON(&types.StoreKVPair{
		StoreKey: mockStoreKey2.Name(),
		Key:      mockKey3,
		Value:    mockValue3,
		Delete:   false,
	})
	require.Nil(t, err)

	// send the ABCI messages
	err = testStreamingService.ListenEndBlock(testingCtx, testEndBlockReq, testEndBlockRes)
	require.Nil(t, err)

	// consume stored messages
	topics := []string{EndBlockReqTopic, EndBlockResTopic, StateChangeTopic}
	msgs, err := poll(bootstrapServers, topics, 5)
	require.Nil(t, err)

	// validate data stored in Kafka
	require.Equal(t, expectedEndBlockReqBytes, getMessageValueForTopic(msgs, EndBlockReqTopic, 0))
	require.Equal(t, expectedKVPair1, getMessageValueForTopic(msgs, StateChangeTopic, 9))
	require.Equal(t, expectedKVPair2, getMessageValueForTopic(msgs, StateChangeTopic, 10))
	require.Equal(t, expectedKVPair3, getMessageValueForTopic(msgs, StateChangeTopic, 11))
	require.Equal(t, expectedEndBlockResBytes, getMessageValueForTopic(msgs, EndBlockResTopic, 0))
}

func getMessageValueForTopic(msgs []*kafka.Message, topic string, offset int64) []byte {
	topic = fmt.Sprintf("%s-%s", topicPrefix, topic)
	for _, m := range msgs {
		t := *m.TopicPartition.Topic
		o := int64(m.TopicPartition.Offset)
		if t == topic && o == offset {
			return m.Value
		}
	}
	return []byte{0}
}

func poll(bootstrapServers string, topics []string, expectedMsgCnt int) ([]*kafka.Message, error) {
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":      bootstrapServers,
		// Avoid connecting to IPv6 brokers:
		// This is needed for the ErrAllBrokersDown show-case below
		// when using localhost brokers on OSX, since the OSX resolver
		// will return the IPv6 addresses first.
		// You typically don't need to specify this configuration property.
		"broker.address.family": "v4",
		"group.id":              fmt.Sprintf("testGroup-%d", os.Process{}.Pid),
		"auto.offset.reset":     "earliest"})

	if err != nil {
		panic(fmt.Sprintf("Failed to create consumer: %s\n", err))
	}

	fmt.Printf("Created Consumer %v\n", c)

	var _topics []string
	for _, t := range topics {
		_topics = append(_topics, fmt.Sprintf("%s-%s", topicPrefix, t))
	}

	if err = c.SubscribeTopics(_topics, nil); err != nil {
		panic(fmt.Sprintf("Failed to subscribe to consumer: %s\n", err))
	}

	msgs := make([]*kafka.Message, 0)

	run := true

	for run {
		select {
		case sig := <-sigchan:
			fmt.Printf("Caught signal %v: terminating\n", sig)
			run = false
		default:
			ev := c.Poll(100)
			if ev == nil {
				continue
			}

			switch e := ev.(type) {
			case *kafka.Message:
				msgs = append(msgs, e)
			case kafka.Error:
				// Errors should generally be considered
				// informational, the client will try to
				// automatically recover.
				// But in this example we choose to terminate
				// the application if all brokers are down.
				fmt.Printf("%% Error: %v: %v\n", e.Code(), e)
				if e.Code() == kafka.ErrAllBrokersDown {
					run = false
				}
			default:
				fmt.Printf("Ignored %v\n", e)

				// Workaround so our tests pass.
				// Wait for the expected messages to be delivered before closing the consumer
				if expectedMsgCnt == len(msgs) {
					run = false
				}
			}
		}
	}

	fmt.Printf("Closing consumer\n")
	if err := c.Close(); err != nil {
		return nil, err
	}

	return msgs, nil
}

func createTopics(t *testing.T, topics []string, bootstrapServers string) {

	adminClient, err := kafka.NewAdminClient(&kafka.ConfigMap{
		"bootstrap.servers":       bootstrapServers,
		"broker.version.fallback": "0.10.0.0",
		"api.version.fallback.ms": 0,
	})
	if err != nil {
		fmt.Printf("Failed to create Admin client: %s\n", err)
		t.Fail()
	}

	// Contexts are used to abort or limit the amount of time
	// the Admin call blocks waiting for a result.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create topics on cluster.
	// Set Admin options to wait for the operation to finish (or at most 60s)
	maxDuration, err := time.ParseDuration("60s")
	if err != nil {
		fmt.Printf("time.ParseDuration(60s)")
		t.Fail()
	}

	var _topics []kafka.TopicSpecification
	for _, s := range topics {
		_topics = append(_topics,
			kafka.TopicSpecification{
				Topic:             fmt.Sprintf("%s-%s", topicPrefix, s),
				NumPartitions:     1,
				ReplicationFactor: 1})
	}
	results, err := adminClient.CreateTopics(ctx, _topics, kafka.SetAdminOperationTimeout(maxDuration))
	if err != nil {
		fmt.Printf("Problem during the topicPrefix creation: %v\n", err)
		t.Fail()
	}

	// Check for specific topicPrefix errors
	for _, result := range results {
		if result.Error.Code() != kafka.ErrNoError &&
			result.Error.Code() != kafka.ErrTopicAlreadyExists {
			fmt.Printf("Topic creation failed for %s: %v",
				result.Topic, result.Error.String())
			t.Fail()
		}
	}

	adminClient.Close()
}

func deleteTopics(t *testing.T, topics []string, bootstrapServers string) {
	// Create a new AdminClient.
	// AdminClient can also be instantiated using an existing
	// Producer or Consumer instance, see NewAdminClientFromProducer and
	// NewAdminClientFromConsumer.
	a, err := kafka.NewAdminClient(&kafka.ConfigMap{"bootstrap.servers": bootstrapServers})
	if err != nil {
		fmt.Printf("Failed to create Admin client: %s\n", err)
		t.Fail()
	}

	// Contexts are used to abort or limit the amount of time
	// the Admin call blocks waiting for a result.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Delete topics on cluster
	// Set Admin options to wait for the operation to finish (or at most 60s)
	maxDur, err := time.ParseDuration("60s")
	if err != nil {
		fmt.Printf("ParseDuration(60s)")
		t.Fail()
	}

	var _topics []string
	for _, s := range topics {
		_topics = append(_topics, fmt.Sprintf("%s-%s", topicPrefix, s))
	}

	results, err := a.DeleteTopics(ctx, _topics, kafka.SetAdminOperationTimeout(maxDur))
	if err != nil {
		fmt.Printf("Failed to delete topics: %v\n", err)
		t.Fail()
	}

	// Print results
	for _, result := range results {
		fmt.Printf("%s\n", result)
	}

	a.Close()
}
