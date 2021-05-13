package reflection

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

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

var _ ReflectionServiceServer = (*reflectionServiceServer)(nil)

// ListAllInterfaces implements the ListAllInterfaces method of the
// ReflectionServiceServer interface.
func (r reflectionServiceServer) ListAllInterfaces(_ sdk.Context, _ *ListAllInterfacesRequest) (*ListAllInterfacesResponse, error) {
	ifaces := r.interfaceRegistry.ListAllInterfaces()

	return &ListAllInterfacesResponse{InterfaceNames: ifaces}, nil
}

// ListImplementations implements the ListImplementations method of the
// ReflectionServiceServer interface.
func (r reflectionServiceServer) ListImplementations(_ sdk.Context, req *ListImplementationsRequest) (*ListImplementationsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.InterfaceName == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid interface name")
	}

	impls := r.interfaceRegistry.ListImplementations(req.InterfaceName)

	return &ListImplementationsResponse{ImplementationMessageNames: impls}, nil
}
