package baseapp

import (
	gocontext "context"
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"
	gogogrpc "github.com/cosmos/gogoproto/grpc"
	"google.golang.org/grpc"

	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// QueryServiceTestHelper provides a helper for making gRPC query service
// RPC calls in unit tests. It implements both the gRPC Server and ClientConn
// interfaces needed to register a query service server and create a query
// service client for testing purposes.
type QueryServiceTestHelper struct {
	*GRPCQueryRouter // Embedded router for handling query routing
	Ctx sdk.Context  // SDK context for query execution
}

// Ensure QueryServiceTestHelper implements the required gRPC interfaces
var (
	_ gogogrpc.Server     = &QueryServiceTestHelper{}
	_ gogogrpc.ClientConn = &QueryServiceTestHelper{}
)

// NewQueryServerTestHelper creates a new QueryServiceTestHelper that wraps
// the provided SDK context and initializes a gRPC query router.
func NewQueryServerTestHelper(ctx sdk.Context, interfaceRegistry types.InterfaceRegistry) *QueryServiceTestHelper {
	// Create a new gRPC query router and configure it with the interface registry
	qrt := NewGRPCQueryRouter()
	qrt.SetInterfaceRegistry(interfaceRegistry)
	return &QueryServiceTestHelper{GRPCQueryRouter: qrt, Ctx: ctx}
}

// Invoke implements the gRPC ClientConn.Invoke method.
// It routes the query to the appropriate handler and marshals/unmarshals the data.
func (q *QueryServiceTestHelper) Invoke(_ gocontext.Context, method string, args, reply any, _ ...grpc.CallOption) error {
	// Find the appropriate query handler for the method
	querier := q.Route(method)
	if querier == nil {
		return fmt.Errorf("handler not found for %s", method)
	}
	// Marshal the request arguments
	reqBz, err := q.cdc.Marshal(args)
	if err != nil {
		return err
	}

	// Execute the query with the marshaled request
	res, err := querier(q.Ctx, &abci.RequestQuery{Data: reqBz})
	if err != nil {
		return err
	}

	// Unmarshal the response into the reply object
	err = q.cdc.Unmarshal(res.Value, reply)
	if err != nil {
		return err
	}

	return nil
}

// NewStream implements the gRPC ClientConn.NewStream method.
// Streaming is not supported in this test helper.
func (q *QueryServiceTestHelper) NewStream(gocontext.Context, *grpc.StreamDesc, string, ...grpc.CallOption) (grpc.ClientStream, error) {
	return nil, fmt.Errorf("not supported")
}
