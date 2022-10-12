package baseapp

import (
	"bytes"
	"context"
	"fmt"

	"cosmossdk.io/errors"
	"google.golang.org/grpc"

	"cosmossdk.io/core/appmodule"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type InterModuleAuthorizer func(ctx context.Context, methodName string, req interface{}, callingModule string) bool

type invokerFactory func(callInfo callInfo) (invoker, error)

type invoker func(ctx context.Context, request, response interface{}, opts ...grpc.CallOption) error

type callInfo struct {
	method      string
	derivedPath []byte
}

type interModuleClient struct {
	module         string
	address        []byte
	path           []byte
	invokerFactory invokerFactory
}

func newInterModuleClient(module string, path []byte, invokerFactory invokerFactory) *interModuleClient {
	return &interModuleClient{
		module:         module,
		path:           path,
		invokerFactory: invokerFactory,
		address:        address.Module(module, path),
	}
}

func (c *interModuleClient) Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) error {
	invoker, err := c.invokerFactory(callInfo{
		method:      method,
		derivedPath: c.path,
	})
	if err != nil {
		return err
	}

	return invoker(ctx, args, reply, opts...)
}

func (c *interModuleClient) NewStream(context.Context, *grpc.StreamDesc, string, ...grpc.CallOption) (grpc.ClientStream, error) {
	return nil, fmt.Errorf("unsupported")
}

func (c *interModuleClient) Address() []byte {
	return c.address
}

var _ appmodule.InterModuleClient = &interModuleClient{}

type rootInterModuleClient struct {
	*interModuleClient
}

func newRootInterModuleClient(module string, invokerFactory invokerFactory) *rootInterModuleClient {
	return &rootInterModuleClient{newInterModuleClient(module, nil, invokerFactory)}
}

func (r *rootInterModuleClient) DerivedClient(key []byte) appmodule.InterModuleClient {
	return newInterModuleClient(r.module, key, r.invokerFactory)
}

var _ appmodule.RootInterModuleClient = &rootInterModuleClient{}

type interModuleRouter struct {
	msgRouter  *MsgServiceRouter
	authorizer InterModuleAuthorizer
}

func (rtr *interModuleRouter) invoker(methodName string, allowMsgCall func(context.Context, string, sdk.Msg) error) (invoker, error) {
	msgHandler, found := rtr.msgRouter.routes[methodName]

	if !found {
		return nil, errors.Wrap(sdkerrors.ErrInvalidRequest, fmt.Sprintf("cannot find method named %s", methodName))
	}

	return func(ctx context.Context, request interface{}, response interface{}, opts ...grpc.CallOption) error {
		// cache wrap the multistore so that inter-module writes are atomic
		// see https://github.com/cosmos/cosmos-sdk/issues/8030
		sdkCtx := sdk.UnwrapSDKContext(ctx)
		cacheMs := sdkCtx.MultiStore().CacheMultiStore()
		sdkCtx = sdkCtx.WithMultiStore(cacheMs)

		msg, ok := request.(sdk.Msg)
		if !ok {
			return errors.Wrapf(sdkerrors.ErrInvalidRequest, "expected an sdk.Msg, got %t", request)
		}

		err := msg.ValidateBasic()
		if err != nil {
			return err
		}

		err = allowMsgCall(ctx, methodName, msg)
		if err != nil {
			return err
		}

		// TODO
		_, err = msgHandler(sdkCtx, msg)
		if err != nil {
			return err
		}

		// only commit writes if there is no error so that calls are atomic
		cacheMs.Write()

		return nil
	}, nil
}

func (rtr *interModuleRouter) invokerFactory(moduleName string) invokerFactory {
	return func(callInfo callInfo) (invoker, error) {
		//if moduleName != callingModule {
		//	return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized,
		//		fmt.Sprintf("expected a call from module %s, but module %s is calling", moduleName, callingModule.ModuleName))
		//}

		moduleAddr := address.Module(moduleName, callInfo.derivedPath)

		writeCondition := func(ctx context.Context, methodName string, msgReq sdk.Msg) error {
			signers := msgReq.GetSigners()
			if len(signers) != 1 {
				return fmt.Errorf("inter module Msg invocation requires a single expected signer (%s), but %s expects multiple signers (%+v),  ", moduleAddr, methodName, signers)
			}

			signer := signers[0]

			if bytes.Equal(moduleAddr, signer) {
				return nil
			}

			if rtr.authorizer != nil && rtr.authorizer(ctx, methodName, msgReq, moduleName) {
				return nil
			}

			return sdkerrors.Wrap(sdkerrors.ErrUnauthorized,
				fmt.Sprintf("expected %s, got %s", signers[0], moduleAddr))
		}

		return rtr.invoker(callInfo.method, writeCondition)
	}
}
