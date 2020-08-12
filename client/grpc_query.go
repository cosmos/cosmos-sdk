package client

import (
	gocontext "context"
	"fmt"
	"strconv"

	gogogrpc "github.com/gogo/protobuf/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/encoding/proto"
	"google.golang.org/grpc/metadata"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec/types"
)

var _ gogogrpc.ClientConn = Context{}

var protoCodec = encoding.GetCodec(proto.Name)

// Invoke implements the grpc ClientConn.Invoke method
func (ctx Context) Invoke(grpcCtx gocontext.Context, method string, args, reply interface{}, _ ...grpc.CallOption) error {
	reqBz, err := protoCodec.Marshal(args)
	if err != nil {
		return err
	}
	resBz, height, err := ctx.QueryWithData(method, reqBz)
	if err != nil {
		return err
	}

	fmt.Println("grpcCtx=", grpcCtx)
	// Add GRPCBlockHeightHeader to the gRPC response. To achieve that, we:
	// - create an empty Stream (since we don't have any existing one to attach
	//   to),
	// - add the stream to the grpcCtx,
	// - once the stream is set, we can add headers to it.
	stream, err := grpc.NewClientStream(grpcCtx, &grpc.StreamDesc{}, ctx, method)
	if err != nil {
		return err
	}
	grpcCtx = grpc.NewContextWithServerTransportStream(grpcCtx, stream)

	fmt.Println("HEIGHT=", height)
	md := metadata.Pairs(baseapp.GRPCBlockHeightHeader, strconv.FormatInt(height, 10))

	err = grpc.SetHeader(grpcCtx, md)
	if err != nil {
		return err
	}

	err = protoCodec.Unmarshal(resBz, reply)
	if err != nil {
		return err
	}

	if ctx.InterfaceRegistry != nil {
		return types.UnpackInterfaces(reply, ctx.InterfaceRegistry)
	}

	return nil
}

// NewStream implements the grpc ClientConn.NewStream method
func (Context) NewStream(gocontext.Context, *grpc.StreamDesc, string, ...grpc.CallOption) (grpc.ClientStream, error) {
	return nil, fmt.Errorf("streaming rpc not supported")
}
