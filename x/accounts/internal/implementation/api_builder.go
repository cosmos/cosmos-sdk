package implementation

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"
)

var (
	errNoInitHandler    = errors.New("no init handler")
	errNoExecuteHandler = errors.New("account does not accept messages")
	errInvalidMessage   = errors.New("invalid message")
)

// NewInitBuilder creates a new InitBuilder instance.
func NewInitBuilder() *InitBuilder {
	return &InitBuilder{}
}

// InitBuilder defines a smart account's initialisation handler builder.
type InitBuilder struct {
	// handler is the handler function that will be called when the smart account is initialized.
	// Although the function here is defined to take an any, the smart account will work
	// with a typed version of it.
	handler func(ctx context.Context, initRequest any) (initResponse any, err error)

	// decodeRequest is the function that will be used to decode the init request from bytes.
	decodeRequest func([]byte) (any, error)

	// encodeResponse is the function that will be used to encode the init response to bytes.
	encodeResponse func(any) ([]byte, error)
}

// makeHandler returns the handler function that will be called when the smart account is initialized.
// It returns an error if no handler was registered.
func (i *InitBuilder) makeHandler() (func(ctx context.Context, initRequest any) (initResponse any, err error), error) {
	if i.handler == nil {
		return nil, errNoInitHandler
	}
	return i.handler, nil
}

// NewExecuteBuilder creates a new ExecuteBuilder instance.
func NewExecuteBuilder() *ExecuteBuilder {
	return &ExecuteBuilder{
		handlers: make(map[string]func(ctx context.Context, executeRequest any) (executeResponse any, err error)),
	}
}

// ExecuteBuilder defines a smart account's execution router, it will be used to map an execution message
// to a handler function for a specific account.
type ExecuteBuilder struct {
	// handlers is a map of handler functions that will be called when the smart account is executed.
	handlers map[string]func(ctx context.Context, executeRequest any) (executeResponse any, err error)
	// err is the error that occurred before building the handler function.
	err error
}

func (r *ExecuteBuilder) getMessageName(msg any) (string, error) {
	protoMsg, ok := msg.(protoreflect.ProtoMessage)
	if !ok {
		return "", fmt.Errorf("%w: expected protoreflect.Message, got %T", errInvalidMessage, msg)
	}
	return string(protoMsg.ProtoReflect().Descriptor().FullName()), nil
}

func (r *ExecuteBuilder) makeHandler() (func(ctx context.Context, executeRequest any) (executeResponse any, err error), error) {
	// if no handler is registered it's fine, it means the account will not be accepting execution or query messages.
	if len(r.handlers) == 0 {
		return func(ctx context.Context, _ any) (_ any, err error) {
			return nil, errNoExecuteHandler
		}, nil
	}

	if r.err != nil {
		return nil, r.err
	}

	// build the real execution handler
	return func(ctx context.Context, executeRequest any) (executeResponse any, err error) {
		messageName, err := r.getMessageName(executeRequest)
		if err != nil {
			return nil, fmt.Errorf("%w: unable to get message name", err)
		}
		handler, ok := r.handlers[messageName]
		if !ok {
			return nil, fmt.Errorf("%w: no handler for message %s", errInvalidMessage, messageName)
		}
		return handler(ctx, executeRequest)
	}, nil
}

func (r *ExecuteBuilder) makeRequestDecoder() func(requestBytes []byte) (any, error) {
	return func(requestBytes []byte) (any, error) {
		anyPB := new(anypb.Any)
		err := proto.Unmarshal(requestBytes, anyPB)
		if err != nil {
			return nil, err
		}

		msg, err := anyPB.UnmarshalNew()
		if err != nil {
			return nil, err
		}

		// we do not check if it is part of a valid message set as an account can handle
		// and the handler will do so.
		return msg, nil
	}
}

func (r *ExecuteBuilder) makeResponseEncoder() func(executeResponse any) ([]byte, error) {
	return func(executeResponse any) ([]byte, error) {
		executeResponsePB, ok := executeResponse.(protoreflect.ProtoMessage)
		if !ok {
			return nil, fmt.Errorf("%w: expected protoreflect.Message, got %T", errInvalidMessage, executeResponse)
		}
		anyPB, err := anypb.New(executeResponsePB)
		if err != nil {
			return nil, err
		}

		// we do not check if it is part of an account's valid response message set
		// as make handler will never allow for an invalid response to be returned.
		return proto.Marshal(anyPB)
	}
}

// NewQueryBuilder creates a new QueryBuilder instance.
func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{
		er: NewExecuteBuilder(),
	}
}

// QueryBuilder defines a smart account's query router, it will be used to map a query message
// to a handler function for a specific account.
type QueryBuilder struct {
	// er is the ExecuteBuilder, since there's no difference between the execution and query handlers API.
	er *ExecuteBuilder
}

func (r *QueryBuilder) makeHandler() (func(ctx context.Context, queryRequest any) (queryResponse any, err error), error) {
	return r.er.makeHandler()
}
