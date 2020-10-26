package tx

import (
	gocontext "context"
	"fmt"

	"google.golang.org/grpc"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Invoke implements the grpc ClientConn.Invoke method. This is to that we can
// use ADR-031 service `Msg`s with wrapper.
func (w *wrapper) Invoke(_ gocontext.Context, method string, args, reply interface{}, _ ...grpc.CallOption) error {
	req, ok := args.(sdk.MsgRequest)
	if !ok {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "%T should implement %T", args, (*sdk.MsgRequest)(nil))
	}

	w.AppendMsgs(sdk.ServiceMsg{
		MethodName: method,
		Request:    req,
	})

	return nil
}

// NewStream implements the grpc ClientConn.NewStream method.
func (w *wrapper) NewStream(gocontext.Context, *grpc.StreamDesc, string, ...grpc.CallOption) (grpc.ClientStream, error) {
	return nil, fmt.Errorf("not supported")
}
