package msgservice

import (
	"context"
	"fmt"

	gogogrpc "github.com/gogo/protobuf/grpc"
	"google.golang.org/grpc"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ServiceMsgClientConn is an instance of grpc.ClientConn that is used to build
// transactions with MsgClient's (Service Msgs). It is intended to be replaced
// by the work in https://github.com/cosmos/cosmos-sdk/issues/7541 and
// https://github.com/cosmos/cosmos-sdk/issues/8270 when those are ready.
type ServiceMsgClientConn struct {
	msgs []sdk.Msg
}

var _ gogogrpc.ClientConn = &ServiceMsgClientConn{}

// Invoke implements gogogrpc.ClientConn.
func (t *ServiceMsgClientConn) Invoke(_ context.Context, method string, args, _ interface{}, _ ...grpc.CallOption) error {
	req, ok := args.(sdk.MsgRequest)
	if !ok {
		return fmt.Errorf("%T should implement %T", args, (*sdk.MsgRequest)(nil))
	}

	err := req.ValidateBasic()
	if err != nil {
		return err
	}

	t.msgs = append(t.msgs, sdk.ServiceMsg{
		MethodName: method,
		Request:    req,
	})

	return nil
}

// NewStream implements gogogrpc.ClientConn.
func (t *ServiceMsgClientConn) NewStream(context.Context, *grpc.StreamDesc, string, ...grpc.CallOption) (grpc.ClientStream, error) {
	return nil, fmt.Errorf("not supported")
}

// GetMsgs returns all messages in the ServiceMsgClientConn.
func (t *ServiceMsgClientConn) GetMsgs() []sdk.Msg {
	return t.msgs
}
