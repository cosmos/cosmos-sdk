package stf

import (
	"context"
	"fmt"
	"strings"

	"cosmossdk.io/core/router"
	"cosmossdk.io/core/store"
	"google.golang.org/protobuf/runtime/protoiface"
)

// NewRouterService creates a router.Service which allows to invoke messages and queries using the msg router.
func NewRouterService(storeService store.KVStoreService, queryRouter interface{}, msgRouter interface{}) router.Service {
	return &routerService{
		queryRouterService: &queryRouterService{
			storeService: storeService, // TODO: this will be used later on for authenticating modules before routing
			router:       queryRouter,
		},
		msgRouterService: &msgRouterService{
			storeService: storeService, // TODO: this will be used later on for authenticating modules before routing
			router:       msgRouter,
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
	storeService store.KVStoreService
	router       interface{}
}

// CanInvoke returns an error if the given message cannot be invoked.
func (m *msgRouterService) CanInvoke(ctx context.Context, typeURL string) error {
	if typeURL == "" {
		return fmt.Errorf("missing type url")
	}

	_ = strings.TrimPrefix(typeURL, "/")

	return nil
}

// InvokeTyped execute a message and fill-in a response.
// The response must be known and passed as a parameter.
// Use InvokeUntyped if the response type is not known.
func (m *msgRouterService) InvokeTyped(ctx context.Context, msg, resp protoiface.MessageV1) error {
	return nil
}

// InvokeUntyped execute a message and returns a response.
func (m *msgRouterService) InvokeUntyped(ctx context.Context, msg protoiface.MessageV1) (protoiface.MessageV1, error) {
	return nil, nil
}

var _ router.Router = (*queryRouterService)(nil)

type queryRouterService struct {
	storeService store.KVStoreService
	router       interface{}
}

// CanInvoke returns an error if the given request cannot be invoked.
func (m *queryRouterService) CanInvoke(ctx context.Context, typeURL string) error {
	if typeURL == "" {
		return fmt.Errorf("missing type url")
	}

	_ = strings.TrimPrefix(typeURL, "/")

	return nil
}

// InvokeTyped execute a message and fill-in a response.
// The response must be known and passed as a parameter.
// Use InvokeUntyped if the response type is not known.
func (m *queryRouterService) InvokeTyped(ctx context.Context, req, resp protoiface.MessageV1) error {
	return nil
}

// InvokeUntyped execute a message and returns a response.
func (m *queryRouterService) InvokeUntyped(ctx context.Context, req protoiface.MessageV1) (protoiface.MessageV1, error) {
	return nil, nil
}
