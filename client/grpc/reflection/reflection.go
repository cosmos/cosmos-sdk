package reflection

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/codec/types"
)

type reflectionServiceServer struct {
	interfaceRegistry types.InterfaceRegistry
}

// NewReflectionServiceServer creates a new reflectionServiceServer.
func NewReflectionServiceServer(interfaceRegistry types.InterfaceRegistry) ReflectionServiceServer {
	return &reflectionServiceServer{interfaceRegistry: interfaceRegistry}
}

var _ ReflectionServiceServer = &reflectionServiceServer{}

func (r reflectionServiceServer) ListInterfaces(_ context.Context, _ *ListInterfacesRequest) (*ListInterfacesResponse, error) {
	ifaces := r.interfaceRegistry.ListInterfaces()

	return &ListInterfacesResponse{InterfaceNames: ifaces}, nil
}

func (r reflectionServiceServer) ListImplementations(_ context.Context, request *ListImplementationsRequest) (*ListImplementationsResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if len(request.InterfaceName) == 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid interface name")
	}

	impls := r.interfaceRegistry.ListImplementations(request.InterfaceName)

	return &ListImplementationsResponse{ImplementationMessageNames: impls}, nil
}
