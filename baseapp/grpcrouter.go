package baseapp

import (
	"fmt"
	"reflect"

	gogogrpc "github.com/gogo/protobuf/grpc"
	abci "github.com/tendermint/tendermint/abci/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/encoding/proto"

	"github.com/cosmos/cosmos-sdk/client/grpc/reflection"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var protoCodec = encoding.GetCodec(proto.Name)

// GRPCQueryRouter routes ABCI Query requests to GRPC handlers
type GRPCQueryRouter struct {
	routes map[string]GRPCQueryHandler
	// returnTypes is a map of FQ method name => its return type. It is used
	// for cache purposes: the first time a method handler is run, we save its
	// return type in this map. Then, on subsequent method handler calls, we
	// decode the ABCI response bytes using the cached return type.
	returnTypes       map[string]reflect.Type
	interfaceRegistry codectypes.InterfaceRegistry
	serviceData       []serviceData
}

// serviceData represents a gRPC service, along with its handler.
type serviceData struct {
	serviceDesc *grpc.ServiceDesc
	handler     interface{}
}

var _ gogogrpc.Server = &GRPCQueryRouter{}

// NewGRPCQueryRouter creates a new GRPCQueryRouter
func NewGRPCQueryRouter() *GRPCQueryRouter {
	return &GRPCQueryRouter{
		returnTypes: map[string]reflect.Type{},
		routes:      map[string]GRPCQueryHandler{},
	}
}

// GRPCQueryHandler defines a function type which handles ABCI Query requests
// using gRPC
type GRPCQueryHandler = func(ctx sdk.Context, req abci.RequestQuery) (abci.ResponseQuery, error)

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
		fqName := fmt.Sprintf("/%s/%s", sd.ServiceName, method.MethodName)
		methodHandler := method.Handler

		// Check that each service is only registered once. If a service is
		// registered more than once, then we should error. Since we can't
		// return an error (`Server.RegisterService` interface restriction) we
		// panic (at startup).
		_, found := qrt.routes[fqName]
		if found {
			panic(
				fmt.Errorf(
					"gRPC query service %s has already been registered. Please make sure to only register each service once. "+
						"This usually means that there are conflicting modules registering the same gRPC query service",
					fqName,
				),
			)
		}

		qrt.routes[fqName] = func(ctx sdk.Context, req abci.RequestQuery) (abci.ResponseQuery, error) {
			// call the method handler from the service description with the handler object,
			// a wrapped sdk.Context with proto-unmarshaled data from the ABCI request data
			res, err := methodHandler(handler, sdk.WrapSDKContext(ctx), func(i interface{}) error {
				err := protoCodec.Unmarshal(req.Data, i)
				if err != nil {
					return err
				}
				if qrt.interfaceRegistry != nil {
					return codectypes.UnpackInterfaces(i, qrt.interfaceRegistry)
				}

				return nil
			}, nil)

			// If it's the first time we call this handler, then we save
			// the return type of the handler in the `returnTypes` map.
			// The return type will be used for decoding subsequent requests.
			if _, found := qrt.returnTypes[fqName]; !found {
				qrt.returnTypes[fqName] = reflect.TypeOf(res)
			}

			if err != nil {
				return abci.ResponseQuery{}, err
			}

			// proto marshal the result bytes
			resBytes, err := protoCodec.Marshal(res)
			if err != nil {
				return abci.ResponseQuery{}, err
			}

			// return the result bytes as the response value
			return abci.ResponseQuery{
				Height: req.Height,
				Value:  resBytes,
			}, nil
		}
	}

	qrt.serviceData = append(qrt.serviceData, serviceData{
		serviceDesc: sd,
		handler:     handler,
	})
}

// SetInterfaceRegistry sets the interface registry for the router. This will
// also register the interface reflection gRPC service.
func (qrt *GRPCQueryRouter) SetInterfaceRegistry(interfaceRegistry codectypes.InterfaceRegistry) {
	qrt.interfaceRegistry = interfaceRegistry

	// Once we have an interface registry, we can register the interface
	// registry reflection gRPC service.
	reflection.RegisterReflectionServiceServer(
		qrt,
		reflection.NewReflectionServiceServer(interfaceRegistry),
	)
}

// returnTypeOf returns the return type of a gRPC method handler. With the way the
// `returnTypes` cache map is set up, the return type of a method handler is
// guaranteed to be found if it's retrieved **after** the method handler ran at
// least once. If not, then a logic error is return.
func (qrt *GRPCQueryRouter) returnTypeOf(method string) (reflect.Type, error) {
	returnType, found := qrt.returnTypes[method]
	if !found {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrLogic, "cannot find %s return type", method)
	}

	return returnType, nil
}
