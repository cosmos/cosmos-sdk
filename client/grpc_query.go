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
func (ctx Context) Invoke(grpcCtx gocontext.Context, method string, args, reply interface{}, opts ...grpc.CallOption) error {
	reqBz, err := protoCodec.Marshal(args)
	if err != nil {
		return err
	}
	resBz, height, err := ctx.QueryWithData(method, reqBz)
	if err != nil {
		return err
	}

	err = protoCodec.Unmarshal(resBz, reply)
	if err != nil {
		return err
	}

	// Create header metadata. For now the headers contain:
	// - block height
	// We then parse all the call options, if the call option is a
	// HeaderCallOption, then we manually set the value of that header to the
	// metadata.
	md := metadata.Pairs(baseapp.GRPCBlockHeightHeader, strconv.FormatInt(height, 10))
	for _, callOpt := range opts {
		header, ok := callOpt.(grpc.HeaderCallOption)
		if !ok {
			continue
		}

		*header.HeaderAddr = md
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
