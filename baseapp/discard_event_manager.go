package baseapp

import (
	abci "github.com/cometbft/cometbft/abci/types"
	proto "github.com/cosmos/gogoproto/proto"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// discardEventManager is an EventManager that discards all events.
// This can be used as the EventMangerI implementation for sdk.Context
// when we don't want to collect events for performance reasons.
type discardEventManager struct{}

var _ sdk.EventManagerI = discardEventManager{}

func (d discardEventManager) Events() sdk.Events {
	return sdk.EmptyEvents()
}

func (d discardEventManager) ABCIEvents() []abci.Event {
	return []abci.Event{}
}

func (d discardEventManager) EmitTypedEvent(proto.Message) error {
	return nil
}

func (d discardEventManager) EmitTypedEvents(...proto.Message) error {
	return nil
}

func (d discardEventManager) EmitEvent(sdk.Event) {}

func (d discardEventManager) EmitEvents(sdk.Events) {}
