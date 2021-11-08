package server

import (
	"bytes"
	"context"
	"fmt"
	"reflect"

	gogogrpc "github.com/gogo/protobuf/grpc"
	"github.com/gogo/protobuf/proto"
	"google.golang.org/grpc"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type handler struct {
	f            func(ctx context.Context, args, reply interface{}) error
	commitWrites bool
	moduleName   string
}

type router struct {
	handlers         map[string]handler
	providedServices map[reflect.Type]bool
	authzMiddleware  AuthorizationMiddleware
	msgServiceRouter *baseapp.MsgServiceRouter
}

type registrar struct {
	*router
	baseServer   gogogrpc.Server
	commitWrites bool
	moduleName   string
}

var _ gogogrpc.Server = registrar{}

func (r registrar) RegisterService(sd *grpc.ServiceDesc, ss interface{}) {
	r.providedServices[reflect.TypeOf(sd.HandlerType)] = true

	r.baseServer.RegisterService(sd, ss)

	for _, method := range sd.Methods {
		fqName := fmt.Sprintf("/%s/%s", sd.ServiceName, method.MethodName)
		methodHandler := method.Handler

		var requestTypeName string

		_, _ = methodHandler(nil, context.Background(), func(i interface{}) error {
			req, ok := i.(proto.Message)
			if !ok {
				// We panic here because there is no other alternative and the app cannot be initialized correctly
				// this should only happen if there is a problem with code generation in which case the app won't
				// work correctly anyway.
				panic(fmt.Errorf("can't register request type %T for service method %s", i, fqName))
			}
			requestTypeName = TypeURL(req)
			return nil
		}, noopInterceptor)

		f := func(ctx context.Context, args, reply interface{}) error {
			res, err := methodHandler(ss, ctx, func(i interface{}) error { return nil },
				func(ctx context.Context, _ interface{}, _ *grpc.UnaryServerInfo, unaryHandler grpc.UnaryHandler) (resp interface{}, err error) {
					return unaryHandler(ctx, args)
				})
			if err != nil {
				return err
			}

			resValue := reflect.ValueOf(res)
			if !resValue.IsZero() && reply != nil {
				reflect.ValueOf(reply).Elem().Set(resValue.Elem())
			}
			return nil
		}
		r.handlers[requestTypeName] = handler{
			f:            f,
			commitWrites: r.commitWrites,
			moduleName:   r.moduleName,
		}
	}
}

func (rtr *router) invoker(methodName string, writeCondition func(context.Context, string, sdk.Msg) error) (Invoker, error) {
	return func(ctx context.Context, request interface{}, response interface{}, opts ...interface{}) error {
		req, ok := request.(proto.Message)
		if !ok {
			return fmt.Errorf("expected proto.Message, got %T for service method %s", request, methodName)
		}

		typeURL := TypeURL(req)
		handler, found := rtr.handlers[typeURL]

		msg, isMsg := request.(sdk.Msg)
		if !found && !isMsg {
			return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, fmt.Sprintf("cannot find method named %s", methodName))
		}

		// cache wrap the multistore so that inter-module writes are atomic
		// see https://github.com/cosmos/cosmos-sdk/issues/8030
		regenCtx := types.UnwrapSDKContext(ctx)
		cacheMs := regenCtx.MultiStore().CacheMultiStore()
		ctx = sdk.WrapSDKContext(regenCtx.WithMultiStore(cacheMs))

		// msg handler
		if writeCondition != nil && (handler.commitWrites || isMsg) {
			err := msg.ValidateBasic()
			if err != nil {
				return err
			}

			err = writeCondition(ctx, methodName, msg)
			if err != nil {
				return err
			}

			// ADR-033 router
			if found {
				err = handler.f(ctx, request, response)
				if err != nil {
					return err
				}
			} else {
				// routing using baseapp.MsgServiceRouter
				sdkCtx := sdk.UnwrapSDKContext(ctx)
				handler := rtr.msgServiceRouter.HandlerByTypeURL(typeURL)
				if handler == nil {
					return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized message route: %s;", typeURL)
				}

				_, err = handler(sdkCtx, msg)
				if err != nil {
					return err
				}
			}

			// only commit writes if there is no error so that calls are atomic
			cacheMs.Write()
		} else {
			// query handler
			err := handler.f(ctx, request, response)
			if err != nil {
				return err
			}

			cacheMs.Write()
		}
		return nil

	}, nil
}

func (rtr *router) invokerFactory(moduleName string) InvokerFactory {
	return func(callInfo CallInfo) (Invoker, error) {
		moduleID := callInfo.Caller
		if moduleName != moduleID.ModuleName {
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized,
				fmt.Sprintf("expected a call from module %s, but module %s is calling", moduleName, moduleID.ModuleName))
		}

		moduleAddr := moduleID.Address()

		writeCondition := func(ctx context.Context, methodName string, msgReq sdk.Msg) error {
			signers := msgReq.GetSigners()
			if len(signers) != 1 {
				return fmt.Errorf("inter module Msg invocation requires a single expected signer (%s), but %s expects multiple signers (%+v),  ", moduleAddr, methodName, signers)
			}

			signer := signers[0]

			if bytes.Equal(moduleAddr, signer) {
				return nil
			}

			if rtr.authzMiddleware != nil && rtr.authzMiddleware(sdk.UnwrapSDKContext(ctx), methodName, msgReq, moduleAddr) {
				return nil
			}

			return sdkerrors.Wrap(sdkerrors.ErrUnauthorized,
				fmt.Sprintf("expected %s, got %s", signers[0], moduleAddr))
		}

		return rtr.invoker(callInfo.Method, writeCondition)
	}
}

func (rtr *router) testTxFactory(signers []sdk.AccAddress) InvokerFactory {
	signerMap := map[string]bool{}
	for _, signer := range signers {
		signerMap[signer.String()] = true
	}

	return func(callInfo CallInfo) (Invoker, error) {
		return rtr.invoker(callInfo.Method, func(_ context.Context, _ string, req sdk.Msg) error {
			for _, signer := range req.GetSigners() {
				if _, found := signerMap[signer.String()]; !found {
					return sdkerrors.ErrUnauthorized
				}
			}
			return nil
		})
	}
}

func (rtr *router) testQueryFactory() InvokerFactory {
	return func(callInfo CallInfo) (Invoker, error) {
		return rtr.invoker(callInfo.Method, nil)
	}
}

func TypeURL(req proto.Message) string {
	return "/" + proto.MessageName(req)
}

func noopInterceptor(_ context.Context, _ interface{}, _ *grpc.UnaryServerInfo, _ grpc.UnaryHandler) (interface{}, error) {
	return nil, nil
}
