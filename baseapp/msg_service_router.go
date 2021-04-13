package baseapp

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	gogogrpc "github.com/gogo/protobuf/grpc"
	"github.com/gogo/protobuf/proto"
	"google.golang.org/grpc"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/pkg/protohelpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// MsgServiceRouter routes fully-qualified Msg service methods to their handler.
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
type MsgServiceHandler = func(ctx sdk.Context, req sdk.Msg) (*sdk.Result, error)

// Handler returns the MsgServiceHandler for a given query route path or nil
// if not found.
func (msr *MsgServiceRouter) Handler(methodName string) MsgServiceHandler {
	return msr.routes[methodName]
}

// RegisterService implements the gRPC Server.RegisterService method. sd is a gRPC
// service description, handler is an object which implements that gRPC service.
//
// This function PANICs:
// - if interface registry is nil
// - or if a service is being registered twice.
func (msr *MsgServiceRouter) RegisterService(sd *grpc.ServiceDesc, handler interface{}) {
	// register the request types in the interface registry
	err := msr.registerInputs(sd)
	if err != nil {
		panic(err)
	}
	// map requests to handlers
	err = msr.registerMethods(sd, handler)
	if err != nil {
		panic(err)
	}
}

// registerMethods will register the handlers for each input in the given grpc.ServiceDesc
func (msr *MsgServiceRouter) registerMethods(sd *grpc.ServiceDesc, handler interface{}) error {
	// Adds a top-level query handler based on the gRPC service name.
	for _, method := range sd.Methods {
		fqMethod := fmt.Sprintf("/%s/%s", sd.ServiceName, method.MethodName)
		methodHandler := method.Handler

		// Check that each service is only registered once. If a service is
		// registered more than once, then we should error. Since we can't
		// return an error (`Server.RegisterService` interface restriction) we
		// panic (at startup).
		_, found := msr.routes[fqMethod]
		if found {
			return fmt.Errorf(
				"msg service %s has already been registered. Please make sure to only register each service once. "+
					"This usually means that there are conflicting modules registering the same msg service",
				fqMethod,
			)
		}

		msr.routes[fqMethod] = func(ctx sdk.Context, req sdk.Msg) (*sdk.Result, error) {
			ctx = ctx.WithEventManager(sdk.NewEventManager())
			interceptor := func(goCtx context.Context, _ interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
				goCtx = context.WithValue(goCtx, sdk.SdkContextKey, ctx)
				return handler(goCtx, req)
			}
			// Call the method handler from the service description with the handler object.
			// We don't do any decoding here because the decoding was already done.
			res, err := methodHandler(handler, sdk.WrapSDKContext(ctx), noopDecoder, interceptor)
			if err != nil {
				return nil, err
			}

			resMsg, ok := res.(proto.Message)
			if !ok {
				return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "Expecting proto.Message, got %T", resMsg)
			}

			return sdk.WrapServiceResult(ctx, resMsg, err)
		}
	}
	return nil
}

// registerInputs registers the requests of a MsgServer service descriptor
// in the interface registry to allow BaseApp to get the correct requests
// from the Deliver/CheckTx implementation.
func (msr *MsgServiceRouter) registerInputs(gsd *grpc.ServiceDesc) error {
	if msr.interfaceRegistry == nil {
		return fmt.Errorf("nil interface registry")
	}
	// ok, so we've got a service desc, it does not allow us to know the
	// inputs, but we can parse the raw descriptor and get it from there.
	sd, err := protohelpers.ServiceDescriptorFromGRPCServiceDesc(gsd)
	if err != nil {
		return err
	}
	// we iterate methods, get the inputs, fetch the concrete types and register them
	for i := 0; i < sd.Methods().Len(); i++ {
		md := sd.Methods().Get(i)
		if md.IsStreamingServer() || md.IsStreamingClient() {
			return fmt.Errorf("streaming RPC are not supported, found in %s", md.FullName())
		}
		// get the concrete type
		typ := proto.MessageType((string)(md.Input().FullName()))
		if typ == nil {
			return fmt.Errorf("concrete type for %s not found", md.Input().FullName())
		}
		v := reflect.New(typ).Elem()
		msg, ok := v.Interface().(sdk.Msg)
		if !ok {
			return fmt.Errorf("type %T is not an sdk.Msg", v.Interface())
		}
		msr.interfaceRegistry.RegisterImplementations((*sdk.Msg)(nil), msg)
		// TODO(fdymylja): remove this once ServiceMsgs are no longer supported.
		svcMsgTypeURL := fmt.Sprintf("/%s/%s", sd.FullName(), md.Name())
		msr.interfaceRegistry.RegisterCustomTypeURL((*sdk.Msg)(nil), svcMsgTypeURL, msg)
	}
	return nil
}

func (msr *MsgServiceRouter) SetInterfaceRegistry(ir codectypes.InterfaceRegistry) {
	msr.interfaceRegistry = ir
}

// IsServiceMsg checks if a type URL corresponds to a service method name,
// i.e. /cosmos.bank.Msg/Send vs /cosmos.bank.MsgSend
func IsServiceMsg(typeURL string) bool {
	return strings.Count(typeURL, "/") >= 2
}

func noopDecoder(_ interface{}) error { return nil }
