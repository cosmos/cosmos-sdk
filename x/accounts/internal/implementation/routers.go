package implementation

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/protobuf/reflect/protoreflect"
)

var (
	errNoInitHandler    = errors.New("no init handler")
	errNoExecuteHandler = errors.New("account does not accept messages")
	ErrInvalidMessage   = errors.New("invalid message")
)

// NewInitRouter creates a new InitBuilder instance.
func NewInitRouter() *InitBuilder {
	return &InitBuilder{}
}

// InitBuilder defines a smart account's initialisation handler builder.
type InitBuilder struct {
	// handler is the handler function that will be called when the smart account is initialised.
	// Although the function here is defined to take an interface{}, the smart account will work
	// with a typed version of it.
	handler func(ctx context.Context, initRequest interface{}) (initResponse interface{}, err error)
}

// makeHandler returns the handler function that will be called when the smart account is initialised.
// It returns an error if no handler was registered.
func (i *InitBuilder) makeHandler() (func(ctx context.Context, initRequest interface{}) (initResponse interface{}, err error), error) {
	if i.handler == nil {
		return nil, errNoInitHandler
	}
	return i.handler, nil
}

// RegisterHandler registers a handler function that will be called when the smart account is initialised.
func (i *InitBuilder) RegisterHandler(handler func(ctx context.Context, initRequest interface{}) (initResponse interface{}, err error)) {
	i.handler = handler
}

// NewExecuteRouter creates a new ExecuteRouter instance.
func NewExecuteRouter() *ExecuteRouter {
	return &ExecuteRouter{
		handlers: make(map[string]func(ctx context.Context, executeRequest interface{}) (executeResponse interface{}, err error)),
	}
}

// ExecuteRouter defines a smart account's execution router, it will be used to map an execution message
// to a handler function for a specific account.
type ExecuteRouter struct {
	// handlers is a map of handler functions that will be called when the smart account is executed.
	handlers map[string]func(ctx context.Context, executeRequest interface{}) (executeResponse interface{}, err error)

	// err is the error that occurred before building the handler function.
	err error
}

func (r *ExecuteRouter) getMessageName(msg interface{}) (string, error) {
	protoMsg, ok := msg.(protoreflect.ProtoMessage)
	if !ok {
		return "", fmt.Errorf("%w: expected protoreflect.Message, got %T", ErrInvalidMessage, msg)
	}
	return string(protoMsg.ProtoReflect().Descriptor().FullName()), nil
}

func (r *ExecuteRouter) makeHandler() (func(ctx context.Context, executeRequest interface{}) (executeResponse interface{}, err error), error) {
	// if no handler is registered it's fine, it means the account will not be accepting execution or query messages.
	if len(r.handlers) == 0 {
		return func(ctx context.Context, _ interface{}) (_ interface{}, err error) {
			return nil, errNoExecuteHandler
		}, nil
	}

	if r.err != nil {
		return nil, r.err
	}

	// build the real execution handler
	return func(ctx context.Context, executeRequest interface{}) (executeResponse interface{}, err error) {
		messageName, err := r.getMessageName(executeRequest)
		if err != nil {
			return nil, fmt.Errorf("%w: unable to get message name", err)
		}
		handler, ok := r.handlers[messageName]
		if !ok {
			return nil, fmt.Errorf("%w: no handler for message %s", ErrInvalidMessage, messageName)
		}
		return handler(ctx, executeRequest)
	}, nil
}

// NewQueryRouter creates a new QueryRouter instance.
func NewQueryRouter() *QueryRouter {
	return &QueryRouter{
		er: NewExecuteRouter(),
	}
}

// QueryRouter defines a smart account's query router, it will be used to map a query message
// to a handler function for a specific account.
type QueryRouter struct {
	// er is the ExecuteRouter, since there's no difference between the execution and query handlers API.
	er *ExecuteRouter
}

func (r *QueryRouter) makeHandler() (func(ctx context.Context, queryRequest interface{}) (queryResponse interface{}, err error), error) {
	return r.er.makeHandler()
}
