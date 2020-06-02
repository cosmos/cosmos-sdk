package baseapp

import (
	gocontext "context"
	"fmt"

	gogogrpc "github.com/gogo/protobuf/grpc"
	abci "github.com/tendermint/tendermint/abci/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/encoding/proto"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var protoCodec = encoding.GetCodec(proto.Name)

// GRPCRouter routes ABCI Query requests to GRPC handlers
type GRPCRouter struct {
	routes map[string]GRPCQueryHandler
}

var _ gogogrpc.Server

// NewGRPCRouter creates a new GRPCRouter
func NewGRPCRouter() *GRPCRouter {
	return &GRPCRouter{
		routes: map[string]GRPCQueryHandler{},
	}
}

// GRPCQueryHandler defines a function type which handles ABCI Query requests
// using gRPC
type GRPCQueryHandler = func(ctx sdk.Context, req abci.RequestQuery) (abci.ResponseQuery, error)

// Route returns the Querier for a given query route path.
func (qrt *GRPCRouter) Route(path string) GRPCQueryHandler {
	handler, found := qrt.routes[path]
	if !found {
		return nil
	}
	return handler
}

// RegisterService implements the grpc Server.RegisterService method
func (qrt *GRPCRouter) RegisterService(sd *grpc.ServiceDesc, handler interface{}) {
	// adds a top-level querier based on the GRPC service name
	for _, method := range sd.Methods {
		fqName := fmt.Sprintf("/%s/%s", sd.ServiceName, method.MethodName)
		methodHandler := method.Handler

		qrt.routes[fqName] = func(ctx sdk.Context, req abci.RequestQuery) (abci.ResponseQuery, error) {
			res, err := methodHandler(handler, sdk.WrapSDKContext(ctx), func(i interface{}) error {
				return protoCodec.Unmarshal(req.Data, i)
			}, nil)
			if err != nil {
				return abci.ResponseQuery{}, err
			}

			resBytes, err := protoCodec.Marshal(res)
			if err != nil {
				return abci.ResponseQuery{}, err
			}

			return abci.ResponseQuery{
				Height: req.Height,
				Value:  resBytes,
			}, nil
		}
	}
}

// QueryServiceTestHelper provides a helper for making grpc query service
// rpc calls in unit tests. It implements both the grpc Server and ClientConn
// interfaces needed to register a query service server and create a query
// service client.
type QueryServiceTestHelper struct {
	*GRPCRouter
	ctx sdk.Context
}

// NewQueryServerTestHelper creates a new QueryServiceTestHelper that wraps
// the provided sdk.Context
func NewQueryServerTestHelper(ctx sdk.Context) *QueryServiceTestHelper {
	return &QueryServiceTestHelper{GRPCRouter: NewGRPCRouter(), ctx: ctx}
}

// Invoke implements the grpc ClientConn.Invoke method
func (q *QueryServiceTestHelper) Invoke(_ gocontext.Context, method string, args, reply interface{}, _ ...grpc.CallOption) error {
	querier := q.Route(method)
	if querier == nil {
		return fmt.Errorf("handler not found for %s", method)
	}
	reqBz, err := protoCodec.Marshal(args)
	if err != nil {
		return err
	}
	res, err := querier(q.ctx, abci.RequestQuery{Data: reqBz})

	if err != nil {
		return err
	}
	return protoCodec.Unmarshal(res.Value, reply)
}

// NewStream implements the grpc ClientConn.NewStream method
func (q *QueryServiceTestHelper) NewStream(gocontext.Context, *grpc.StreamDesc, string, ...grpc.CallOption) (grpc.ClientStream, error) {
	return nil, fmt.Errorf("not supported")
}

var _ gogogrpc.Server = &QueryServiceTestHelper{}
var _ gogogrpc.ClientConn = &QueryServiceTestHelper{}
