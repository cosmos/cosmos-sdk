package runtime

import (
	"context"
	"fmt"

	"github.com/cosmos/gogoproto/proto"
	protov2 "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/runtime/protoiface"

	"cosmossdk.io/core/router"
	"cosmossdk.io/core/store"

	"github.com/cosmos/cosmos-sdk/baseapp"
)

func NewMsgRouterService(storeService store.KVStoreService, router *baseapp.MsgServiceRouter) router.Service {
	return &msgRouterService{
		storeService: storeService,
		router:       router,
		resolver:     protoregistry.GlobalTypes,
	}
}

type msgRouterService struct {
	storeService store.KVStoreService
	router       *baseapp.MsgServiceRouter
	resolver     protoregistry.MessageTypeResolver
}

// InvokeTyped implements router.Service.
func (m *msgRouterService) InvokeTyped(ctx context.Context, msg, msgResp protoiface.MessageV1) error {
	messageName := msgTypeURL(msg)
	handler := m.router.HybridHandlerByMsgName(messageName)
	if handler == nil {
		return fmt.Errorf("unknown message: %s", messageName)
	}

	return handler(ctx, msg, msgResp)
}

// InvokeUntyped implements router.Service.
func (m *msgRouterService) InvokeUntyped(ctx context.Context, msg protoiface.MessageV1) (protoiface.MessageV1, error) {
	messageName := msgTypeURL(msg)
	respName := m.router.ResponseNameByRequestName(messageName)
	if respName == "" {
		return nil, fmt.Errorf("could not find response type for message %T", msg)
	}

	// get response type
	resp, err := m.resolver.FindMessageByName(protoreflect.FullName(respName))
	if err != nil {
		return nil, err
	}

	handler := m.router.HybridHandlerByMsgName(messageName)
	if handler == nil {
		return nil, fmt.Errorf("unknown message: %s", messageName)
	}

	msgResp := resp.New().Interface().(protoiface.MessageV1)
	err = handler(ctx, msg, msgResp)
	return msgResp, err
}

// msgTypeURL returns the TypeURL of a `sdk.Msg`.
func msgTypeURL(msg proto.Message) string {
	if m, ok := msg.(protov2.Message); ok {
		return "/" + string(m.ProtoReflect().Descriptor().FullName())
	}

	return "/" + proto.MessageName(msg)
}
