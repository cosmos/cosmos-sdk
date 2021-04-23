package baseapp

import (
	"context"
	"reflect"

	gogogrpc "github.com/gogo/protobuf/grpc"
	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcrecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/cosmos/cosmos-sdk/client"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx"
)

// GRPCQueryRouter returns the GRPCQueryRouter of a BaseApp.
func (app *BaseApp) GRPCQueryRouter() *GRPCQueryRouter { return app.grpcQueryRouter }

// RegisterGRPCServer registers gRPC services directly with the gRPC server.
func (app *BaseApp) RegisterGRPCServer(clientCtx client.Context, server gogogrpc.Server) {
	// Define an interceptor for all gRPC queries: this interceptor will route
	// the query through the `clientCtx`, which itself queries Tendermint.
	interceptor := func(grpcCtx context.Context, req interface{}, info *grpc.UnaryServerInfo, _ grpc.UnaryHandler) (interface{}, error) {
		// Two things can happen here:
		// 1. either we're broadcasting a Tx, in which case we call Tendermint's broadcast endpoint directly,
		// 2. or we are querying for state, in which case we call ABCI's Query.

		// Case 1. Broadcasting a Tx.
		if reqProto, ok := req.(*tx.BroadcastTxRequest); ok {
			if !ok {
				return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "expected %T, got %T", (*tx.BroadcastTxRequest)(nil), req)
			}

			return client.TxServiceBroadcast(grpcCtx, clientCtx, reqProto)
		}

		// Case 2. Querying state.
		inMd, _ := metadata.FromIncomingContext(grpcCtx)
		abciRes, outMd, err := client.RunGRPCQuery(clientCtx, grpcCtx, info.FullMethod, req, inMd)
		if err != nil {
			return nil, err
		}

		// We need to know the return type of the grpc method for
		// unmarshalling abciRes.Value.
		//
		// When we call each method handler for the first time, we save its
		// return type in the `returnTypes` map (see the method handler in
		// `grpcrouter.go`). By this time, the method handler has already run
		// at least once (in the RunGRPCQuery call), so we're sure the
		// returnType maps is populated for this method. We're retrieving it
		// for decoding.
		returnType, err := app.GRPCQueryRouter().returnTypeOf(info.FullMethod)
		if err != nil {
			return nil, err
		}

		// returnType is a pointer to a struct. Here, we're creating res which
		// is a new pointer to the underlying struct.
		res := reflect.New(returnType.Elem()).Interface()

		err = protoCodec.Unmarshal(abciRes.Value, res)
		if err != nil {
			return nil, err
		}

		// Send the metadata header back. The metadata currently includes:
		// - block height.
		err = grpc.SendHeader(grpcCtx, outMd)
		if err != nil {
			return nil, err
		}

		return res, nil
	}

	// Loop through all services and methods, add the interceptor, and register
	// the service.
	for _, data := range app.GRPCQueryRouter().serviceData {
		desc := data.serviceDesc
		newMethods := make([]grpc.MethodDesc, len(desc.Methods))

		for i, method := range desc.Methods {
			methodHandler := method.Handler
			newMethods[i] = grpc.MethodDesc{
				MethodName: method.MethodName,
				Handler: func(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
					return methodHandler(srv, ctx, dec, grpcmiddleware.ChainUnaryServer(
						grpcrecovery.UnaryServerInterceptor(),
						interceptor,
					))
				},
			}
		}

		newDesc := &grpc.ServiceDesc{
			ServiceName: desc.ServiceName,
			HandlerType: desc.HandlerType,
			Methods:     newMethods,
			Streams:     desc.Streams,
			Metadata:    desc.Metadata,
		}

		server.RegisterService(newDesc, data.handler)
	}
}
