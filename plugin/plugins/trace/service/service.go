package service

import (
	"errors"
	"fmt"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/gogo/protobuf/proto"
	"sync"
	"time"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ baseapp.StreamingService = (*TraceStreamingService)(nil)

// Event message key enum types for listen events.
type Event string
const (
	BeginBlockEvent Event = "BEGIN_BLOCK"
	EndBlockEvent         = "END_BLOCK"
	DeliverTxEvent        = "DELIVER_TX"
)

// EventType message key enum types for the event types.
type EventType string
const (
	RequestEventType     EventType = "REQUEST"
	ResponseEventType              = "RESPONSE"
	StateChangeEventType           = "STATE_CHANGE"
)

// LogMsgFmt message output format
const (
	LogMsgFmt = `block_height:%d => event:%s => event_id:%d => event_type:%s => event_type_id:%d`
)

// TraceStreamingService is a concrete implementation of streaming.Service that writes state changes to log file.
type TraceStreamingService struct {
	listeners             map[types.StoreKey][]types.WriteListener // the listeners that will be initialized with BaseApp
	srcChan               <-chan []byte                            // the channel that all of the WriteListeners write their data out to
	codec                 codec.BinaryCodec                        // binary marshaller used for re-marshalling the ABCI messages to write them out to the destination files
	stateCache            [][]byte                                 // cache the protobuf binary encoded StoreKVPairs in the order they are received
	stateCacheLock        *sync.Mutex                              // mutex for the state cache
	currentBlockNumber    int64                                    // the current block number
	currentTxIndex        int64                                    // the index of the current tx
	quitChan              chan struct{}                            // channel used for synchronize closure
	successChan           chan bool                                // channel used for signaling success or failure of message delivery to external service
	deliveredMessages     bool                                     // True if messages were delivered, false otherwise.
	deliveredBlockChan    chan struct{}                            // channel used for signaling the delivery of all messages for the current block.
	deliverBlockWaitLimit time.Duration                            // the time to wait for service to deliver current block messages before timing out.
	printDataToStdout     bool                                     // Print types.StoreKVPair data stored in each event to stdout.
}

// IntermediateWriter is used so that we do not need to update the underlying io.Writer inside the StoreKVPairWriteListener
// everytime we begin writing
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

// NewTraceStreamingService creates a new TraceStreamingService for the provided
// storeKeys, BinaryCodec and deliverBlockWaitLimit (in milliseconds)
func NewTraceStreamingService(
	storeKeys             []types.StoreKey,
	c                     codec.BinaryCodec,
	deliverBlockWaitLimit time.Duration,
	printDataToStdout     bool,
) (*TraceStreamingService, error) {
	successChan := make(chan bool, 1)
	listenChan := make(chan []byte)
	iw := NewIntermediateWriter(listenChan)
	listener := types.NewStoreKVPairWriteListener(iw, c)
	listeners := make(map[types.StoreKey][]types.WriteListener, len(storeKeys))
	// in this case, we are using the same listener for each Store
	for _, key := range storeKeys {
		listeners[key] = append(listeners[key], listener)
	}

	tss := &TraceStreamingService{
		listeners:             listeners,
		srcChan:               listenChan,
		codec:                 c,
		stateCache:            make([][]byte, 0),
		stateCacheLock:        new(sync.Mutex),
		successChan:           successChan,
		deliveredMessages:     true,
		deliverBlockWaitLimit: deliverBlockWaitLimit,
		printDataToStdout:     printDataToStdout,
	}

	return tss, nil
}

// Listeners returns the TraceStreamingService's underlying WriteListeners, use for registering them with the BaseApp
func (tss *TraceStreamingService) Listeners() map[types.StoreKey][]types.WriteListener {
	return tss.listeners
}

// ListenBeginBlock satisfies the Hook interface
// It writes out the received BeginBlockEvent request and response and the resulting state changes to the log
func (tss *TraceStreamingService) ListenBeginBlock(
	ctx sdk.Context,
	req abci.RequestBeginBlock,
	res abci.ResponseBeginBlock,
) error {
	tss.setBeginBlock(req)
	eventId := int64(1)
	eventTypeId := 1

	// write req
	key := fmt.Sprintf(LogMsgFmt, tss.currentBlockNumber, BeginBlockEvent, eventId, RequestEventType, eventTypeId)
	if err := tss.writeEventReqRes(ctx, key, &req); err != nil {
		return err
	}

	// write state changes
	if err := tss.writeStateChange(ctx, string(BeginBlockEvent), eventId); err != nil {
		return err
	}

	// write res
	key = fmt.Sprintf(LogMsgFmt, tss.currentBlockNumber, EndBlockEvent, 1, ResponseEventType, 1)
	if err := tss.writeEventReqRes(ctx, key, &res); err != nil {
		return err
	}

	return nil
}

func (tss *TraceStreamingService) setBeginBlock(req abci.RequestBeginBlock) {
	tss.currentBlockNumber = req.GetHeader().Height
	tss.currentTxIndex = 0
	tss.deliveredBlockChan = make(chan struct{})
	tss.deliveredMessages = true // Reset to true. Will be set to false when delivery of any message fails.
}

// ListenDeliverTx satisfies the Hook interface
// It writes out the received DeliverTxEvent request and response and the resulting state changes out to a file as described
// in the above the naming schema
func (tss *TraceStreamingService) ListenDeliverTx(
	ctx sdk.Context,
	req abci.RequestDeliverTx,
	res abci.ResponseDeliverTx,
) error {
	eventId := tss.getDeliverTxId()
	eventTypeId := 1

	// write req
	key := fmt.Sprintf(LogMsgFmt, tss.currentBlockNumber, DeliverTxEvent, eventId, RequestEventType, eventTypeId)
	if err := tss.writeEventReqRes(ctx, key, &req); err != nil {
		return err
	}

	// write state changes
	if err := tss.writeStateChange(ctx, DeliverTxEvent, eventId); err != nil {
		return err
	}

	// write res
	key = fmt.Sprintf(LogMsgFmt, tss.currentBlockNumber, DeliverTxEvent, eventId, ResponseEventType, 1)
	if err := tss.writeEventReqRes(ctx, key, &res); err != nil {
		return err
	}

	return nil
}

func (tss *TraceStreamingService) getDeliverTxId() int64 {
	tss.currentTxIndex++
	return tss.currentTxIndex
}

// ListenEndBlock satisfies the Hook interface
// It writes out the received EndBlockEvent request and response and the resulting state changes out to a file as described
// in the above the naming schema
func (tss *TraceStreamingService) ListenEndBlock(
	ctx sdk.Context,
	req abci.RequestEndBlock,
	res abci.ResponseEndBlock,
) error {
	eventId := int64(1)
	eventTypeId := 1

	// write req
	key := fmt.Sprintf(LogMsgFmt, tss.currentBlockNumber, EndBlockEvent, eventId, RequestEventType, eventTypeId)
	if err := tss.writeEventReqRes(ctx, key, &req); err != nil {
		return err
	}

	// write state changes
	if err := tss.writeStateChange(ctx, EndBlockEvent, eventId); err != nil {
		return err
	}

	// write res
	key = fmt.Sprintf(LogMsgFmt, tss.currentBlockNumber, EndBlockEvent, eventId, ResponseEventType, eventTypeId)
	if err := tss.writeEventReqRes(ctx, key, &res); err != nil {
		return err
	}

	// Acknowledge that the EndBlockEvent request, response and state changes have been written
	close(tss.deliveredBlockChan)
	return nil
}

// ListenSuccess returns a chan that is used to acknowledge successful receipt of messages by the external service
// after some configurable delay, `false` is sent to this channel from the service to signify failure of receipt.
// For fire-and-forget model, set the chan to always be `true`:
//
//   func (tss *TraceStreamingService) ListenSuccess() <-chan bool {
//       tss.successChan <- true
//	     return tss.successChan
//   }
func (tss *TraceStreamingService) ListenSuccess() <-chan bool {
	// Synchronize the work between app.Commit() and message writes for the current block.
	// Wait until ListenEndBlock() is finished or timeout is reached before responding back.
	var deliveredBlock bool
	maxWait := time.NewTicker(tss.deliverBlockWaitLimit)
	defer maxWait.Stop()
	loop:
		for {
			select {
			case <-tss.deliveredBlockChan:
				deliveredBlock = true
				break loop
			case <-maxWait.C:
				deliveredBlock = false
				break loop
			}
		}

	if deliveredBlock == false {
		tss.deliveredMessages = false
	}

	tss.successChan <- tss.deliveredMessages
	return tss.successChan
}

// Stream spins up a goroutine select loop which awaits length-prefixed binary encoded KV pairs and caches them in the order they were received
// Do we need this and an intermediate writer? We could just write directly to the buffer on calls to Write
// But then we don't support a Stream interface, which could be needed for other types of streamers
func (tss *TraceStreamingService) Stream(wg *sync.WaitGroup) error {
	if tss.quitChan != nil {
		return errors.New("`Stream` has already been called. The stream needs to be closed before it can be started again")
	}
	tss.quitChan = make(chan struct{})
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-tss.quitChan:
				return
			case by := <-tss.srcChan:
				tss.stateCacheLock.Lock()
				tss.stateCache = append(tss.stateCache, by)
				tss.stateCacheLock.Unlock()
			}
		}
	}()
	return nil
}

// Close satisfies the io.Closer interface
func (tss *TraceStreamingService) Close() error {
	close(tss.quitChan)
	return nil
}

func (tss *TraceStreamingService) writeStateChange(ctx sdk.Context, event string, eventId int64) error {
	// write all state changes cached for this stage
	tss.stateCacheLock.Lock()
	kodec := tss.codec.(*codec.ProtoCodec)
	kvPair := new(types.StoreKVPair)
	for i, stateChange := range tss.stateCache {
		key := fmt.Sprintf(LogMsgFmt, tss.currentBlockNumber, event, eventId, StateChangeEventType, i+1)
		if err := kodec.UnmarshalLengthPrefixed(stateChange, kvPair); err != nil {
			return err
		}
		if err := tss.writeEventReqRes(ctx, key, kvPair); err != nil {
			return err
		}
	}

	// reset cache
	tss.stateCache = nil
	tss.stateCacheLock.Unlock()

	return nil
}

func (tss *TraceStreamingService) writeEventReqRes(ctx sdk.Context, key string, data proto.Message) error {
	var m = fmt.Sprintf("%v => data:omitted", key)
	if tss.printDataToStdout {
		m = fmt.Sprintf("%v => data:%v", key, data)
	}
	ctx.Logger().Debug(m)
	return nil
}