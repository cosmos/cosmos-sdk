package queryclient

import (
	"context"
	"fmt"

	gogogrpc "github.com/cosmos/gogoproto/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"
)

var (
	_ gogogrpc.ClientConn = &QueryHelper{}
	_ gogogrpc.Server     = &QueryHelper{}
)

// GRPCQueryHandler defines a function type which handles mocked ABCI Query requests
// using gRPC
type GRPCQueryHandler = func(ctx context.Context, req *QueryRequest) (*QueryResponse, error)

// QueryRequest is a light mock of cometbft abci.QueryRequest.
type QueryRequest struct {
	Data   []byte
	Height int64
}

// QueryResponse is a light mock of cometbft abci.QueryResponse.
type QueryResponse struct {
	Value  []byte
	Height int64
}

// QueryHelper is a test utility for building a query client from a proto interface registry.
type QueryHelper struct {
	cdc    encoding.Codec
	routes map[string]GRPCQueryHandler
}

func NewQueryHelper(cdc encoding.Codec) *QueryHelper {
	qh := &QueryHelper{
		cdc:    cdc,
		routes: map[string]GRPCQueryHandler{},
	}

	return qh
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

	res, err := querier(ctx, &QueryRequest{Data: reqBz})
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

// Route returns the GRPCQueryHandler for a given query route path or nil
// if not found
func (q *QueryHelper) Route(path string) GRPCQueryHandler {
	handler, found := q.routes[path]
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
func (q *QueryHelper) RegisterService(sd *grpc.ServiceDesc, handler interface{}) {
	// adds a top-level query handler based on the gRPC service name
	for _, method := range sd.Methods {
		q.registerABCIQueryHandler(sd, method, handler)
	}
}

func (q *QueryHelper) registerABCIQueryHandler(sd *grpc.ServiceDesc, method grpc.MethodDesc, handler interface{}) {
	fqName := fmt.Sprintf("/%s/%s", sd.ServiceName, method.MethodName)
	methodHandler := method.Handler

	_, found := q.routes[fqName]
	if found {
		panic(fmt.Sprintf("handler for %s already registered", fqName))
	}

	q.routes[fqName] = func(ctx context.Context, req *QueryRequest) (*QueryResponse, error) {
		// call the method handler from the service description with the handler object,
		// a wrapped sdk.Context with proto-unmarshaled data from the ABCI request data
		res, err := methodHandler(handler, ctx, func(i interface{}) error {
			return q.cdc.Unmarshal(req.Data, i)
		}, nil)
		if err != nil {
			return nil, err
		}

		// proto marshal the result bytes
		var resBytes []byte
		resBytes, err = q.cdc.Marshal(res)
		if err != nil {
			return nil, err
		}

		// return the result bytes as the response value
		return &QueryResponse{
			Height: req.Height,
			Value:  resBytes,
		}, nil
	}
}
