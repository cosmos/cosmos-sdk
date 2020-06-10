package client

import (
	gocontext "context"
	"fmt"

	gogogrpc "github.com/gogo/protobuf/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/encoding/proto"
)

var _ gogogrpc.ClientConn = Context{}

var protoCodec = encoding.GetCodec(proto.Name)

// Invoke implements the grpc ClientConn.Invoke method
func (ctx Context) Invoke(_ gocontext.Context, method string, args, reply interface{}, _ ...grpc.CallOption) error {
	reqBz, err := protoCodec.Marshal(args)
	if err != nil {
		return err
	}
	resBz, _, err := ctx.QueryWithData(method, reqBz)
	if err != nil {
		return err
	}
	return protoCodec.Unmarshal(resBz, reply)
}

// NewStream implements the grpc ClientConn.NewStream method
func (Context) NewStream(gocontext.Context, *grpc.StreamDesc, string, ...grpc.CallOption) (grpc.ClientStream, error) {
	return nil, fmt.Errorf("streaming rpc not supported")
}
