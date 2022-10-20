package event

import (
	"context"

	"google.golang.org/protobuf/proto"
)

type ListenerRegistrar interface {
	// AddEventListener adds an event listener for the provided type to the handler.
	AddEventListener(evtType proto.Message, listener func(context.Context, proto.Message))

	// AddEventInterceptor adds an event interceptor for the provided type to the handler which can
	// return an error to interrupt state machine processing and revert uncommitted state changes.
	AddEventInterceptor(evtType proto.Message, interceptor func(context.Context, proto.Message) error)
}
