// Package reflection provides the cosmos.reflection.v1.ReflectionService gRPC
// implementation without depending on cosmossdk.io/api/cosmos/reflection/v1.
package reflection

import (
	"context"
	"fmt"

	gogoproto "github.com/cosmos/gogoproto/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/descriptorpb"
)

// ServiceName is the fully-qualified name of the ReflectionService.
const ServiceName = "cosmos.reflection.v1.ReflectionService"

// FileDescriptorsRequest is the request type for FileDescriptors.
// It mirrors cosmos.reflection.v1.FileDescriptorsRequest (empty message).
type FileDescriptorsRequest struct{}

func (r *FileDescriptorsRequest) Reset()         {}
func (r *FileDescriptorsRequest) String() string { return "FileDescriptorsRequest{}" }
func (r *FileDescriptorsRequest) ProtoMessage()  {}

// FileDescriptorsResponse is the response type for FileDescriptors.
// It mirrors cosmos.reflection.v1.FileDescriptorsResponse.
type FileDescriptorsResponse struct {
	Files []*descriptorpb.FileDescriptorProto
}

func (r *FileDescriptorsResponse) Reset() {}
func (r *FileDescriptorsResponse) String() string {
	return fmt.Sprintf("FileDescriptorsResponse{Files: %d}", len(r.Files))
}
func (r *FileDescriptorsResponse) ProtoMessage() {}

// ServiceServer is the server API for the cosmos.reflection.v1.ReflectionService.
type ServiceServer interface {
	FileDescriptors(context.Context, *FileDescriptorsRequest) (*FileDescriptorsResponse, error)
}

// UnimplementedServiceServer must be embedded for forward compatibility.
type UnimplementedServiceServer struct{}

func (UnimplementedServiceServer) FileDescriptors(context.Context, *FileDescriptorsRequest) (*FileDescriptorsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method FileDescriptors not implemented")
}

// RegisterReflectionServiceServer registers srv with the gRPC service registrar.
func RegisterReflectionServiceServer(s grpc.ServiceRegistrar, srv ServiceServer) {
	s.RegisterService(&ReflectionService_ServiceDesc, srv)
}

// ReflectionService_ServiceDesc is the grpc.ServiceDesc for ReflectionService.
var ReflectionService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: ServiceName,
	HandlerType: (*ServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "FileDescriptors",
			Handler:    _ReflectionService_FileDescriptors_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "cosmos/reflection/v1/reflection.proto",
}

func _ReflectionService_FileDescriptors_Handler(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
	in := new(FileDescriptorsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ServiceServer).FileDescriptors(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/cosmos.reflection.v1.ReflectionService/FileDescriptors",
	}
	handler := func(ctx context.Context, req any) (any, error) {
		return srv.(ServiceServer).FileDescriptors(ctx, req.(*FileDescriptorsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func init() {
	gogoproto.RegisterType((*FileDescriptorsRequest)(nil), "cosmos.reflection.v1.FileDescriptorsRequest")
	gogoproto.RegisterType((*FileDescriptorsResponse)(nil), "cosmos.reflection.v1.FileDescriptorsResponse")
}
