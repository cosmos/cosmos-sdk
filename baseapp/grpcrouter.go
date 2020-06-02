package baseapp

import (
	"fmt"

	gogogrpc "github.com/gogo/protobuf/grpc"
	abci "github.com/tendermint/tendermint/abci/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/encoding/proto"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var protoCodec = encoding.GetCodec(proto.Name)

// GRPCRouter routes ABCI Query requests to GRPC handlers
type GRPCRouter struct {
	routes map[string]GRPCQueryHandler
}

var _ gogogrpc.Server

// NewGRPCRouter creates a new GRPCRouter
func NewGRPCRouter() *GRPCRouter {
	return &GRPCRouter{
		routes: map[string]GRPCQueryHandler{},
	}
}

// GRPCQueryHandler defines a function type which handles ABCI Query requests
// using gRPC
type GRPCQueryHandler = func(ctx sdk.Context, req abci.RequestQuery) (abci.ResponseQuery, error)

// Route returns the GRPCQueryHandler for a given query route path or nil
// if not found
func (qrt *GRPCRouter) Route(path string) GRPCQueryHandler {
	handler, found := qrt.routes[path]
	if !found {
		return nil
	}
	return handler
}

// RegisterService implements the gRPC Server.RegisterService method. sd is a gRPC
// service description, handler is an object which implements that gRPC service
func (qrt *GRPCRouter) RegisterService(sd *grpc.ServiceDesc, handler interface{}) {
	// adds a top-level query handler based on the gRPC service name
	for _, method := range sd.Methods {
		fqName := fmt.Sprintf("/%s/%s", sd.ServiceName, method.MethodName)
		methodHandler := method.Handler

		qrt.routes[fqName] = func(ctx sdk.Context, req abci.RequestQuery) (abci.ResponseQuery, error) {
			// call the method handler from the service description with the handler object,
			// a wrapped sdk.Context with proto-unmarshaled data from the ABCI request data
			res, err := methodHandler(handler, sdk.WrapSDKContext(ctx), func(i interface{}) error {
				return protoCodec.Unmarshal(req.Data, i)
			}, nil)
			if err != nil {
				return abci.ResponseQuery{}, err
			}

			// proto marshal the result bytes
			resBytes, err := protoCodec.Marshal(res)
			if err != nil {
				return abci.ResponseQuery{}, err
			}

			// return the result bytes as the response value
			return abci.ResponseQuery{
				Height: req.Height,
				Value:  resBytes,
			}, nil
		}
	}
}
