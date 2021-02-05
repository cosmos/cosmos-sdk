package msgservice

import (
	"context"
	"fmt"

	gogogrpc "github.com/gogo/protobuf/grpc"
	grpc "google.golang.org/grpc"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ gogogrpc.ClientConn = &ServiceMsgClientConn{}

// ServiceMsgClientConn is an instance of grpc.ClientConn that is used to test building
// transactions with MsgClient's. It is intended to be replaced by the work in
// https://github.com/cosmos/cosmos-sdk/issues/7541 when that is ready.
type ServiceMsgClientConn struct {
	msgs []sdk.Msg
}

// Invoke implements the grpc ClientConn.Invoke method
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

// NewStream implements the grpc ClientConn.NewStream method
func (t *ServiceMsgClientConn) NewStream(context.Context, *grpc.StreamDesc, string, ...grpc.CallOption) (grpc.ClientStream, error) {
	return nil, fmt.Errorf("not supported")
}

// GetMsgs returns ServiceMsgClientConn.msgs
func (t *ServiceMsgClientConn) GetMsgs() []sdk.Msg {
	return t.msgs
}
