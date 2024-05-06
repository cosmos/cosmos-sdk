package runtime

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/cosmos/gogoproto/proto"
	protov2 "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/runtime/protoiface"

	"cosmossdk.io/core/router"
	"cosmossdk.io/core/store"

	"github.com/cosmos/cosmos-sdk/baseapp"
)

// NewRouterService creates a router.Service which allows to invoke messages and queries using the msg router.
// NOTE: modulesAllowList should eventually allow more customizable permissions (module x module or module x module message)
// Currently a model present in the module allow list map, can use the msg router service. When no modules are provided, all modules can use both routers.
// A module is always allowed to use the query router.
func NewRouterService(storeService store.KVStoreService, queryRouter baseapp.QueryRouter, msgRouter baseapp.MessageRouter, modulesAllowList map[string]bool) router.Service {
	return &routerService{
		queryRouterService: &queryRouterService{
			router: queryRouter,
		},
		msgRouterService: &msgRouterService{
			storeService:     storeService,
			modulesAllowList: modulesAllowList,
			router:           msgRouter,
		},
	}
}

var _ router.Service = (*routerService)(nil)

type routerService struct {
	queryRouterService router.Router
	msgRouterService   router.Router
}

// MessageRouterService implements router.Service.
func (r *routerService) MessageRouterService() router.Router {
	return r.msgRouterService
}

// QueryRouterService implements router.Service.
func (r *routerService) QueryRouterService() router.Router {
	return r.queryRouterService
}

var _ router.Router = (*msgRouterService)(nil)

type msgRouterService struct {
	storeService     store.KVStoreService
	router           baseapp.MessageRouter
	modulesAllowList map[string]bool
}

// CanInvoke returns an error if the given message cannot be invoked.
func (m *msgRouterService) CanInvoke(ctx context.Context, typeURL string) error {
	if err := m.isAllowed(ctx); err != nil {
		return err
	}

	if typeURL == "" {
		return fmt.Errorf("missing type url")
	}

	typeURL = strings.TrimPrefix(typeURL, "/")

	handler := m.router.HybridHandlerByMsgName(typeURL)
	if handler == nil {
		return fmt.Errorf("unknown message: %s", typeURL)
	}

	return nil
}

// InvokeTyped execute a message and fill-in a response.
// The response must be known and passed as a parameter.
// Use InvokeUntyped if the response type is not known.
func (m *msgRouterService) InvokeTyped(ctx context.Context, msg, resp protoiface.MessageV1) error {
	if err := m.isAllowed(ctx); err != nil {
		return err
	}

	messageName := msgTypeURL(msg)
	handler := m.router.HybridHandlerByMsgName(messageName)
	if handler == nil {
		return fmt.Errorf("unknown message: %s", messageName)
	}

	return handler(ctx, msg, resp)
}

// InvokeUntyped execute a message and returns a response.
func (m *msgRouterService) InvokeUntyped(ctx context.Context, msg protoiface.MessageV1) (protoiface.MessageV1, error) {
	if err := m.isAllowed(ctx); err != nil {
		return nil, err
	}

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

func (m *msgRouterService) isAllowed(ctx context.Context) error {
	caller, _ := m.storeService.OpenKVStore(ctx).Get([]byte("storeKey")) // TODO(@julienrbrt): store storeKey/modules of modules at a specific key to enable this.
	if len(caller) > 0 {
		allow, ok := m.modulesAllowList[string(caller)]
		if !allow || !ok {
			return fmt.Errorf("%s not allowed to use msg router service: %v", caller, m.modulesAllowList)
		}
	}

	return nil
}

var _ router.Router = (*queryRouterService)(nil)

type queryRouterService struct {
	router baseapp.QueryRouter
}

// CanInvoke returns an error if the given request cannot be invoked.
func (m *queryRouterService) CanInvoke(ctx context.Context, typeURL string) error {
	if typeURL == "" {
		return fmt.Errorf("missing type url")
	}

	typeURL = strings.TrimPrefix(typeURL, "/")

	handlers := m.router.HybridHandlerByRequestName(typeURL)
	if len(handlers) == 0 {
		return fmt.Errorf("unknown request: %s", typeURL)
	} else if len(handlers) > 1 {
		return fmt.Errorf("ambiguous request, query have multiple handlers: %s", typeURL)
	}

	return nil
}

// InvokeTyped execute a message and fill-in a response.
// The response must be known and passed as a parameter.
// Use InvokeUntyped if the response type is not known.
func (m *queryRouterService) InvokeTyped(ctx context.Context, req, resp protoiface.MessageV1) error {
	reqName := msgTypeURL(req)
	handlers := m.router.HybridHandlerByRequestName(reqName)
	if len(handlers) == 0 {
		return fmt.Errorf("unknown request: %s", reqName)
	} else if len(handlers) > 1 {
		return fmt.Errorf("ambiguous request, query have multiple handlers: %s", reqName)
	}

	return handlers[0](ctx, req, resp)
}

// InvokeUntyped execute a message and returns a response.
func (m *queryRouterService) InvokeUntyped(ctx context.Context, req protoiface.MessageV1) (protoiface.MessageV1, error) {
	reqName := msgTypeURL(req)
	respName := m.router.ResponseNameByRequestName(reqName)
	if respName == "" {
		return nil, fmt.Errorf("could not find response type for request %s (%T)", reqName, req)
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
		return string(m.ProtoReflect().Descriptor().FullName())
	}

	return proto.MessageName(msg)
}
