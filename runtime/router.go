package runtime

import (
	"context"
	"fmt"
	"reflect"

	"github.com/cosmos/gogoproto/proto"
	protov2 "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/runtime/protoiface"

	"cosmossdk.io/core/router"
	"cosmossdk.io/core/store"

	"github.com/cosmos/cosmos-sdk/baseapp"
)

func NewMsgRouterService(storeService store.KVStoreService, router baseapp.MessageRouter) router.Service {
	return &msgRouterService{
		storeService: storeService,
		router:       router,
		resolver:     protoregistry.GlobalTypes,
	}
}

type msgRouterService struct {
	storeService store.KVStoreService
	router       baseapp.MessageRouter
	resolver     protoregistry.MessageTypeResolver
}

// InvokeTyped execute a message and fill-in a response.
// The response must be known and passed as a parameter.
// Use InvokeUntyped if the response type is not known.
func (m *msgRouterService) InvokeTyped(ctx context.Context, msg, resp protoiface.MessageV1) error {
	messageName := msgTypeURL(msg)
	handler := m.router.HybridHandlerByMsgName(messageName)
	if handler == nil {
		return fmt.Errorf("unknown message: %s", messageName)
	}

	return handler(ctx, msg, resp)
}

// InvokeUntyped execute a message and returns a response.
func (m *msgRouterService) InvokeUntyped(ctx context.Context, msg protoiface.MessageV1) (protoiface.MessageV1, error) {
	messageName := msgTypeURL(msg)
	respName := m.router.ResponseNameByRequestName(messageName)
	if respName == "" {
		return nil, fmt.Errorf("could not find response type for message %T", msg)
	}

	// get response type
	typ := proto.MessageType(respName)
	if typ == nil {
		return nil, fmt.Errorf("no message type found for %s", respName)
	}
	msgResp := reflect.New(typ.Elem()).Interface().(protoiface.MessageV1)

	handler := m.router.HybridHandlerByMsgName(messageName)
	if handler == nil {
		return nil, fmt.Errorf("unknown message: %s", messageName)
	}

	err := handler(ctx, msg, msgResp)
	return msgResp, err
}

// msgTypeURL returns the TypeURL of a `sdk.Msg`.
func msgTypeURL(msg proto.Message) string {
	if m, ok := msg.(protov2.Message); ok {
		return "/" + string(m.ProtoReflect().Descriptor().FullName())
	}

	return "/" + proto.MessageName(msg)
}
