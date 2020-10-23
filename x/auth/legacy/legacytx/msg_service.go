package legacytx

import (
	gocontext "context"
	"fmt"

	"google.golang.org/grpc"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Invoke implements the grpc ClientConn.Invoke method.
// TODO Full amino support still needs to be added as part of https://github.com/cosmos/cosmos-sdk/issues/7541.
func (s *StdTxBuilder) Invoke(_ gocontext.Context, method string, args, reply interface{}, _ ...grpc.CallOption) error {
	req, ok := args.(sdk.MsgRequest)
	if !ok {
		return fmt.Errorf("%T should implement %T", args, (*sdk.MsgRequest)(nil))
	}

	s.SetMsgs(sdk.ServiceMsg{
		MethodName: method,
		Request:    req,
	})

	return nil
}

// NewStream implements the grpc ClientConn.NewStream method.
func (s *StdTxBuilder) NewStream(gocontext.Context, *grpc.StreamDesc, string, ...grpc.CallOption) (grpc.ClientStream, error) {
	return nil, fmt.Errorf("not supported")
}
