package runtime

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	gogoproto "github.com/cosmos/gogoproto/proto"
	protov2 "google.golang.org/protobuf/proto"

	"cosmossdk.io/core/router"

	"github.com/cosmos/cosmos-sdk/baseapp"
)

// NewMsgRouterService return new implementation of router.Service.
func NewMsgRouterService(msgRouter baseapp.MessageRouter) router.Service {
	return &msgRouterService{
		router: msgRouter,
	}
}

var _ router.Service = (*msgRouterService)(nil)

type msgRouterService struct {
	// TODO: eventually authenticate modules to use the message router
	router baseapp.MessageRouter
}

// CanInvoke returns an error if the given message cannot be invoked.
func (m *msgRouterService) CanInvoke(ctx context.Context, typeURL string) error {
	if typeURL == "" {
		return errors.New("missing type url")
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
// Use InvokeUntyped if the response type is unknown.
func (m *msgRouterService) InvokeTyped(ctx context.Context, msg, resp gogoproto.Message) error {
	messageName := msgTypeURL(msg)
	handler := m.router.HybridHandlerByMsgName(messageName)
	if handler == nil {
		return fmt.Errorf("unknown message: %s", messageName)
	}

	return handler(ctx, msg, resp)
}

// InvokeUntyped execute a message and returns a response.
func (m *msgRouterService) InvokeUntyped(ctx context.Context, msg gogoproto.Message) (gogoproto.Message, error) {
	messageName := msgTypeURL(msg)
	respName := m.router.ResponseNameByMsgName(messageName)
	if respName == "" {
		return nil, fmt.Errorf("could not find response type for message %s (%T)", messageName, msg)
	}

	// get response type
	typ := gogoproto.MessageType(respName)
	if typ == nil {
		return nil, fmt.Errorf("no message type found for %s", respName)
	}
	msgResp, ok := reflect.New(typ.Elem()).Interface().(gogoproto.Message)
	if !ok {
		return nil, fmt.Errorf("could not create response message %s", respName)
	}

	return msgResp, m.InvokeTyped(ctx, msg, msgResp)
}

// NewQueryRouterService return new implementation of router.Service.
func NewQueryRouterService(queryRouter baseapp.QueryRouter) router.Service {
	return &queryRouterService{
		router: queryRouter,
	}
}

var _ router.Service = (*queryRouterService)(nil)

type queryRouterService struct {
	router baseapp.QueryRouter
}

// CanInvoke returns an error if the given request cannot be invoked.
func (m *queryRouterService) CanInvoke(ctx context.Context, typeURL string) error {
	if typeURL == "" {
		return errors.New("missing type url")
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
// Use InvokeUntyped if the response type is unknown.
func (m *queryRouterService) InvokeTyped(ctx context.Context, req, resp gogoproto.Message) error {
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
func (m *queryRouterService) InvokeUntyped(ctx context.Context, req gogoproto.Message) (gogoproto.Message, error) {
	reqName := msgTypeURL(req)
	respName := m.router.ResponseNameByRequestName(reqName)
	if respName == "" {
		return nil, fmt.Errorf("could not find response type for request %s (%T)", reqName, req)
	}

	// get response type
	typ := gogoproto.MessageType(respName)
	if typ == nil {
		return nil, fmt.Errorf("no message type found for %s", respName)
	}
	reqResp, ok := reflect.New(typ.Elem()).Interface().(gogoproto.Message)
	if !ok {
		return nil, fmt.Errorf("could not create response request %s", respName)
	}

	return reqResp, m.InvokeTyped(ctx, req, reqResp)
}

// msgTypeURL returns the TypeURL of a proto message.
func msgTypeURL(msg gogoproto.Message) string {
	if m, ok := msg.(protov2.Message); ok {
		return string(m.ProtoReflect().Descriptor().FullName())
	}

	return gogoproto.MessageName(msg)
}
