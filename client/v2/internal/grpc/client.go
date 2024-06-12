package grpc

import (
	"context"
	"fmt"

	gogogrpc "github.com/cosmos/gogoproto/grpc"
	"google.golang.org/grpc"
)

var _ gogogrpc.ClientConn = ClientConn{}

type ClientConn struct{}

func (c ClientConn) Invoke(ctx context.Context, method string, args, reply interface{}, opts ...grpc.CallOption) error {
	//TODO implement me
	panic("implement me")
}

func (c ClientConn) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	return nil, fmt.Errorf("streaming rpc not supported")
}
