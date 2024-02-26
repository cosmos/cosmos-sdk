package runtime

import (
	"context"

	"github.com/cosmos/gogoproto/proto"
	protov2 "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/runtime/protoiface"

	"cosmossdk.io/core/router"
	"cosmossdk.io/core/store"

	"github.com/cosmos/cosmos-sdk/baseapp"
)

func NewMsgRouterService(storeService store.KVStoreService, router *baseapp.MsgServiceRouter) router.Service {
	return &msgRouterService{
		storeService: storeService,
		router:       router,
	}
}

type msgRouterService struct {
	storeService store.KVStoreService
	router       *baseapp.MsgServiceRouter
}

// InvokeTyped implements router.Service.
func (m *msgRouterService) InvokeTyped(ctx context.Context, req, res protoiface.MessageV1) error {
	messageName := msgTypeURL(req)
	return m.router.HybridHandlerByMsgName(messageName)(ctx, req, res)
}

// msgTypeURL returns the TypeURL of a `sdk.Msg`.
func msgTypeURL(msg proto.Message) string {
	if m, ok := msg.(protov2.Message); ok {
		return "/" + string(m.ProtoReflect().Descriptor().FullName())
	}

	return "/" + proto.MessageName(msg)
}
