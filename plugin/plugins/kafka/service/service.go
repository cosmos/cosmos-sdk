package service

import (
	"errors"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
	abci "github.com/tendermint/tendermint/abci/types"
	"sync"
)

/*
This service writes all messages to a single topicPrefix with only one partition to maintain the order of messages.

The naming schema and data format for the messages this service writes out to Kafka is as such:

After every `BeginBlockEvent` request a new message key prefix is created with the name `block-{N}-begin`, where N is the block number.
All subsequent state changes are written out to this topicPrefix until the first `DeliverTxEvent` request is received. At the head of these files,
the length-prefixed protobuf encoded `BeginBlockEvent` request is written, and the response is written at the tail.

After every `DeliverTxEvent` request a new file is created with the name `block-{N}-tx-{M}` where N is the block number and M
is the tx number in the block (i.e. 0, 1, 2...). All subsequent state changes are written out to this file until the next
`DeliverTxEvent` request is received or an `EndBlockEvent` request is received. At the head of these files, the length-prefixed protobuf
encoded `DeliverTxEvent` request is written, and the response is written at the tail.

After every `EndBlockEvent` request a new file is created with the name `block-{N}-end`, where N is the block number. All
subsequent state changes are written out to this file until the next `BeginBlockEvent` request is received. At the head of these files,
the length-prefixed protobuf encoded `EndBlockEvent` request is written, and the response is written at the tail.
*/

// Event Kafka message key enum types for listen events.
type Event string
const (
	BeginBlockEvent Event = "BEGIN_BLOCK"
	EndBlockEvent         = "END_BLOCK"
	DeliverTxEvent        = "DELIVER_TX"
)

// EventType Kafka message key enum types for the event types.
type EventType string
const (
	RequestEventType     EventType = "REQUEST"
	ResponseEventType              = "RESPONSE"
	StateChangeEventType           = "STATE_CHANGE"
)

// EventTypeValueTypeTopic Kafka topic name enum types
type EventTypeValueTypeTopic string
const (
	BeginBlockReqTopic EventTypeValueTypeTopic = "begin-block-req"
	BeginBlockResTopic                         = "begin-block-res"
	EndBlockReqTopic                           = "end-block-req"
	EndBlockResTopic                           = "end-block-res"
	DeliverTxReqTopic                          = "deliver-tx-req"
	DeliverTxResTopic                          = "deliver-tx-res"
	StateChangeTopic                           = "state-change"
)

// MsgKeyFtm Kafka message composite key format enum types
const (
	MsgKeyFtm = `{"block_height":%d,"event":"%s","event_id":%d,"event_type":"%s","event_type_id":%d}`
)

var _ baseapp.StreamingService = (*KafkaStreamingService)(nil)

// KafkaStreamingService is a concrete implementation of streaming.Service that writes state changes out to Kafka
type KafkaStreamingService struct {
	listeners              map[types.StoreKey][]types.WriteListener // the listeners that will be initialized with BaseApp
	srcChan                <-chan []byte                            // the channel that all of the WriteListeners write their data out to
	topicPrefix            string                                   // topicPrefix prefix name
	producer               *kafka.Producer                          // the producer instance that will be used to send messages to Kafka
	flushTimeoutMs         int                                      // the time to wait for outstanding messages and requests to complete delivery (milliseconds)
	codec                  codec.BinaryCodec                        // binary marshaller used for re-marshalling the ABCI messages to write them out to the destination files
	stateCache             [][]byte                                 // cache the protobuf binary encoded StoreKVPairs in the order they are received
	stateCacheLock         *sync.Mutex                              // mutex for the state cache
	currentBlockNumber     int64                                    // the current block number
	currentTxIndex         int64                                    // the index of the current tx
	quitChan               chan struct{}                            // channel used for synchronize closure
	deliveryChan           chan kafka.Event                         // Kafka producer delivery report channel
	haltAppOnDeliveryError bool                                     // true if the app should be halted on streaming errors, false otherwise
}

// IntermediateWriter is used so that we do not need to update the underlying io.Writer inside the StoreKVPairWriteListener
// everytime we begin writing to Kafka topic(s)
type IntermediateWriter struct {
	outChan chan<- []byte
}

// NewIntermediateWriter create an instance of an intermediateWriter that sends to the provided channel
func NewIntermediateWriter(outChan chan<- []byte) *IntermediateWriter {
	return &IntermediateWriter{
		outChan: outChan,
	}
}

// Write satisfies io.Writer
func (iw *IntermediateWriter) Write(b []byte) (int, error) {
	iw.outChan <- b
	return len(b), nil
}

// NewKafkaStreamingService creates a new KafkaStreamingService
func NewKafkaStreamingService(
	producerConfig         kafka.ConfigMap,
	topicPrefix            string,
	flushTimeoutMs         int,
	storeKeys              []types.StoreKey,
	c                      codec.BinaryCodec,
	haltAppOnDeliveryError bool,
) (*KafkaStreamingService, error) {
	listenChan := make(chan []byte)
	iw := NewIntermediateWriter(listenChan)
	listener := types.NewStoreKVPairWriteListener(iw, c)
	listeners := make(map[types.StoreKey][]types.WriteListener, len(storeKeys))
	// in this case, we are using the same listener for each Store
	for _, key := range storeKeys {
		listeners[key] = append(listeners[key], listener)
	}
	// Initialize the producer and connect to Kafka cluster
	p, err := kafka.NewProducer(&producerConfig)
	if err != nil {
		return nil, err
	}

	kss := &KafkaStreamingService{
		listeners:              listeners,
		srcChan:                listenChan,
		topicPrefix:            topicPrefix,
		producer:               p,
		flushTimeoutMs:         flushTimeoutMs,
		codec:                  c,
		stateCache:             make([][]byte, 0),
		stateCacheLock:         new(sync.Mutex),
		deliveryChan:           make(chan kafka.Event),
		haltAppOnDeliveryError: haltAppOnDeliveryError,
	}

	return kss, nil
}

// Listeners returns the KafkaStreamingService's underlying WriteListeners, use for registering them with the BaseApp
func (kss *KafkaStreamingService) Listeners() map[types.StoreKey][]types.WriteListener {
	return kss.listeners
}

// ListenBeginBlock satisfies the Hook interface
// It writes out the received BeginBlockEvent request and response and the resulting state changes out to a Kafka topicPrefix
// as described in the above the naming schema
func (kss *KafkaStreamingService) ListenBeginBlock(
	ctx sdk.Context,
	req abci.RequestBeginBlock,
	res abci.ResponseBeginBlock,
) error {
	kss.setBeginBlock(req)
	eventId := int64(1)
	eventTypeId := 1

	// write req
	key := fmt.Sprintf(MsgKeyFtm, kss.currentBlockNumber, BeginBlockEvent, eventId, RequestEventType, eventTypeId)
	if err := kss.writeAsJsonToKafka(ctx, string(BeginBlockReqTopic), key, &req); err != nil {
		return err
	}

	// write state changes
	if err := kss.writeStateChange(ctx, string(BeginBlockEvent), eventId); err != nil {
		return err
	}

	// write res
	key = fmt.Sprintf(MsgKeyFtm, kss.currentBlockNumber, EndBlockEvent, 1, ResponseEventType, 1)
	if err := kss.writeAsJsonToKafka(ctx, BeginBlockResTopic, key, &res); err != nil {
		return err
	}

	return nil
}

func (kss *KafkaStreamingService) setBeginBlock(req abci.RequestBeginBlock) {
	kss.currentBlockNumber = req.GetHeader().Height
	kss.currentTxIndex = 0
}

// ListenDeliverTx satisfies the Hook interface
// It writes out the received DeliverTxEvent request and response and the resulting state changes out to a file as described
// in the above the naming schema
func (kss *KafkaStreamingService) ListenDeliverTx(
	ctx sdk.Context,
	req abci.RequestDeliverTx,
	res abci.ResponseDeliverTx,
) error {
	eventId := kss.getDeliverTxId()
	eventTypeId := 1

	// write req
	key := fmt.Sprintf(MsgKeyFtm, kss.currentBlockNumber, DeliverTxEvent, eventId, RequestEventType, eventTypeId)
	if err := kss.writeAsJsonToKafka(ctx, DeliverTxReqTopic, key, &req); err != nil {
		return err
	}

	// write state changes
	if err := kss.writeStateChange(ctx, DeliverTxEvent, eventId); err != nil {
		return err
	}

	// write res
	key = fmt.Sprintf(MsgKeyFtm, kss.currentBlockNumber, DeliverTxEvent, eventId, ResponseEventType, 1)
	if err := kss.writeAsJsonToKafka(ctx, DeliverTxResTopic, key, &res); err != nil {
		return err
	}

	return nil
}

func (kss *KafkaStreamingService) getDeliverTxId() int64 {
	kss.currentTxIndex++
	return kss.currentTxIndex
}

// ListenEndBlock satisfies the Hook interface
// It writes out the received EndBlockEvent request and response and the resulting state changes out to a file as described
// in the above the naming schema
func (kss *KafkaStreamingService) ListenEndBlock(
	ctx sdk.Context,
	req abci.RequestEndBlock,
	res abci.ResponseEndBlock,
) error {
	eventId := int64(1)
	eventTypeId := 1

	// write req
	key := fmt.Sprintf(MsgKeyFtm, kss.currentBlockNumber, EndBlockEvent, eventId, RequestEventType, eventTypeId)
	if err := kss.writeAsJsonToKafka(ctx, EndBlockReqTopic, key, &req); err != nil {
		return err
	}

	// write state changes
	if err := kss.writeStateChange(ctx, EndBlockEvent, eventId); err != nil {
		return err
	}

	// write res
	key = fmt.Sprintf(MsgKeyFtm, kss.currentBlockNumber, EndBlockEvent, eventId, ResponseEventType, eventTypeId)
	if err := kss.writeAsJsonToKafka(ctx, EndBlockResTopic, key, &res); err != nil {
		return err
	}

	return nil
}

// HaltAppOnDeliveryError whether or not to halt the application when delivery of massages fails
// in ListenBeginBlock, ListenEndBlock, ListenDeliverTx. Setting this to `false` will give fire-and-forget semantics.
func (kss *KafkaStreamingService) HaltAppOnDeliveryError() bool {
	return kss.haltAppOnDeliveryError
}

// Stream spins up a goroutine select loop which awaits length-prefixed binary encoded KV pairs and caches them in the order they were received
// Do we need this and an intermediate writer? We could just write directly to the buffer on calls to Write
// But then we don't support a Stream interface, which could be needed for other types of streamers
func (kss *KafkaStreamingService) Stream(wg *sync.WaitGroup) error {
	if kss.quitChan != nil {
		return errors.New("`Stream` has already been called. The stream needs to be closed before it can be started again")
	}
	kss.quitChan = make(chan struct{})
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-kss.quitChan:
				return
			case by := <-kss.srcChan:
				kss.stateCacheLock.Lock()
				kss.stateCache = append(kss.stateCache, by)
				kss.stateCacheLock.Unlock()
			}
		}
	}()
	return nil
}

// Close satisfies the io.Closer interface
func (kss *KafkaStreamingService) Close() error {
	kss.producer.Flush(kss.flushTimeoutMs)
	close(kss.quitChan)
	close(kss.deliveryChan)
	kss.producer.Close()
	return nil
}

func (kss *KafkaStreamingService) writeStateChange(ctx sdk.Context, event string, eventId int64) error {
	// write all state changes cached for this stage to Kafka
	kss.stateCacheLock.Lock()
	kodec := kss.codec.(*codec.ProtoCodec)
	kvPair := new(types.StoreKVPair)
	for i, stateChange := range kss.stateCache {
		key := fmt.Sprintf(MsgKeyFtm, kss.currentBlockNumber, event, eventId, StateChangeEventType, i+1)
		if err := kodec.UnmarshalLengthPrefixed(stateChange, kvPair); err != nil {
			return err
		}
		if err := kss.writeAsJsonToKafka(ctx, StateChangeTopic, key, kvPair); err != nil {
			return err
		}
	}

	// reset cache
	kss.stateCache = nil
	kss.stateCacheLock.Unlock()

	return nil
}

func (kss *KafkaStreamingService) writeAsJsonToKafka(
	ctx sdk.Context,
	topic string,
	key string,
	data proto.Message,
) error {
	kodec := kss.codec.(*codec.ProtoCodec)
	json, err := kodec.MarshalJSON(data)
	if err != nil {
		return err
	}
	if len(kss.topicPrefix) > 0 {
		topic = fmt.Sprintf("%s-%s", kss.topicPrefix, topic)
	}

	// produce message
	if err := kss.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          json,
		Key:            []byte(key),
	}, kss.deliveryChan); err != nil {
		return err
	}

	return kss.checkDeliveryReport(ctx)
}

// checkDeliveryReport checks kafka.Producer delivery report for successful or failed messages
func (kss *KafkaStreamingService) checkDeliveryReport(ctx sdk.Context) error {
	e := <-kss.deliveryChan
	m := e.(*kafka.Message)
	topic := *m.TopicPartition.Topic
	partition := m.TopicPartition.Partition
	offset := m.TopicPartition.Offset
	key := string(m.Key)
	topicErr := m.TopicPartition.Error
	logger := ctx.Logger()

	if topicErr != nil {
		logger.Error("Delivery failed: ", "topic", topic, "partition", partition, "key", key, "err", topicErr)
		return topicErr
	} else {
		logger.Debug("Delivered message:", "topic", topic, "partition", partition, "offset", offset, "key", key)
	}

	return nil
}