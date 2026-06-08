package services

import (
	"context"

	gogoproto "github.com/cosmos/gogoproto/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"cosmossdk.io/core/autocli"
)

// AutoCLI gRPC service stubs for cosmos.autocli.v1.Query without importing
// cosmossdk.io/api/cosmos/autocli/v1 (pulsar).

const autoCLIQueryServiceName = "cosmos.autocli.v1.Query"

// AppOptionsRequest is the request type for cosmos.autocli.v1.Query/AppOptions.
type AppOptionsRequest struct{}

func (*AppOptionsRequest) Reset()         {}
func (*AppOptionsRequest) String() string { return "AppOptionsRequest{}" }
func (*AppOptionsRequest) ProtoMessage()  {}

// AppOptionsResponse is the response type for cosmos.autocli.v1.Query/AppOptions.
type AppOptionsResponse struct {
	ModuleOptions map[string]*autocli.ModuleOptions `protobuf:"bytes,1,rep,name=module_options,json=moduleOptions,proto3" json:"module_options,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (*AppOptionsResponse) Reset()         {}
func (*AppOptionsResponse) String() string { return "AppOptionsResponse{}" }
func (*AppOptionsResponse) ProtoMessage()  {}

func init() {
	gogoproto.RegisterType((*AppOptionsRequest)(nil), "cosmos.autocli.v1.AppOptionsRequest")
	gogoproto.RegisterType((*AppOptionsResponse)(nil), "cosmos.autocli.v1.AppOptionsResponse")
}

// AutoCLIQueryServer is the server interface for the autocli query service.
type AutoCLIQueryServer interface {
	AppOptions(context.Context, *AppOptionsRequest) (*AppOptionsResponse, error)
}

// UnimplementedAutoCLIQueryServer must be embedded for forward compatibility.
type UnimplementedAutoCLIQueryServer struct{}

func (UnimplementedAutoCLIQueryServer) AppOptions(context.Context, *AppOptionsRequest) (*AppOptionsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AppOptions not implemented")
}

// RegisterAutoCLIQueryServer registers the AutoCLI query server with the gRPC registrar.
func RegisterAutoCLIQueryServer(s grpc.ServiceRegistrar, srv AutoCLIQueryServer) {
	s.RegisterService(&AutoCLIQuery_ServiceDesc, srv)
}

// AutoCLIQuery_ServiceDesc is the grpc.ServiceDesc for the autocli query service.
var AutoCLIQuery_ServiceDesc = grpc.ServiceDesc{
	ServiceName: autoCLIQueryServiceName,
	HandlerType: (*AutoCLIQueryServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "AppOptions",
			Handler:    _AutoCLIQuery_AppOptions_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "cosmos/autocli/v1/query.proto",
}

func _AutoCLIQuery_AppOptions_Handler(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
	in := new(AppOptionsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AutoCLIQueryServer).AppOptions(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/" + autoCLIQueryServiceName + "/AppOptions",
	}
	handler := func(ctx context.Context, req any) (any, error) {
		return srv.(AutoCLIQueryServer).AppOptions(ctx, req.(*AppOptionsRequest))
	}
	return interceptor(ctx, in, info, handler)
}
