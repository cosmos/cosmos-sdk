package reflection

import (
	"context"
	"fmt"
	"log"
	"reflect"

	"github.com/gogo/protobuf/proto"
	"google.golang.org/grpc"

	sdktypes "github.com/cosmos/cosmos-sdk/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/codec/types"
)

type reflectionServiceServer struct {
	interfaceRegistry types.InterfaceRegistry
	infoProvider      ServiceInfoProvider

	queries []*grpc.ServiceDesc
}

type ServiceInfoProvider interface {
	ListServices() []*grpc.ServiceDesc
}

// NewReflectionServiceServer creates a new reflectionServiceServer.
func NewReflectionServiceServer(interfaceRegistry types.InterfaceRegistry, infoProvider ServiceInfoProvider) ReflectionServiceServer {
	return &reflectionServiceServer{interfaceRegistry: interfaceRegistry, infoProvider: infoProvider}
}

var _ ReflectionServiceServer = (*reflectionServiceServer)(nil)

// ListAllInterfaces implements the ListAllInterfaces method of the
// ReflectionServiceServer interface.
func (r *reflectionServiceServer) ListAllInterfaces(_ context.Context, _ *ListAllInterfacesRequest) (*ListAllInterfacesResponse, error) {
	ifaces := r.interfaceRegistry.ListAllInterfaces()

	return &ListAllInterfacesResponse{InterfaceNames: ifaces}, nil
}

// ListImplementations implements the ListImplementations method of the
// ReflectionServiceServer interface.
func (r *reflectionServiceServer) ListImplementations(_ context.Context, req *ListImplementationsRequest) (*ListImplementationsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.InterfaceName == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid interface name")
	}

	impls := r.interfaceRegistry.ListImplementations(req.InterfaceName)
	protoNames := make([]string, len(impls))

	for i, impl := range impls {
		pb, err := r.interfaceRegistry.Resolve(impl)
		// we should panic but let's return an error
		if err != nil {
			return nil, status.Errorf(codes.NotFound, "can not solve %s: %s", impl, err.Error())
		}
		// we should panic here too
		name := proto.MessageName(pb)
		if name == "" {
			return nil, status.Errorf(codes.NotFound, "can not get proto name for %s")
		}
		protoNames[i] = name
	}
	return &ListImplementationsResponse{
		ImplementationMessageNames:      impls,
		ImplementationMessageProtoNames: protoNames,
	}, nil
}

func (r *reflectionServiceServer) ListDeliverables(_ context.Context, _ *ListDeliverablesRequest) (*ListDeliverablesResponse, error) {
	implementersName := r.interfaceRegistry.ListImplementations(sdktypes.ServiceMsgInterfaceName)

	deliverables := make([]*DeliverableDescriptor, len(implementersName))

	for i, name := range implementersName {
		resolved, err := r.interfaceRegistry.Resolve(name)
		if err != nil {
			return nil, status.Error(codes.Unknown, err.Error())
		}
		msg := resolved.(sdktypes.MsgRequest)
		deliverables[i] = &DeliverableDescriptor{
			Method:    name,
			ProtoName: proto.MessageName(msg),
		}
	}
	return &ListDeliverablesResponse{Deliverables: deliverables}, nil
}

func (r *reflectionServiceServer) ListQueryServices(ctx context.Context, request *ListQueriesRequest) (*ListQueriesResponse, error) {
	defer func() {
		r := recover()
		if r != nil {
			log.Printf("%#v", r)
		}
	}()
	svcs := r.infoProvider.ListServices()
	queries := make([]*QueryDescriptor, len(svcs))
	for i, q := range svcs {
		queries[i] = &QueryDescriptor{
			ServiceName: q.ServiceName,
			ProtoFile:   q.Metadata.(string),
		}
	}
	return &ListQueriesResponse{Queries: queries}, nil
}

func (r *reflectionServiceServer) ResolveProtoType(ctx context.Context, request *ResolveProtoTypeRequest) (*ResolveProtoTypeResponse, error) {
	typ, err := func() (typ reflect.Type, err error) {
		defer func() {
			r := recover()
			if r != nil {
				err = fmt.Errorf("type is not a recognized protobuf type: %s", request.Name)
			}
		}()

		typ = proto.MessageType(request.Name)
		return
	}()
	if err != nil {
		return nil, status.Errorf(codes.NotFound, err.Error())
	}

	if typ == nil {
		return nil, status.Errorf(codes.InvalidArgument, "resolution of type %s returned null", request.Name)
	}

	v := reflect.New(typ).Elem()

	ok := v.CanInterface()
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "type cannot be cast to interface: %s", request.Name)
	}

	vIf := v.Interface()

	pbMsg, ok := vIf.(proto.Message)
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "type %T does not implement proto message: %s", vIf, request.Name)
	}

	// get descriptor
	type descriptor interface {
		Descriptor() ([]byte, []int)
	}

	pbDescriptor, ok := pbMsg.(descriptor)
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "type does not implement the Descriptor interface: %s", request.Name)
	}
	rawDesc, pos := pbDescriptor.Descriptor()
	return &ResolveProtoTypeResponse{
		RawDescriptor: rawDesc,
		Indexes: func() []int64 {
			pos64 := make([]int64, len(pos))
			for i, p := range pos {
				pos64[i] = int64(p)
			}

			return pos64
		}(),
	}, nil
}

func (r *reflectionServiceServer) ResolveService(ctx context.Context, request *ResolveServiceRequest) (*ResolveServiceResponse, error) {
	rawDesc, err := func() (rawDesc []byte, err error) {
		defer func() {
			r := recover()
			if r != nil {
				err = fmt.Errorf("%#v", err)
			}

			rawDesc = proto.FileDescriptor(request.FileName)
			return
		}()

		return
	}()

	if err != nil {
		return nil, status.Errorf(codes.NotFound, "service from file was not found: %s", request.FileName)
	}

	return &ResolveServiceResponse{RawDescriptor: rawDesc}, nil
}
