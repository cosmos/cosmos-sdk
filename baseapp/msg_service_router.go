package baseapp

import (
	"fmt"

	gogogrpc "github.com/gogo/protobuf/grpc"
	"google.golang.org/grpc"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MsgServiceRouter routes Msg Service fully-qualified service methods to their
// handler.
type MsgServiceRouter struct {
	interfaceRegistry codectypes.InterfaceRegistry
	routes            map[string]MsgServiceHandler
}

var _ gogogrpc.Server = &MsgServiceRouter{}

// NewMsgServiceRouter creates a new MsgServiceRouter.
func NewMsgServiceRouter() *MsgServiceRouter {
	return &MsgServiceRouter{
		routes: map[string]MsgServiceHandler{},
	}
}

// MsgServiceHandler defines a function type which handles Msg service message.
// It's similar to sdk.Handler, but with simplified version of `Msg`, without
// `Route()`, `Type()` and `GetSignBytes()`.
type MsgServiceHandler = func(ctx sdk.Context, msgRequest sdk.MsgRequest) (*sdk.Result, error)

// Route returns the MsgServiceHandler for a given query route path or nil
// if not found.
func (msr *MsgServiceRouter) Route(path string) MsgServiceHandler {
	handler, found := msr.routes[path]
	if !found {
		return nil
	}

	return handler
}

// RegisterService implements the gRPC Server.RegisterService method. sd is a gRPC
// service description, handler is an object which implements that gRPC service.
func (msr *MsgServiceRouter) RegisterService(sd *grpc.ServiceDesc, handler interface{}) {
	// adds a top-level query handler based on the gRPC service name
	for _, method := range sd.Methods {
		fqMethod := fmt.Sprintf("/%s/%s", sd.ServiceName, method.MethodName)
		methodHandler := method.Handler

		msr.routes[fqMethod] = func(ctx sdk.Context, msgRequest sdk.MsgRequest) (*sdk.Result, error) {
			// call the method handler from the service description with the handler object,
			// a wrapped sdk.Context with proto-unmarshaled data from the ABCI request data
			res, err := methodHandler(handler, sdk.WrapSDKContext(ctx), func(i interface{}) error {
				if msr.interfaceRegistry != nil {
					return codectypes.UnpackInterfaces(i, msr.interfaceRegistry)
				}
				return nil
			}, nil)
			if err != nil {
				return nil, err
			}

			// proto marshal the result bytes
			resBytes, err := protoCodec.Marshal(res)
			if err != nil {
				return nil, err
			}

			// return the result bytes as the response value
			return &sdk.Result{
				Data: resBytes,
			}, nil
		}
	}
}

// SetInterfaceRegistry sets the interface registry for the router.
func (msr *MsgServiceRouter) SetInterfaceRegistry(interfaceRegistry codectypes.InterfaceRegistry) {
	msr.interfaceRegistry = interfaceRegistry
}
