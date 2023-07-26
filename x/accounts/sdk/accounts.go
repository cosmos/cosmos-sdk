package sdk

import (
	"context"
	"fmt"

	"github.com/cosmos/gogoproto/proto"
)

type Msg[T any] interface {
	*T
	proto.Message
}

type QueryRouter struct{ er *ExecuteRouter }

func (e *QueryRouter) Handler() func(ctx context.Context, msg proto.Message) (proto.Message, error) {
	return e.er.Handler()
}

type ExecuteRouter struct {
	methodsMap map[string]func(ctx context.Context, msg proto.Message) (proto.Message, error)
}

func (e *ExecuteRouter) Handler() func(ctx context.Context, msg proto.Message) (proto.Message, error) {
	if e.methodsMap == nil {
		return func(_ context.Context, _ proto.Message) (proto.Message, error) {
			return nil, fmt.Errorf("this account does not accept execute messages")
		}
	}
	return func(ctx context.Context, msg proto.Message) (proto.Message, error) {
		name := proto.MessageName(msg)
		handler, exists := e.methodsMap[name]
		if !exists {
			return nil, fmt.Errorf("unknown method %s", name)
		}
		return handler(ctx, msg)
	}
}

func RegisterQueryHandler[
	Req any, Resp any, ReqP Msg[Req], RespP Msg[Resp],
](router *QueryRouter, handler func(ctx context.Context, msg Req) (Resp, error)) error {
	if router.er == nil {
		router.er = &ExecuteRouter{}
	}
	return RegisterExecuteHandler[Req, Resp, ReqP, RespP](router.er, handler)
}

func RegisterExecuteHandler[
	Req any, Resp any, ReqP Msg[Req], RespP Msg[Resp],
](router *ExecuteRouter, handler func(ctx context.Context, msg Req) (Resp, error)) error {
	if router.methodsMap == nil {
		router.methodsMap = make(map[string]func(ctx context.Context, msg proto.Message) (proto.Message, error))
	}
	methodName := proto.MessageName(ReqP(new(Req)))
	_, exists := router.methodsMap[methodName]
	if exists {
		return fmt.Errorf("execute method %s already registered", methodName)
	}
	router.methodsMap[methodName] = func(ctx context.Context, msg proto.Message) (proto.Message, error) {
		concreteReq, ok := msg.(ReqP)
		if !ok {
			return nil, fmt.Errorf("invalid message type %T, wanted: %s", msg, methodName)
		}

		resp, err := handler(ctx, *concreteReq)
		if err != nil {
			return nil, err
		}
		return RespP(&resp), nil
	}
	return nil
}

// Account defines a generic account interface.
type Account[IReq, IResp any, IReqP Msg[IReq], IRespP Msg[IResp]] interface {
	// Init is given a request and its duty is to initialise the account.
	Init(ctx context.Context, initMsg IReq) (IResp, error)
	// RegisterExecuteHandlers is given a router and the account should register
	// its execute handlers.
	RegisterExecuteHandlers(router *ExecuteRouter) error
	// RegisterQueryHandlers is given a router and the account should register
	// its query handlers.
	RegisterQueryHandlers(router *QueryRouter) error
}
