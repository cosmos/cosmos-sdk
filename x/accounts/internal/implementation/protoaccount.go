package implementation

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// ProtoMsg is a generic interface for protobuf messages.
type ProtoMsg[T any] interface {
	*T
	protoreflect.ProtoMessage
}

// RegisterInitHandler registers an initialisation handler for a smart account that uses protobuf.
func RegisterInitHandler[
	Req any, ProtoReq ProtoMsg[Req], Resp any, ProtoResp ProtoMsg[Resp],
](router *InitBuilder, handler func(req ProtoReq) (ProtoResp, error)) {
	reqName := ProtoReq(new(Req)).ProtoReflect().Descriptor().FullName()
	router.RegisterHandler(func(ctx context.Context, initRequest interface{}) (initResponse interface{}, err error) {
		concrete, ok := initRequest.(ProtoReq)
		if !ok {
			return nil, fmt.Errorf("%w: wanted %s, got %T", ErrInvalidMessage, reqName, initRequest)
		}
		return handler(concrete)
	})
}

// RegisterExecuteHandler registers an execution handler for a smart account that uses protobuf.
func RegisterExecuteHandler[
	Req any, ProtoReq ProtoMsg[Req], Resp any, ProtoResp ProtoMsg[Resp],
](router *ExecuteBuilder, handler func(req ProtoReq) (ProtoResp, error)) {
	reqName := ProtoReq(new(Req)).ProtoReflect().Descriptor().FullName()
	// check if not registered already
	if _, ok := router.handlers[string(reqName)]; ok {
		router.err = fmt.Errorf("handler already registered for message %s", reqName)
		return
	}

	router.handlers[string(reqName)] = func(ctx context.Context, executeRequest interface{}) (executeResponse interface{}, err error) {
		concrete, ok := executeRequest.(ProtoReq)
		if !ok {
			return nil, fmt.Errorf("%w: wanted %s, got %T", ErrInvalidMessage, reqName, executeRequest)
		}
		return handler(concrete)
	}
}

// RegisterQueryHandler registers a query handler for a smart account that uses protobuf.
func RegisterQueryHandler[
	Req any, ProtoReq ProtoMsg[Req], Resp any, ProtoResp ProtoMsg[Resp],
](router *QueryBuilder, handler func(req ProtoReq) (ProtoResp, error)) {
	RegisterExecuteHandler(router.er, handler)
}
