package baseapp

import (
	gocontext "context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec/types"

	gogogrpc "github.com/gogo/protobuf/grpc"
	abci "github.com/tendermint/tendermint/abci/types"
	"google.golang.org/grpc"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// QueryServiceTestHelper provides a helper for making grpc query service
// rpc calls in unit tests. It implements both the grpc Server and ClientConn
// interfaces needed to register a query service server and create a query
// service client.
type QueryServiceTestHelper struct {
	*GRPCQueryRouter
	ctx sdk.Context
}

// NewQueryServerTestHelper creates a new QueryServiceTestHelper that wraps
// the provided sdk.Context
func NewQueryServerTestHelper(ctx sdk.Context, anyUnpacker types.AnyUnpacker) *QueryServiceTestHelper {
	qrt := NewGRPCQueryRouter()
	qrt.SetAnyUnpacker(anyUnpacker)
	return &QueryServiceTestHelper{GRPCQueryRouter: qrt, ctx: ctx}
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

	err = protoCodec.Unmarshal(res.Value, reply)
	if err != nil {
		return err
	}

	if q.anyUnpacker != nil {
		return types.UnpackInterfaces(reply, q.anyUnpacker)
	}

	return nil
}

// NewStream implements the grpc ClientConn.NewStream method
func (q *QueryServiceTestHelper) NewStream(gocontext.Context, *grpc.StreamDesc, string, ...grpc.CallOption) (grpc.ClientStream, error) {
	return nil, fmt.Errorf("not supported")
}

var _ gogogrpc.Server = &QueryServiceTestHelper{}
var _ gogogrpc.ClientConn = &QueryServiceTestHelper{}
