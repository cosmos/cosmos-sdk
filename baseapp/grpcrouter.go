package baseapp

import (
	"context"
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"
	gogogrpc "github.com/cosmos/gogoproto/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"
	"google.golang.org/protobuf/runtime/protoiface"

	"github.com/cosmos/cosmos-sdk/baseapp/internal/protocompat"
	"github.com/cosmos/cosmos-sdk/client/grpc/reflection"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GRPCQueryRouter routes ABCI Query requests to gRPC handlers.
// It provides a bridge between the ABCI query interface and gRPC services,
// supporting both traditional ABCI queries and hybrid message handling.
type GRPCQueryRouter struct {
	// routes maps query paths to their corresponding ABCI query handlers
	routes map[string]GRPCQueryHandler
	// hybridHandlers maps request message names to hybrid handlers that can
	// seamlessly handle both gogoproto and protov2 messages
	hybridHandlers map[string][]func(ctx context.Context, req, resp protoiface.MessageV1) error
	// binaryCodec is used to encode/decode binary protobuf messages
	binaryCodec codec.BinaryCodec
	// cdc is the gRPC codec used by the router to correctly marshal/unmarshal messages
	cdc encoding.Codec
	// serviceData contains the registered gRPC services and their handler implementations
	serviceData []serviceData
}

// serviceData represents a gRPC service descriptor along with its handler implementation.
type serviceData struct {
	serviceDesc *grpc.ServiceDesc
	handler     any
}

// Ensure GRPCQueryRouter implements the gogogrpc.Server interface
var _ gogogrpc.Server = &GRPCQueryRouter{}

// NewGRPCQueryRouter creates a new GRPCQueryRouter instance
// with initialized route and handler maps.
func NewGRPCQueryRouter() *GRPCQueryRouter {
	return &GRPCQueryRouter{
		routes:         map[string]GRPCQueryHandler{},
		hybridHandlers: map[string][]func(ctx context.Context, req, resp protoiface.MessageV1) error{},
	}
}

// GRPCQueryHandler defines a function type that handles ABCI Query requests
// using gRPC. It takes an SDK context and ABCI request, returning an ABCI response.
type GRPCQueryHandler = func(ctx sdk.Context, req *abci.RequestQuery) (*abci.ResponseQuery, error)

// Route returns the GRPCQueryHandler for a given query route path.
// Returns nil if no handler is found for the specified path.
func (qrt *GRPCQueryRouter) Route(path string) GRPCQueryHandler {
	handler, found := qrt.routes[path]
	if !found {
		return nil
	}
	return handler
}

// RegisterService implements the gRPC Server.RegisterService method.
// sd is a gRPC service descriptor, handler is an object that implements that gRPC service.
//
// This function PANICS if:
// - a protobuf service is registered twice (duplicate registration)
func (qrt *GRPCQueryRouter) RegisterService(sd *grpc.ServiceDesc, handler any) {
	// Register handlers for each method in the service
	for _, method := range sd.Methods {
		// Register the ABCI query handler for this method
		err := qrt.registerABCIQueryHandler(sd, method, handler)
		if err != nil {
			panic(err)
		}
		// Register the hybrid handler for this method
		err = qrt.registerHybridHandler(sd, method, handler)
		if err != nil {
			panic(err)
		}
	}

	// Store the service data for later reference
	qrt.serviceData = append(qrt.serviceData, serviceData{
		serviceDesc: sd,
		handler:     handler,
	})
}

func (qrt *GRPCQueryRouter) registerABCIQueryHandler(sd *grpc.ServiceDesc, method grpc.MethodDesc, handler any) error {
	fqName := fmt.Sprintf("/%s/%s", sd.ServiceName, method.MethodName)
	methodHandler := method.Handler

	// Check that each service is only registered once. If a service is
	// registered more than once, we should error. Since we can't
	// return an error (due to `Server.RegisterService` interface restriction),
	// we panic at startup to prevent runtime issues.
	_, found := qrt.routes[fqName]
	if found {
		return fmt.Errorf(
			"gRPC query service %s has already been registered. Please make sure to only register each service once. "+
				"This usually means that there are conflicting modules registering the same gRPC query service",
			fqName,
		)
	}

	// Create the ABCI query handler for this method
	qrt.routes[fqName] = func(ctx sdk.Context, req *abci.RequestQuery) (*abci.ResponseQuery, error) {
		// Call the method handler with the handler object and SDK context
		// The decoder function unmarshals the ABCI request data into the expected message type
		res, err := methodHandler(handler, ctx, func(i any) error {
			return qrt.cdc.Unmarshal(req.Data, i)
		}, nil)
		if err != nil {
			return nil, err
		}

		// Marshal the result into protobuf bytes
		var resBytes []byte
		resBytes, err = qrt.cdc.Marshal(res)
		if err != nil {
			return nil, err
		}

		// Return the result bytes as the ABCI response value
		return &abci.ResponseQuery{
			Height: req.Height,
			Value:  resBytes,
		}, nil
	}
	return nil
}

// HybridHandlerByRequestName returns the hybrid handlers for a given request message name.
// These handlers can process both gogoproto and protov2 messages seamlessly.
func (qrt *GRPCQueryRouter) HybridHandlerByRequestName(name string) []func(ctx context.Context, req, resp protoiface.MessageV1) error {
	return qrt.hybridHandlers[name]
}

// registerHybridHandler registers a hybrid handler that can handle both gogoproto and protov2 messages.
func (qrt *GRPCQueryRouter) registerHybridHandler(sd *grpc.ServiceDesc, method grpc.MethodDesc, handler any) error {
	// Extract the full message name from the method descriptor
	inputName, err := protocompat.RequestFullNameFromMethodDesc(sd, method)
	if err != nil {
		return err
	}
	// Create a hybrid handler that can process both message types
	methodHandler, err := protocompat.MakeHybridHandler(qrt.binaryCodec, sd, method, handler)
	if err != nil {
		return err
	}
	// Store the hybrid handler for this request message name
	qrt.hybridHandlers[string(inputName)] = append(qrt.hybridHandlers[string(inputName)], methodHandler)
	return nil
}

// SetInterfaceRegistry sets the interface registry for the router and initializes
// the codec. This will also register the interface reflection gRPC service.
func (qrt *GRPCQueryRouter) SetInterfaceRegistry(interfaceRegistry codectypes.InterfaceRegistry) {
	// Initialize the codec with the interface registry
	qrt.cdc = codec.NewProtoCodec(interfaceRegistry).GRPCCodec()
	qrt.binaryCodec = codec.NewProtoCodec(interfaceRegistry)
	// Register the interface reflection gRPC service for runtime introspection
	reflection.RegisterReflectionServiceServer(qrt, reflection.NewReflectionServiceServer(interfaceRegistry))
}
