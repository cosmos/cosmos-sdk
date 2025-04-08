package baseapp

import (
	abci "github.com/cometbft/cometbft/abci/types"
	proto "github.com/cosmos/gogoproto/proto"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// discardingEventManager is an EventManager that discards all events.
type discardingEventManager struct{}

func (d discardingEventManager) Events() sdk.Events {
	return sdk.EmptyEvents()
}

func (d discardingEventManager) ABCIEvents() []abci.Event {
	return []abci.Event{}
}

func (d discardingEventManager) EmitTypedEvent(proto.Message) error {
	return nil
}

func (d discardingEventManager) EmitTypedEvents(...proto.Message) error {
	return nil
}

func (d discardingEventManager) EmitEvent(sdk.Event) {}

func (d discardingEventManager) EmitEvents(sdk.Events) {}

var _ sdk.EventManagerI = discardingEventManager{}
