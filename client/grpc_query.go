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

	fmt.Println("grpcCtx=", grpcCtx)
	// Add GRPCBlockHeightHeader to the gRPC response. To achieve that, we:
	// - create an empty Stream (since we don't have any existing one to attach
	//   to),
	// - add the stream to the grpcCtx,
	// - once the stream is set, we can add headers to it.
	clientConn := grpc.ClientConn{}
	stream, err := grpc.NewClientStream(grpcCtx, &grpc.StreamDesc{}, &clientConn, method, opts...)
	if err != nil {
		return err
	}
	// myStream := newServerStream(method)
	grpcCtx = grpc.NewContextWithServerTransportStream(grpcCtx, &stream)

	fmt.Println("HEIGHT=", height)
	md := metadata.Pairs(baseapp.GRPCBlockHeightHeader, strconv.FormatInt(height, 10))

	err = grpc.SetHeader(grpcCtx, md)
	if err != nil {
		return err
	}
	// fmt.Println("myStream=", myStream)

	// for _, callOpt := range opts {
	// 	header, ok := callOpt.(grpc.HeaderCallOption)
	// 	if !ok {
	// 		continue
	// 	}

	// 	header.HeaderAddr = &md
	// }

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

var _ grpc.ServerTransportStream = (*serverStream)(nil)

// serverStream is a simple struct that implements grpc.ServerTransportStream.
// As per gRPC docs: "This can be used to mock an actual transport stream for
// tests of handler code that use, for example, grpc.SetHeader (which requires
// some stream to be in context)." This is exactly what we're doing here, apart
// from the fact that it's not in a test. Happy to find an alternate less hacky
// solution.
type serverStream struct {
	method string
	header metadata.MD
}

func (s *serverStream) Method() string {
	return s.method
}

func (s *serverStream) SendHeader(md metadata.MD) error {
	return fmt.Errorf("not supported")
}

func (s serverStream) Header() metadata.MD {
	return s.header
}

func (s *serverStream) SetHeader(md metadata.MD) error {
	s.header = md

	return nil
}

func (s *serverStream) SetTrailer(md metadata.MD) error {
	return fmt.Errorf("not supported")
}

func newServerStream(method string) serverStream {
	return serverStream{
		method: method,
	}
}
