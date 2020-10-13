package baseapp

import (
	gocontext "context"
	"fmt"

	gogogrpc "github.com/gogo/protobuf/grpc"
	"google.golang.org/grpc"

	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MsgServiceTestHelper provides a helper for making grpc msg service
// rpc calls in unit tests. It implements both the grpc Server and ClientConn
// interfaces needed to register a msg service server and create a msg
// service client.
type MsgServiceTestHelper struct {
	*MsgServiceRouter
	ctx sdk.Context
}

var _ gogogrpc.Server = &MsgServiceTestHelper{}
var _ gogogrpc.ClientConn = &MsgServiceTestHelper{}

// NewMsgServerTestHelper creates a new MsgServiceTestHelper that wraps
// the provided sdk.Context
func NewMsgServerTestHelper(ctx sdk.Context, interfaceRegistry types.InterfaceRegistry) *MsgServiceTestHelper {
	qrt := NewMsgServiceRouter()
	qrt.SetInterfaceRegistry(interfaceRegistry)
	return &MsgServiceTestHelper{MsgServiceRouter: qrt, ctx: ctx}
}

// Invoke implements the grpc ClientConn.Invoke method
func (q *MsgServiceTestHelper) Invoke(_ gocontext.Context, method string, args, reply interface{}, _ ...grpc.CallOption) error {
	querier := q.Route(method)
	if querier == nil {
		return fmt.Errorf("handler not found for %s", method)
	}
	reqBz, err := protoCodec.Marshal(args)
	if err != nil {
		return err
	}

	res, err := querier(q.ctx, reqBz)
	if err != nil {
		return err
	}

	err = protoCodec.Unmarshal(res.Data, reply)
	if err != nil {
		return err
	}

	if q.interfaceRegistry != nil {
		return types.UnpackInterfaces(reply, q.interfaceRegistry)
	}

	return nil
}

// NewStream implements the grpc ClientConn.NewStream method
func (q *MsgServiceTestHelper) NewStream(gocontext.Context, *grpc.StreamDesc, string, ...grpc.CallOption) (grpc.ClientStream, error) {
	return nil, fmt.Errorf("not supported")
}
