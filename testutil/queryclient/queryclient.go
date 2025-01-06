package queryclient

import (
	"context"
	"fmt"

	abci "github.com/cometbft/cometbft/api/cometbft/abci/v1"
	gogogrpc "github.com/cosmos/gogoproto/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"

	"github.com/cosmos/cosmos-sdk/client/grpc/reflection"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
)

var (
	_ gogogrpc.ClientConn = &QueryHelper{}
	_ gogogrpc.Server     = &QueryHelper{}
)

// QueryHelper is a test utility for building a query client from a proto interface registry.
type QueryHelper struct {
	cdc    encoding.Codec
	routes map[string]GRPCQueryHandler
}

func NewQueryHelper(interfaceRegistry codectypes.InterfaceRegistry) *QueryHelper {
	// instantiate the codec
	cdc := codec.NewProtoCodec(interfaceRegistry).GRPCCodec()
	// Once we have an interface registry, we can register the interface
	// registry reflection gRPC service.

	qH := &QueryHelper{
		cdc:    cdc,
		routes: map[string]GRPCQueryHandler{},
	}

	reflection.RegisterReflectionServiceServer(qH, reflection.NewReflectionServiceServer(interfaceRegistry))

	return qH
}

// Invoke implements the grpc ClientConn.Invoke method
func (q *QueryHelper) Invoke(ctx context.Context, method string, args, reply interface{}, _ ...grpc.CallOption) error {
	querier := q.Route(method)
	if querier == nil {
		return fmt.Errorf("handler not found for %s", method)
	}
	reqBz, err := q.cdc.Marshal(args)
	if err != nil {
		return err
	}

	res, err := querier(ctx, &abci.QueryRequest{Data: reqBz})
	if err != nil {
		return err
	}

	err = q.cdc.Unmarshal(res.Value, reply)
	if err != nil {
		return err
	}

	return nil
}

// NewStream implements the grpc ClientConn.NewStream method
func (q *QueryHelper) NewStream(context.Context, *grpc.StreamDesc, string, ...grpc.CallOption) (grpc.ClientStream, error) {
	panic("not implemented")
}

// GRPCQueryHandler defines a function type which handles ABCI Query requests
// using gRPC
type GRPCQueryHandler = func(ctx context.Context, req *abci.QueryRequest) (*abci.QueryResponse, error)

// Route returns the GRPCQueryHandler for a given query route path or nil
// if not found
func (qrt *QueryHelper) Route(path string) GRPCQueryHandler {
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
func (qrt *QueryHelper) RegisterService(sd *grpc.ServiceDesc, handler interface{}) {
	// adds a top-level query handler based on the gRPC service name
	for _, method := range sd.Methods {
		qrt.registerABCIQueryHandler(sd, method, handler)
	}
}

func (qrt *QueryHelper) registerABCIQueryHandler(sd *grpc.ServiceDesc, method grpc.MethodDesc, handler interface{}) {
	fqName := fmt.Sprintf("/%s/%s", sd.ServiceName, method.MethodName)
	methodHandler := method.Handler

	_, found := qrt.routes[fqName]
	if found {
		panic(fmt.Sprintf("handler for %s already registered", fqName))
	}

	qrt.routes[fqName] = func(ctx context.Context, req *abci.QueryRequest) (*abci.QueryResponse, error) {
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
}
