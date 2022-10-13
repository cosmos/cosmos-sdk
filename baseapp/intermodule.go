package baseapp

import (
	"bytes"
	"context"
	"fmt"
	"reflect"

	"cosmossdk.io/errors"
	"github.com/cosmos/gogoproto/proto"
	"google.golang.org/grpc"
	protov2 "google.golang.org/protobuf/proto"

	"cosmossdk.io/core/intermodule"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type InterModuleAuthorizer func(ctx context.Context, methodName string, req interface{}, callingModule string) bool

func (app *BaseApp) SetInterModuleAuthorizer(authorizer InterModuleAuthorizer) {
	app.interModuleAuthorizer = authorizer
}

func (app *BaseApp) InterModuleClient(moduleName string) intermodule.Client {
	return newInterModuleClient(moduleName, nil, app)
}

func newInterModuleClient(module string, paths [][]byte, bApp *BaseApp) *interModuleClient {
	var addr []byte
	n := len(paths)
	if n == 0 {
		addr = address.Module(module, nil)
	} else {
		addr = address.Module(module, paths[0])
		for i := 1; i < n; i++ {
			addr = address.Derive(addr, paths[i])
		}
	}
	return &interModuleClient{
		module:  module,
		paths:   paths,
		bApp:    bApp,
		address: addr,
	}
}

type interModuleClient struct {
	module  string
	address []byte
	paths   [][]byte
	bApp    *BaseApp
}

func (c *interModuleClient) InvokerByMethod(method string) (intermodule.Invoker, error) {
	msgHandler, found := c.bApp.msgServiceRouter.routes[method]

	if !found {
		return nil, errors.Wrap(sdkerrors.ErrInvalidRequest, fmt.Sprintf("cannot find method named %s", method))
	}

	return func(ctx context.Context, request any, opts ...grpc.CallOption) (any, error) {
		// cache wrap the multistore so that inter-module writes are atomic
		// see https://github.com/cosmos/cosmos-sdk/issues/8030
		sdkCtx := sdk.UnwrapSDKContext(ctx)
		cacheMs := sdkCtx.MultiStore().CacheMultiStore()
		sdkCtx = sdkCtx.WithMultiStore(cacheMs)

		origEvtMr := sdkCtx.EventManager()
		evtMgr := sdk.NewEventManager()
		sdkCtx = sdkCtx.WithEventManager(evtMgr)

		msg, ok := request.(sdk.Msg)
		if !ok {
			return nil, errors.Wrapf(sdkerrors.ErrInvalidRequest, "expected an sdk.Msg, got %t", request)
		}

		err := msg.ValidateBasic()
		if err != nil {
			return nil, err
		}

		// first check the inter-module authorizer to see if this request is allowed under a special case
		if c.bApp.interModuleAuthorizer == nil || !c.bApp.interModuleAuthorizer(ctx, method, msg, c.module) {
			signers := msg.GetSigners()
			if len(signers) != 1 {
				return nil, fmt.Errorf("inter module Msg invocation requires a single expected signer (%s), but %s expects multiple signers (%+v),  ", c.address, method, signers)
			}

			signer := signers[0]

			if !bytes.Equal(c.address, signer) {
				return nil, errors.Wrap(sdkerrors.ErrUnauthorized,
					fmt.Sprintf("expected %s, got %s", signers[0], c.address))
			}
		}

		res, err := msgHandler(sdkCtx, msg)
		if err != nil {
			return nil, err
		}

		// only commit writes and events if there is no error so that calls are atomic
		cacheMs.Write()
		for _, event := range evtMgr.Events() {
			origEvtMr.EmitEvent(event)
		}

		return res, nil
	}, nil
}

func (c *interModuleClient) InvokerByRequest(request interface{}) (intermodule.Invoker, error) {
	var method string
	if msg, ok := request.(protov2.Message); ok {
		method = string(msg.ProtoReflect().Descriptor().FullName())
	} else if msg, ok := request.(proto.Message); ok {
		method = proto.MessageName(msg)
	} else {
		return nil, fmt.Errorf("expected a proto message, got %T", request)
	}

	return c.InvokerByMethod(method)
}

func (c *interModuleClient) Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) error {
	var invoker intermodule.Invoker
	invoker, err := c.InvokerByRequest(method)
	if err != nil {
		return err
	}

	res, err := invoker(ctx, args, opts...)

	// reflection is needed here because of the way that gRPC clients code is
	// generated in that it allocates an instance of the response type, but
	// we already have one from the msg handler, so we just need to set the pointer
	resValue := reflect.ValueOf(res)
	if !resValue.IsZero() && reply != nil {
		reflect.ValueOf(reply).Elem().Set(resValue.Elem())
	}

	return nil
}

func (c *interModuleClient) NewStream(context.Context, *grpc.StreamDesc, string, ...grpc.CallOption) (grpc.ClientStream, error) {
	return nil, fmt.Errorf("unsupported")
}

func (c *interModuleClient) Address() []byte {
	return c.address
}

func (c *interModuleClient) DerivedClient(key []byte) intermodule.Client {
	n := len(c.paths)
	paths := make([][]byte, n+1)
	copy(paths, c.paths)
	paths[n] = key
	return newInterModuleClient(c.module, paths, c.bApp)
}

var _ intermodule.Client = &interModuleClient{}
