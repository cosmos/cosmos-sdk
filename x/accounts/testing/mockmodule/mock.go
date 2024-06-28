package mockmodule

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/protobuf/runtime/protoiface"

	"cosmossdk.io/core/router"
)

type msgRouter struct{}

func (m msgRouter) CanInvoke(ctx context.Context, typeURL string) error {
	return errors.New("not implemented")
}

func (m msgRouter) InvokeTyped(ctx context.Context, req, res protoiface.MessageV1) error {
	typedReq, ok := req.(*MsgEcho)
	if !ok {
		return fmt.Errorf("invalid msg request got %T", req)
	}
	typedRes, ok := res.(*MsgEchoResponse)
	if !ok {
		return fmt.Errorf("invalid msg response got %T", res)
	}
	typedRes.MsgEcho = typedReq.Msg
	return nil
}

func (m msgRouter) InvokeUntyped(ctx context.Context, req protoiface.MessageV1) (res protoiface.MessageV1, err error) {
	// TODO implement me
	panic("implement me")
}

func MockMsgRouter() router.Service {
	return msgRouter{}
}

type queryRouter struct{}

func (q queryRouter) CanInvoke(ctx context.Context, typeURL string) error {
	return errors.New("do not call")
}

func (q queryRouter) InvokeTyped(ctx context.Context, req, res protoiface.MessageV1) error {
	typedReq, ok := req.(*QueryEchoRequest)
	if !ok {
		return fmt.Errorf("invalid query request got %T", req)
	}
	typedRes, ok := res.(*QueryEchoResponse)
	if !ok {
		return fmt.Errorf("invalid query response got %T", res)
	}
	typedRes.MsgEcho = typedReq.Msg
	return nil
}

func (q queryRouter) InvokeUntyped(ctx context.Context, req protoiface.MessageV1) (res protoiface.MessageV1, err error) {
	typedReq, ok := req.(*QueryEchoRequest)
	if !ok {
		return nil, fmt.Errorf("invalid request")
	}
	return &QueryEchoResponse{MsgEcho: typedReq.Msg}, nil
}

func MockQueryRouter() router.Service {
	return queryRouter{}
}
