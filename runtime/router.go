package runtime

import (
	"context"
	"fmt"
	"reflect"

	"github.com/cosmos/gogoproto/proto"
	protov2 "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/runtime/protoiface"

	"cosmossdk.io/core/router"
	"cosmossdk.io/core/store"

	"github.com/cosmos/cosmos-sdk/baseapp"
)

// NewRouterService creates a router.Service which allows to invoke messages and queries using the msg router.
func NewRouterService(storeService store.KVStoreService, queryRouter *baseapp.GRPCQueryRouter, msgRouter baseapp.MessageRouter) router.Router {
	return &routerService{
		queryRouterService: &queryRouterService{
			storeService: storeService, // TODO: this will be used later on as authenticating modules before routing
			router:       queryRouter,
		},
		msgRouterService: &msgRouterService{
			storeService: storeService, // TODO: this will be used later on as authenticating modules before routing
			router:       msgRouter,
		},
	}
}

var _ router.Router = (*routerService)(nil)

type routerService struct {
	queryRouterService router.Service
	msgRouterService   router.Service
}

// MessageRouterService implements router.Router.
func (r *routerService) MessageRouterService() router.Service {
	return r.msgRouterService
}

// QueryRouterService implements router.Router.
func (r *routerService) QueryRouterService() router.Service {
	return r.queryRouterService
}

var _ router.Service = (*msgRouterService)(nil)

type msgRouterService struct {
	storeService store.KVStoreService
	router       baseapp.MessageRouter
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
	respName := m.router.ResponseNameByMsgName(messageName)
	if respName == "" {
		return nil, fmt.Errorf("could not find response type for message %s (%T)", messageName, msg)
	}

	// get response type
	typ := proto.MessageType(respName)
	if typ == nil {
		return nil, fmt.Errorf("no message type found for %s", respName)
	}
	msgResp, ok := reflect.New(typ.Elem()).Interface().(protoiface.MessageV1)
	if !ok {
		return nil, fmt.Errorf("could not create response message %s", respName)
	}

	return msgResp, m.InvokeTyped(ctx, msg, msgResp)
}

var _ router.Service = (*queryRouterService)(nil)

type queryRouterService struct {
	storeService store.KVStoreService
	router       *baseapp.GRPCQueryRouter
}

// InvokeTyped execute a message and fill-in a response.
// The response must be known and passed as a parameter.
// Use InvokeUntyped if the response type is not known.
func (m *queryRouterService) InvokeTyped(ctx context.Context, req, resp protoiface.MessageV1) error {
	messageName := msgTypeURL(req)
	handlers := m.router.HybridHandlerByRequestName(messageName)
	if len(handlers) == 0 {
		return fmt.Errorf("unknown request: %s", messageName)
	}

	return handlers[0](ctx, req, resp)
}

// InvokeUntyped execute a message and returns a response.
func (m *queryRouterService) InvokeUntyped(ctx context.Context, req protoiface.MessageV1) (protoiface.MessageV1, error) {
	messageName := msgTypeURL(req)
	respName := m.router.ResponseNameByRequestName(messageName)
	if respName == "" {
		return nil, fmt.Errorf("could not find response type for request %s (%T)", messageName, req)
	}

	// get response type
	typ := proto.MessageType(respName)
	if typ == nil {
		return nil, fmt.Errorf("no message type found for %s", respName)
	}
	reqResp, ok := reflect.New(typ.Elem()).Interface().(protoiface.MessageV1)
	if !ok {
		return nil, fmt.Errorf("could not create response request %s", respName)
	}

	return reqResp, m.InvokeTyped(ctx, req, reqResp)
}

// msgTypeURL returns the TypeURL of a proto message.
func msgTypeURL(msg proto.Message) string {
	if m, ok := msg.(protov2.Message); ok {
		return "/" + string(m.ProtoReflect().Descriptor().FullName())
	}

	return "/" + proto.MessageName(msg)
}
