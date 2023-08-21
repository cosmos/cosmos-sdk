package implementation

import (
	"context"
	"fmt"

	"cosmossdk.io/x/accounts/internal/implementation"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var getMessageName = func(msg interface{}) (string, error) {
	protoMsg, ok := msg.(protoreflect.Message)
	if !ok {
		return "", fmt.Errorf("message is not a protobuf message")
	}
	return string(protoMsg.Descriptor().FullName()), nil
}

// proto_account.go defines facilities to build a smart account relying on protobuf messages.

// ProtoMsg is a generic interface for protobuf messages.
type ProtoMsg[T any] interface {
	*T
	protoreflect.Message
}

// RegisterInitHandler registers an initialisation handler for a smart account that uses protobuf.
func RegisterInitHandler[
	Req any, ProtoReq ProtoMsg[Req], Resp any, ProtoResp ProtoMsg[Resp],
](router *implementation.InitBuilder, handler func(req *ProtoReq) (*ProtoResp, error)) {
	reqName := ProtoReq(new(Req)).Descriptor().FullName()
	router.RegisterHandler(func(ctx context.Context, initRequest interface{}) (initResponse interface{}, err error) {
		concrete, ok := initRequest.(*ProtoReq)
		if !ok {
			return nil, fmt.Errorf("%w: wanted %s, got %T", implementation.ErrInvalidMessage, reqName, initRequest)
		}
		return handler(concrete)
	})
}

// RegisterExecuteHandler registers an execution handler for a smart account that uses protobuf.
func RegisterExecuteHandler[
	Req any, ProtoReq ProtoMsg[Req], Resp any, ProtoResp ProtoMsg[Resp],
](router *implementation.ExecuteRouter, handler func(req *ProtoReq) (*ProtoResp, error)) {
	// check if message name is registered.
	if router.getMessageName == nil {
		router.getMessageName = getMessageName
	} else {
		// check if equal
		if &router.getMessageName != &getMessageName {
			router.err = fmt.Errorf("message name function already registered")
			return
		}
	}

	reqName := ProtoReq(new(Req)).Descriptor().FullName()
	// check if not registered already
	if _, ok := router.handlers[string(reqName)]; ok {
		router.err = fmt.Errorf("handler already registered for message %s", reqName)
		return
	}

	router.handlers[string(reqName)] = func(ctx context.Context, executeRequest interface{}) (executeResponse interface{}, err error) {
		concrete, ok := executeRequest.(*ProtoReq)
		if !ok {
			return nil, fmt.Errorf("%w: wanted %s, got %T", implementation.ErrInvalidMessage, reqName, executeRequest)
		}
		return handler(concrete)
	}
}

// RegisterQueryHandler registers a query handler for a smart account that uses protobuf.
func RegisterQueryHandler[
	Req any, ProtoReq ProtoMsg[Req], Resp any, ProtoResp ProtoMsg[Resp],
](router *implementation.QueryRouter, handler func(req *ProtoReq) (*ProtoResp, error)) {

}
