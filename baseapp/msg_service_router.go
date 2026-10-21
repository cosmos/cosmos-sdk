package baseapp

import (
	"context"
	"fmt"

	gogogrpc "github.com/cosmos/gogoproto/grpc"
	"github.com/cosmos/gogoproto/proto"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/runtime/protoiface"

	errorsmod "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/baseapp/internal/protocompat"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// MessageRouter ADR 031 request type routing
// https://github.com/cosmos/cosmos-sdk/blob/main/docs/architecture/adr-031-msg-service.md
type MessageRouter interface {
	Handler(msg sdk.Msg) MsgServiceHandler
	HandlerByTypeURL(typeURL string) MsgServiceHandler
}

// MsgServiceRouter routes fully-qualified Msg service methods to their handler.
type MsgServiceRouter struct {
	interfaceRegistry codectypes.InterfaceRegistry
	routes            map[string]MsgServiceHandler
	hybridHandlers    map[string]func(ctx context.Context, req, resp protoiface.MessageV1) error
	circuitBreaker    sdk.CircuitBreaker
}

var _ gogogrpc.Server = &MsgServiceRouter{}

// NewMsgServiceRouter creates a new MsgServiceRouter.
func NewMsgServiceRouter() *MsgServiceRouter {
	return &MsgServiceRouter{
		routes:         map[string]MsgServiceHandler{},
		hybridHandlers: map[string]func(ctx context.Context, req, resp protoiface.MessageV1) error{},
	}
}

func (msr *MsgServiceRouter) SetCircuit(cb sdk.CircuitBreaker) {
	msr.circuitBreaker = cb
}

// MsgServiceHandler defines a function type which handles Msg service message.
type MsgServiceHandler = func(ctx sdk.Context, req sdk.Msg) (*sdk.Result, error)

// Handler returns the MsgServiceHandler for a given msg or nil if not found.
func (msr *MsgServiceRouter) Handler(msg sdk.Msg) MsgServiceHandler {
	return msr.routes[sdk.MsgTypeURL(msg)]
}

// HandlerByTypeURL returns the MsgServiceHandler for a given query route path or nil
// if not found.
func (msr *MsgServiceRouter) HandlerByTypeURL(typeURL string) MsgServiceHandler {
	return msr.routes[typeURL]
}

// RegisterService implements the gRPC Server.RegisterService method. sd is a gRPC
// service description, handler is an object which implements that gRPC service.
//
// This function PANICs:
//   - if it is called before the service `Msg`s have been registered using
//     RegisterInterfaces,
//   - or if a service is being registered twice.
func (msr *MsgServiceRouter) RegisterService(sd *grpc.ServiceDesc, handler any) {
	// Adds a top-level query handler based on the gRPC service name.
	for _, method := range sd.Methods {
		err := msr.registerMsgServiceHandler(sd, method, handler)
		if err != nil {
			panic(err)
		}
		err = msr.registerHybridHandler(sd, method, handler)
		if err != nil {
			panic(err)
		}
	}
}

func (msr *MsgServiceRouter) HybridHandlerByMsgName(msgName string) func(ctx context.Context, req, resp protoiface.MessageV1) error {
	return msr.hybridHandlers[msgName]
}

func (msr *MsgServiceRouter) registerHybridHandler(sd *grpc.ServiceDesc, method grpc.MethodDesc, handler any) error {
	inputName, err := protocompat.RequestFullNameFromMethodDesc(sd, method)
	if err != nil {
		return err
	}
	cdc := codec.NewProtoCodec(msr.interfaceRegistry)
	hybridHandler, err := protocompat.MakeHybridHandler(cdc, sd, method, handler)
	if err != nil {
		return err
	}
	// if circuit breaker is not nil, then we decorate the hybrid handler with the circuit breaker
	if msr.circuitBreaker == nil {
		msr.hybridHandlers[string(inputName)] = hybridHandler
		return nil
	}
	// decorate the hybrid handler with the circuit breaker. It checks the
	// msg type the handler runs, not the type of req: a request of another
	// type whose bytes decode as inputName still runs inputName's handler.
	typeURL := "/" + string(inputName)
	circuitBreakerHybridHandler := func(ctx context.Context, req, resp protoiface.MessageV1) error {
		if err := msgservice.CheckCircuitBreaker(ctx, msr.circuitBreaker, typeURL); err != nil {
			return err
		}
		return hybridHandler(markChecked(ctx, typeURL), req, resp)
	}
	msr.hybridHandlers[string(inputName)] = circuitBreakerHybridHandler
	return nil
}

func (msr *MsgServiceRouter) registerMsgServiceHandler(sd *grpc.ServiceDesc, method grpc.MethodDesc, handler any) error {
	fqMethod := fmt.Sprintf("/%s/%s", sd.ServiceName, method.MethodName)
	methodHandler := method.Handler

	var requestTypeName string

	// NOTE: This is how we pull the concrete request type for each handler for registering in the InterfaceRegistry.
	// This approach is maybe a bit hacky, but less hacky than reflecting on the handler object itself.
	// We use a no-op interceptor to avoid actually calling into the handler itself.
	_, _ = methodHandler(nil, context.Background(), func(i any) error {
		msg, ok := i.(sdk.Msg)
		if !ok {
			// We panic here because there is no other alternative and the app cannot be initialized correctly
			// this should only happen if there is a problem with code generation in which case the app won't
			// work correctly anyway.
			panic(fmt.Errorf("unable to register service method %s: %T does not implement sdk.Msg", fqMethod, i))
		}

		requestTypeName = sdk.MsgTypeURL(msg)
		return nil
	}, noopInterceptor)

	// Check that the service Msg fully-qualified method name has already
	// been registered (via RegisterInterfaces). If the user registers a
	// service without registering according service Msg type, there might be
	// some unexpected behavior down the road. Since we can't return an error
	// (`Server.RegisterService` interface restriction) we panic (at startup).
	reqType, err := msr.interfaceRegistry.Resolve(requestTypeName)
	if err != nil || reqType == nil {
		return fmt.Errorf(
			"type_url %s has not been registered yet. "+
				"Before calling RegisterService, you must register all interfaces by calling the `RegisterInterfaces` "+
				"method on module.BasicManager. Each module should call `msgservice.RegisterMsgServiceDesc` inside its "+
				"`RegisterInterfaces` method with the `_Msg_serviceDesc` generated by proto-gen",
			requestTypeName,
		)
	}

	// Check that each service is only registered once. If a service is
	// registered more than once, then we should error. Since we can't
	// return an error (`Server.RegisterService` interface restriction) we
	// panic (at startup).
	_, found := msr.routes[requestTypeName]
	if found {
		return fmt.Errorf(
			"msg service %s has already been registered. Please make sure to only register each service once. "+
				"This usually means that there are conflicting modules registering the same msg service",
			fqMethod,
		)
	}

	msr.routes[requestTypeName] = func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		if m, ok := msg.(sdk.HasValidateBasic); ok {
			if err := m.ValidateBasic(); err != nil {
				return nil, err
			}
		}

		// Mark the msg the router checked, so the msg server's own
		// msgservice.CheckMsgAllowed call skips a second lookup. The mark
		// always replaces the caller's, so a mark set for an outer msg never
		// covers a msg dispatched from inside its handler.
		checkedURL := ""
		if msr.circuitBreaker != nil {
			checkedURL = sdk.MsgTypeURL(msg)
			if err := msgservice.CheckCircuitBreaker(ctx, msr.circuitBreaker, checkedURL); err != nil {
				return nil, err
			}
		}
		ctx = ctx.WithCheckedMsgTypeURL(checkedURL)

		interceptor := func(goCtx context.Context, _ any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
			goCtx = context.WithValue(goCtx, sdk.SdkContextKey, ctx)
			return handler(goCtx, msg)
		}

		// Call the method handler from the service description with the handler object.
		// We don't do any decoding here because the decoding was already done.
		res, err := methodHandler(handler, ctx, noopDecoder, interceptor)
		if err != nil {
			return nil, err
		}

		resMsg, ok := res.(proto.Message)
		if !ok {
			return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidType, "Expecting proto.Message, got %T", resMsg)
		}

		return sdk.WrapServiceResult(ctx, resMsg, err)
	}
	return nil
}

// SetInterfaceRegistry sets the interface registry for the router.
func (msr *MsgServiceRouter) SetInterfaceRegistry(interfaceRegistry codectypes.InterfaceRegistry) {
	msr.interfaceRegistry = interfaceRegistry
}

// markChecked returns ctx with its sdk.Context marked as dispatching typeURL,
// or ctx unchanged when it has no sdk.Context.
func markChecked(ctx context.Context, typeURL string) context.Context {
	sdkCtx, ok := ctx.Value(sdk.SdkContextKey).(sdk.Context)
	if !ok {
		return ctx
	}
	sdkCtx = sdkCtx.WithCheckedMsgTypeURL(typeURL)
	if _, isSDKCtx := ctx.(sdk.Context); isSDKCtx {
		return sdkCtx
	}
	return context.WithValue(ctx, sdk.SdkContextKey, sdkCtx)
}

func noopDecoder(_ any) error { return nil }
func noopInterceptor(_ context.Context, _ any, _ *grpc.UnaryServerInfo, _ grpc.UnaryHandler) (any, error) {
	return nil, nil
}
