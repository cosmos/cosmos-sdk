package baseapp

import (
	"context"
	"fmt"
	"strings"

	gogogrpc "github.com/gogo/protobuf/grpc"
	"github.com/gogo/protobuf/proto"
	"google.golang.org/grpc"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
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
func (msr *MsgServiceRouter) registerInputs(sd *grpc.ServiceDesc) error {
	if msr.interfaceRegistry == nil {
		return fmt.Errorf("nil interface registry")
	}
	// Adds a top-level type_url based on the Msg service name.
	for _, method := range sd.Methods {
		fqMethod := fmt.Sprintf("/%s/%s", sd.ServiceName, method.MethodName)
		methodHandler := method.Handler

		// NOTE: This is how we pull the concrete request type for each handler for registering in the InterfaceRegistry.
		// This approach is maybe a bit hacky, but less hacky than reflecting on the handler object itself.
		// We use a no-op interceptor to avoid actually calling into the handler itself.
		_, _ = methodHandler(nil, context.Background(), func(i interface{}) error {
			msg, ok := i.(proto.Message)
			if !ok {
				// We panic here because there is no other alternative and the app cannot be initialized correctly
				// this should only happen if there is a problem with code generation in which case the app won't
				// work correctly anyway.
				panic(fmt.Errorf("can't register request type %T for service method %s", i, fqMethod))
			}

			msr.interfaceRegistry.RegisterCustomTypeURL((*sdk.Msg)(nil), fqMethod, msg)
			return nil
		}, noopInterceptor)

	}

	return nil
}

func (msr *MsgServiceRouter) SetInterfaceRegistry(ir codectypes.InterfaceRegistry) {
	msr.interfaceRegistry = ir
}

// gRPC NOOP interceptor
func noopInterceptor(_ context.Context, _ interface{}, _ *grpc.UnaryServerInfo, _ grpc.UnaryHandler) (interface{}, error) {
	return nil, nil
}

// IsServiceMsg checks if a type URL corresponds to a service method name,
// i.e. /cosmos.bank.Msg/Send vs /cosmos.bank.MsgSend
func IsServiceMsg(typeURL string) bool {
	return strings.Count(typeURL, "/") >= 2
}

func noopDecoder(_ interface{}) error { return nil }
