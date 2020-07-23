package grpc

import (
	"context"
	"reflect"

	gogogrpc "github.com/gogo/protobuf/grpc"
	"github.com/gogo/protobuf/proto"
	"github.com/tendermint/tendermint/abci/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/client"
)

type proxyServer struct {
	server      *grpc.Server
	interceptor grpc.UnaryServerInterceptor
}

// NewProxyServer creates a new proxy server with an interceptor.
func NewProxyServer(server *grpc.Server, interceptor grpc.UnaryServerInterceptor) gogogrpc.Server {
	return &proxyServer{server: server, interceptor: interceptor}
}

func (proxy *proxyServer) RegisterService(desc *grpc.ServiceDesc, handler interface{}) {
	newMethods := make([]grpc.MethodDesc, len(desc.Methods))

	for i, method := range desc.Methods {
		method := method // Fix scopelint: Using the variable on range scope `method` in function literal
		newMethods[i] = grpc.MethodDesc{
			MethodName: method.MethodName,
			Handler: func(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
				return method.Handler(srv, ctx, dec, proxy.interceptor)
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

	proxy.server.RegisterService(newDesc, handler)
}

// ABCIQueryProxyInterceptor implements a query interceptor that intercepts
// queries and sends them to the ABCI.
func ABCIQueryProxyInterceptor(clientCtx client.Context) grpc.UnaryServerInterceptor {
	resTypeCache := map[string]reflect.Type{}

	return func(_ context.Context, req interface{}, info *grpc.UnaryServerInfo, _ grpc.UnaryHandler) (resp interface{}, err error) {
		msg, ok := req.(proto.Message)
		if !ok {
			return nil, status.Errorf(codes.Internal, "unable to proto marshal")
		}

		msgBz, err := proto.Marshal(msg)
		if err != nil {
			return nil, err
		}

		abciReq := types.RequestQuery{
			Data: msgBz,
			Path: info.FullMethod,
		}

		resType, ok := resTypeCache[info.FullMethod]
		if !ok {
			// TODO figure out response type and cache
			// NOTE: this is a bit hacky but...
		}

		abciRes, err := clientCtx.QueryABCI(abciReq)
		if err != nil {
			return nil, err
		}

		if abciRes.Code != 0 {
			return nil, status.Errorf(codes.Internal, abciRes.Log)
		}

		respMsg, ok := reflect.New(resType).Interface().(proto.Message)
		if !ok {
			return nil, status.Errorf(codes.Internal, "unable to proto marshal")
		}

		err = proto.Unmarshal(abciRes.Value, respMsg)
		if err != nil {
			return nil, err
		}

		return resp, nil
	}
}
