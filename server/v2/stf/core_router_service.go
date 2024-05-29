package stf

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"google.golang.org/protobuf/runtime/protoiface"

	appmodulev2 "cosmossdk.io/core/appmodule/v2"
	"cosmossdk.io/core/router"
)

// NewMsgRouterService implements router.Service.
func NewMsgRouterService(msgRouterBuilder *MsgRouterBuilder) router.Service {
	msgRouter, err := msgRouterBuilder.Build()
	if err != nil {
		panic(fmt.Errorf("cannot create msgRouter: %w", err))
	}

	return &msgRouterService{
		builder: msgRouterBuilder,
		handler: msgRouter,
	}
}

var _ router.Service = (*msgRouterService)(nil)

type msgRouterService struct {
	builder *MsgRouterBuilder
	handler appmodulev2.Handler
}

// CanInvoke returns an error if the given message cannot be invoked.
func (m *msgRouterService) CanInvoke(ctx context.Context, typeURL string) error {
	if typeURL == "" {
		return errors.New("missing type url")
	}

	typeURL = strings.TrimPrefix(typeURL, "/")
	if exists := m.builder.HandlerExists(typeURL); exists {
		return fmt.Errorf("unknown request: %s", typeURL)
	}

	return nil
}

// InvokeTyped execute a message and fill-in a response.
// The response must be known and passed as a parameter.
// Use InvokeUntyped if the response type is not known.
func (m *msgRouterService) InvokeTyped(ctx context.Context, msg, resp protoiface.MessageV1) error {
	// see https://github.com/cosmos/cosmos-sdk/pull/20349
	panic("not implemented")
}

// InvokeUntyped execute a message and returns a response.
func (m *msgRouterService) InvokeUntyped(ctx context.Context, msg protoiface.MessageV1) (protoiface.MessageV1, error) {
	return m.handler(ctx, msg)
}

// NewQueryRouterService implements router.Service.
func NewQueryRouterService(queryRouterBuilder *MsgRouterBuilder) router.Service {
	queryRouter, err := queryRouterBuilder.Build()
	if err != nil {
		panic(fmt.Errorf("cannot create queryRouter: %w", err))
	}

	return &queryRouterService{
		builder: queryRouterBuilder,
		handler: queryRouter,
	}
}

var _ router.Service = (*queryRouterService)(nil)

type queryRouterService struct {
	builder *MsgRouterBuilder
	handler appmodulev2.Handler
}

// CanInvoke returns an error if the given request cannot be invoked.
func (m *queryRouterService) CanInvoke(ctx context.Context, typeURL string) error {
	if typeURL == "" {
		return errors.New("missing type url")
	}

	typeURL = strings.TrimPrefix(typeURL, "/")
	if exists := m.builder.HandlerExists(typeURL); exists {
		return fmt.Errorf("unknown request: %s", typeURL)
	}

	return nil
}

// InvokeTyped execute a message and fill-in a response.
// The response must be known and passed as a parameter.
// Use InvokeUntyped if the response type is not known.
func (m *queryRouterService) InvokeTyped(
	ctx context.Context,
	req, resp protoiface.MessageV1,
) error {
	// see https://github.com/cosmos/cosmos-sdk/pull/20349
	panic("not implemented")
}

// InvokeUntyped execute a message and returns a response.
func (m *queryRouterService) InvokeUntyped(
	ctx context.Context,
	req protoiface.MessageV1,
) (protoiface.MessageV1, error) {
	return m.handler(ctx, req)
}
