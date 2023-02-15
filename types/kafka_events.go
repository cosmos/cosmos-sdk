package types

import "cosmossdk.io/api/tendermint/abci"

type (
	// Event is a type alias for an ABCI Event
	Event abci.Event

	// Events defines a slice of Event objects
	Events []Event
)

type KafkaBlockEventManager {
	blockHeight,
	txHashes = string[],
	time,
	// events per transaction hash
	events = IndexerTendermintEvent[][],
}

func NewKafkaBlockEventManager() *KafkaBlockEventManager {
	return &KafkaBlockEventManager{EmptyEvents()}
}

func AddEvent(event Event) {
	eventManager.event = eventManager.event.AppendEvent(event)
}

func (eventManager *KafkaBlockEventManager) GetEvents() Events {
	return eventManager.event
}

// prepare for sending

func (eventManager *KafkaBlockEventManager) PrepareForSending() BlockEvent {
	blockHeight,
	txHashes,
	time,
	IndexerTendermintEvent[],
}

// protobuf
type IndexerTendermintEvent struct {
	//eventIndex is just the index in the txn array
	subtype string
	data bytes
	// block event vs txn event? what is this? do we need this?
}
