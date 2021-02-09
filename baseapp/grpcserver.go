package baseapp

import (
	"context"
	"strconv"

	gogogrpc "github.com/gogo/protobuf/grpc"
	"github.com/gogo/protobuf/proto"
	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcrecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	abci "github.com/tendermint/tendermint/abci/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/client"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
)

// GRPCQueryRouter returns the GRPCQueryRouter of a BaseApp.
func (app *BaseApp) GRPCQueryRouter() *GRPCQueryRouter { return app.grpcQueryRouter }

// RegisterGRPCServer registers gRPC services directly with the gRPC server.
func (app *BaseApp) RegisterGRPCServer(clientCtx client.Context, server gogogrpc.Server) {
	// Define an interceptor for all gRPC queries: this interceptor will create
	// a new sdk.Context, and pass it into the query handler.
	interceptor := func(grpcCtx context.Context, req interface{}, info *grpc.UnaryServerInfo, _ grpc.UnaryHandler) (resp interface{}, err error) {
		// If there's some metadata in the context, retrieve it.
		md, ok := metadata.FromIncomingContext(grpcCtx)
		if !ok {
			return nil, status.Error(codes.Internal, "unable to retrieve metadata")
		}

		// Get height header from the request context, if present.
		var height int64
		if heightHeaders := md.Get(grpctypes.GRPCBlockHeightHeader); len(heightHeaders) > 0 {
			height, err = strconv.ParseInt(heightHeaders[0], 10, 64)
			if err != nil {
				return nil, sdkerrors.Wrapf(
					sdkerrors.ErrInvalidRequest,
					"baseapp.RegisterGRPCServer: invalid height header %q: %v", grpctypes.GRPCBlockHeightHeader, err)
			}
			if err := checkNegativeHeight(height); err != nil {
				return nil, err
			}
		}

		reqProto, ok := req.(proto.Message)
		if !ok {
			return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "expected %T, got %T", (proto.Message)(nil), req)
		}
		reqBz, err := proto.Marshal(reqProto)
		if err != nil {
			return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "cannot proto marshal: %v", err)
		}

		abciReq := abci.RequestQuery{
			Path:   info.FullMethod,
			Data:   reqBz,
			Height: height,
			Prove:  false, // Enable here: https://github.com/cosmos/cosmos-sdk/issues/7036
		}
		abciRes, err := clientCtx.QueryABCI(abciReq)
		if err != nil {
			return nil, err
		}

		err = protoCodec.Unmarshal(abciRes.Value, resp)
		if err != nil {
			return nil, err
		}

		// TODO Add metadata height in response?

		return
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
