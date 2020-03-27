package baseapp

import (
	"fmt"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/gogo/protobuf/proto"
	abci "github.com/tendermint/tendermint/abci/types"
	"google.golang.org/grpc"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type QueryRouter struct {
	routes map[string]sdk.Querier
}

var _ sdk.QueryRouter = NewQueryRouter()

// NewQueryRouter returns a reference to a new QueryRouter.
func NewQueryRouter() *QueryRouter {
	return &QueryRouter{
		routes: map[string]sdk.Querier{},
	}
}

// AddRoute adds a query path to the router with a given Querier. It will panic
// if a duplicate route is given. The route must be alphanumeric.
func (qrt *QueryRouter) AddRoute(path string, q sdk.Querier) sdk.QueryRouter {
	if !isAlphaNumeric(path) {
		panic("route expressions can only contain alphanumeric characters")
	}
	if qrt.routes[path] != nil {
		panic(fmt.Sprintf("route %s has already been initialized", path))
	}

	qrt.routes[path] = q
	return qrt
}

// Route returns the Querier for a given query route path.
func (qrt *QueryRouter) Route(path string) sdk.Querier {
	return qrt.routes[path]
}

func (qrt *QueryRouter) RegisterService(sd *grpc.ServiceDesc, handler interface{}) {
	// adds a top-level querier based on the GRPC service name
	qrt.AddRoute(sd.ServiceName,
		func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
			path0 := path[0]
			for _, md := range sd.Methods {
				// checks each GRPC service method to see if it matches the path
				if md.MethodName == path0 {
					res, err := md.Handler(handler, sdk.WrapSDKContext(ctx), func(i interface{}) error {
						// unmarshal a protobuf message
						protoMsg, ok := i.(proto.Message)
						if !ok {
							return sdkerrors.Wrapf(sdkerrors.ErrProtoUnmarshal, "can't proto unmarshal: %+v", i)
						}
						return proto.Unmarshal(req.Data, protoMsg)
					}, nil)
					if err != nil {
						return nil, err
					}
					protoMsg, ok := res.(proto.Message)
					if !ok {
						return nil, sdkerrors.Wrapf(sdkerrors.ErrProtoMarshal, "can't proto marshal: %+v", res)
					}
					return proto.Marshal(protoMsg)
				}
			}
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unknown query path: %s", path[0])
		})
}
