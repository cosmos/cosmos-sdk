package baseapp

import (
	"fmt"

	abci "github.com/cometbft/cometbft/api/cometbft/abci/v1"
	gogogrpc "github.com/cosmos/gogoproto/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"

	"github.com/cosmos/cosmos-sdk/baseapp/internal/protoutils"
	"github.com/cosmos/cosmos-sdk/client/grpc/reflection"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type QueryRouter interface {
	RegisterService(sd *grpc.ServiceDesc, handler interface{})
	ResponseNameByRequestName(requestName string) string
	HandlerByRequestName(requestName string) GRPCQueryHandler
	Route(path string) GRPCQueryHandler
	SetInterfaceRegistry(interfaceRegistry codectypes.InterfaceRegistry)
}

// GRPCQueryRouter routes ABCI Query requests to GRPC handlers
type GRPCQueryRouter struct {
	// routes maps query handlers used in ABCIQuery.
	routes map[string]GRPCQueryHandler
	// routesByRequestName maps routes based on the request name
	routesByRequestName map[string]GRPCQueryHandler
	// responseByRequestName maps the request name to the response name.
	responseByRequestName map[string]string
	// binaryCodec is used to encode/decode binary protobuf messages.
	binaryCodec codec.BinaryCodec
	// cdc is the gRPC codec used by the router to correctly unmarshal messages.
	cdc encoding.Codec
	// serviceData contains the gRPC services and their handlers.
	serviceData []serviceData
}

// serviceData represents a gRPC service, along with its handler.
type serviceData struct {
	serviceDesc *grpc.ServiceDesc
	handler     interface{}
}

var (
	_ gogogrpc.Server = &GRPCQueryRouter{}
	_ QueryRouter     = &GRPCQueryRouter{}
)

// NewGRPCQueryRouter creates a new GRPCQueryRouter
func NewGRPCQueryRouter() *GRPCQueryRouter {
	return &GRPCQueryRouter{
		routes:                map[string]GRPCQueryHandler{},
		routesByRequestName:   map[string]GRPCQueryHandler{},
		responseByRequestName: map[string]string{},
		binaryCodec:           nil,
		cdc:                   nil,
		serviceData:           nil,
	}
}

// GRPCQueryHandler defines a function type which handles ABCI Query requests
// using gRPC
type GRPCQueryHandler = func(ctx sdk.Context, req *abci.QueryRequest) (*abci.QueryResponse, error)

// Route returns the GRPCQueryHandler for a given query route path or nil
// if not found
func (qrt *GRPCQueryRouter) Route(path string) GRPCQueryHandler {
	handler, found := qrt.routes[path]
	if !found {
		return nil
	}
	return handler
}

// RegisterService implements the gRPC Server.RegisterService method. sd is a gRPC
// service description, handler is an object which implements that gRPC service/
//
// This functions PANICS:
// - if a protobuf service is registered twice.
func (qrt *GRPCQueryRouter) RegisterService(sd *grpc.ServiceDesc, handler interface{}) {
	// adds a top-level query handler based on the gRPC service name
	for _, method := range sd.Methods {
		if err := qrt.registerABCIQueryHandler(sd, method, handler); err != nil {
			panic(err)
		}
	}

	qrt.serviceData = append(qrt.serviceData, serviceData{
		serviceDesc: sd,
		handler:     handler,
	})
}

func (qrt *GRPCQueryRouter) registerABCIQueryHandler(sd *grpc.ServiceDesc, method grpc.MethodDesc, handler interface{}) error {
	fqName := fmt.Sprintf("/%s/%s", sd.ServiceName, method.MethodName)
	methodHandler := method.Handler

	// Check that each service is only registered once. If a service is
	// registered more than once, then we should error. Since we can't
	// return an error (`Server.RegisterService` interface restriction) we
	// panic (at startup).
	_, found := qrt.routes[fqName]
	if found {
		return fmt.Errorf(
			"gRPC query service %s has already been registered. Please make sure to only register each service once. "+
				"This usually means that there are conflicting modules registering the same gRPC query service",
			fqName,
		)
	}

	qrt.routes[fqName] = func(ctx sdk.Context, req *abci.QueryRequest) (*abci.QueryResponse, error) {
		// call the method handler from the service description with the handler object,
		// a wrapped sdk.Context with proto-unmarshaled data from the ABCI request data
		res, err := methodHandler(handler, ctx, func(i interface{}) error {
			return qrt.cdc.Unmarshal(req.Data, i)
		}, nil)
		if err != nil {
			return nil, err
		}

		// proto marshal the result bytes
		var resBytes []byte
		resBytes, err = qrt.cdc.Marshal(res)
		if err != nil {
			return nil, err
		}

		// return the result bytes as the response value
		return &abci.QueryResponse{
			Height: req.Height,
			Value:  resBytes,
		}, nil
	}

	// register response name by request name
	reqName, respName, err := protoutils.RequestAndResponseFullNameFromMethodDesc(sd, method)
	if err != nil {
		return fmt.Errorf("grpc router unable to parse request and response name: %w", err)
	}
	qrt.responseByRequestName[string(reqName)] = string(respName)
	qrt.routesByRequestName[string(reqName)] = qrt.routes[fqName]
	return nil
}

func (qrt *GRPCQueryRouter) ResponseNameByRequestName(requestName string) string {
	return qrt.responseByRequestName[requestName]
}

func (qrt *GRPCQueryRouter) HandlerByRequestName(requestName string) GRPCQueryHandler {
	return qrt.routesByRequestName[requestName]
}

// SetInterfaceRegistry sets the interface registry for the router. This will
// also register the interface reflection gRPC service.
func (qrt *GRPCQueryRouter) SetInterfaceRegistry(interfaceRegistry codectypes.InterfaceRegistry) {
	qrt.binaryCodec = codec.NewProtoCodec(interfaceRegistry)
	// instantiate the codec
	qrt.cdc = codec.NewProtoCodec(interfaceRegistry).GRPCCodec()
	// Once we have an interface registry, we can register the interface
	// registry reflection gRPC service.
	reflection.RegisterReflectionServiceServer(qrt, reflection.NewReflectionServiceServer(interfaceRegistry))
}
