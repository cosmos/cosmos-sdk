package legacytx

import (
	gocontext "context"
	"fmt"

	"google.golang.org/grpc"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Invoke implements the grpc ClientConn.Invoke method. This is so that we can
// use ADR-031 service `Msg`s with StdTxBuilder. Invoking this method will
// **append** the service Msg into the TxBuilder's Msgs array.
// TODO Full amino support still needs to be added as part of https://github.com/cosmos/cosmos-sdk/issues/7541.
func (s *StdTxBuilder) Invoke(_ gocontext.Context, method string, args, reply interface{}, _ ...grpc.CallOption) error {
	req, ok := args.(sdk.MsgRequest)
	if !ok {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "%T should implement %T", args, (*sdk.MsgRequest)(nil))
	}

	msgs := append(s.Msgs, sdk.ServiceMsg{
		MethodName: method,
		Request:    req,
	})
	s.SetMsgs(msgs...)

	return nil
}

// NewStream implements the grpc ClientConn.NewStream method.
func (s *StdTxBuilder) NewStream(gocontext.Context, *grpc.StreamDesc, string, ...grpc.CallOption) (grpc.ClientStream, error) {
	return nil, fmt.Errorf("not supported")
}
