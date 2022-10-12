package baseapp

import (
	"bytes"
	"context"
	"fmt"

	"cosmossdk.io/errors"
	"google.golang.org/grpc"

	"github.com/cosmos/cosmos-sdk/baseapp/intermodule"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (app *BaseApp) InterModuleInvoker(moduleName string, callInfo intermodule.CallInfo) (intermodule.Invoker, error) {
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
