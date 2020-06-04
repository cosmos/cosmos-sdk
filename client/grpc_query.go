package client

import (
	gocontext "context"
	"fmt"

	gogogrpc "github.com/gogo/protobuf/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/encoding/proto"
)

// QueryConn returns a new grpc ClientConn for making grpc query calls that
// get routed to the node's ABCI query handler
func (ctx Context) QueryConn() gogogrpc.ClientConn {
	return cliQueryConn{ctx}
}

type cliQueryConn struct {
	ctx Context
}

var _ gogogrpc.ClientConn = cliQueryConn{}

var protoCodec = encoding.GetCodec(proto.Name)

// Invoke implements the grpc ClientConn.Invoke method
func (c cliQueryConn) Invoke(_ gocontext.Context, method string, args, reply interface{}, _ ...grpc.CallOption) error {
	reqBz, err := protoCodec.Marshal(args)
	if err != nil {
		return err
	}
	resBz, _, err := c.ctx.QueryWithData(method, reqBz)
	if err != nil {
		return err
	}
	return protoCodec.Unmarshal(resBz, reply)
}

// NewStream implements the grpc ClientConn.NewStream method
func (c cliQueryConn) NewStream(gocontext.Context, *grpc.StreamDesc, string, ...grpc.CallOption) (grpc.ClientStream, error) {
	return nil, fmt.Errorf("streaming rpc not supported")
}
