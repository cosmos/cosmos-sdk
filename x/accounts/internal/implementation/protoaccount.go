package implementation

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/proto"
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
](router *InitBuilder, handler func(ctx context.Context, req ProtoReq) (ProtoResp, error)) {
	reqName := ProtoReq(new(Req)).ProtoReflect().Descriptor().FullName()
	respName := ProtoResp(new(Resp)).ProtoReflect().Descriptor().FullName()

	router.handler = func(ctx context.Context, initRequest any) (initResponse any, err error) {
		concrete, ok := initRequest.(ProtoReq)
		if !ok {
			return nil, fmt.Errorf("%w: wanted %s, got %T", errInvalidMessage, reqName, initRequest)
		}
		return handler(ctx, concrete)
	}

	router.decodeRequest = func(b []byte) (any, error) {
		req := new(Req)
		err := proto.Unmarshal(b, ProtoReq(req))
		return req, err
	}

	router.encodeResponse = func(resp any) ([]byte, error) {
		protoResp, ok := resp.(ProtoResp)
		if !ok {
			return nil, fmt.Errorf("%w: wanted %s, got %T", errInvalidMessage, respName, resp)
		}
		return proto.Marshal(protoResp)
	}
}

// RegisterExecuteHandler registers an execution handler for a smart account that uses protobuf.
func RegisterExecuteHandler[
	Req any, ProtoReq ProtoMsg[Req], Resp any, ProtoResp ProtoMsg[Resp],
](router *ExecuteBuilder, handler func(ctx context.Context, req ProtoReq) (ProtoResp, error)) {
	reqName := ProtoReq(new(Req)).ProtoReflect().Descriptor().FullName()
	// check if not registered already
	if _, ok := router.handlers[string(reqName)]; ok {
		router.err = fmt.Errorf("handler already registered for message %s", reqName)
		return
	}

	router.handlers[string(reqName)] = func(ctx context.Context, executeRequest any) (executeResponse any, err error) {
		concrete, ok := executeRequest.(ProtoReq)
		if !ok {
			return nil, fmt.Errorf("%w: wanted %s, got %T", errInvalidMessage, reqName, executeRequest)
		}
		return handler(ctx, concrete)
	}
}

// RegisterQueryHandler registers a query handler for a smart account that uses protobuf.
func RegisterQueryHandler[
	Req any, ProtoReq ProtoMsg[Req], Resp any, ProtoResp ProtoMsg[Resp],
](router *QueryBuilder, handler func(ctx context.Context, req ProtoReq) (ProtoResp, error)) {
	RegisterExecuteHandler(router.er, handler)
}
