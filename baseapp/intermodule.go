package baseapp

import (
	"bytes"
	"context"
	"fmt"
	"reflect"

	"cosmossdk.io/errors"
	"google.golang.org/grpc"

	"cosmossdk.io/core/appmodule"
	"github.com/cosmos/cosmos-sdk/baseapp/intermodule"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (app *BaseApp) SetInterModuleAuthorizer(authorizer intermodule.Authorizer) {
	app.interModuleAuthorizer = authorizer
}

func (app *BaseApp) InterModuleClient(moduleName string) appmodule.RootInterModuleClient {
	return intermodule.NewRootInterModuleClient(moduleName, func(callInfo intermodule.CallInfo) (appmodule.InterModuleInvoker, error) {
		return app.InterModuleInvoker(moduleName, callInfo)
	})
}

func (app *BaseApp) InterModuleInvoker(moduleName string, callInfo intermodule.CallInfo) (appmodule.InterModuleInvoker, error) {
	methodName := callInfo.Method
	msgHandler, found := app.msgServiceRouter.routes[methodName]

	if !found {
		return nil, errors.Wrap(sdkerrors.ErrInvalidRequest, fmt.Sprintf("cannot find method named %s", methodName))
	}

	return func(ctx context.Context, request interface{}, response interface{}, opts ...grpc.CallOption) error {
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
			return errors.Wrapf(sdkerrors.ErrInvalidRequest, "expected an sdk.Msg, got %t", request)
		}

		err := msg.ValidateBasic()
		if err != nil {
			return err
		}

		// first check the inter-module authorizer to see if this request is allowed under a special case
		if app.interModuleAuthorizer == nil || !app.interModuleAuthorizer(ctx, methodName, msg, moduleName) {
			// check the ADR-028 module address against the first signer
			moduleAddr := address.Module(moduleName, callInfo.DerivedPath)

			signers := msg.GetSigners()
			if len(signers) != 1 {
				return fmt.Errorf("inter module Msg invocation requires a single expected signer (%s), but %s expects multiple signers (%+v),  ", moduleAddr, methodName, signers)
			}

			signer := signers[0]

			if !bytes.Equal(moduleAddr, signer) {
				return errors.Wrap(sdkerrors.ErrUnauthorized,
					fmt.Sprintf("expected %s, got %s", signers[0], moduleAddr))
			}
		}

		res, err := msgHandler(sdkCtx, msg)
		if err != nil {
			return err
		}

		// reflection is needed here because of the way that gRPC clients code is
		// generated in that it allocates an instance of the response type, but
		// we already have one from the msg handler, so we just need to set the pointer
		resValue := reflect.ValueOf(res)
		if !resValue.IsZero() && response != nil {
			reflect.ValueOf(response).Elem().Set(resValue.Elem())
		}

		// only commit writes and events if there is no error so that calls are atomic
		cacheMs.Write()
		for _, event := range evtMgr.Events() {
			origEvtMr.EmitEvent(event)
		}

		return nil
	}, nil
}
