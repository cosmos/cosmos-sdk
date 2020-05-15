package baseapp

import (
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"
	"google.golang.org/grpc"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

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
	if !sdk.IsAlphaNumeric(path) {
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

// RegisterService implements the grpc Server.RegisterService method
func (qrt *QueryRouter) RegisterService(sd *grpc.ServiceDesc, handler interface{}) {
	// adds a top-level querier based on the GRPC service name
	qrt.routes[sd.ServiceName] =
		func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
			path0 := path[0]
			for _, md := range sd.Methods {
				// checks each GRPC service method to see if it matches the path
				if md.MethodName != path0 {
					continue
				}
				res, err := md.Handler(handler, sdk.WrapSDKContext(ctx), func(i interface{}) error {
					return protoCodec.Unmarshal(req.Data, i)
				}, nil)
				if err != nil {
					return nil, err
				}
				return protoCodec.Marshal(res)
			}
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unknown query path: %s", path[0])
		}
}
