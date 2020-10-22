package tx

import (
	gocontext "context"
	"fmt"

	"google.golang.org/grpc"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Invoke implements the grpc ClientConn.Invoke method.
func (w *wrapper) Invoke(_ gocontext.Context, method string, args, reply interface{}, _ ...grpc.CallOption) error {
	req, ok := args.(sdk.MsgRequest)
	if !ok {
		return fmt.Errorf("%T should implement %T", args, (*sdk.MsgRequest)(nil))
	}

	err := req.ValidateBasic()
	if err != nil {
		return err
	}

	w.SetMsgs(sdk.ServiceMsg{
		MethodName: method,
		Request:    req,
	})

	return nil
}

// NewStream implements the grpc ClientConn.NewStream method.
func (w *wrapper) NewStream(gocontext.Context, *grpc.StreamDesc, string, ...grpc.CallOption) (grpc.ClientStream, error) {
	return nil, fmt.Errorf("not supported")
}
