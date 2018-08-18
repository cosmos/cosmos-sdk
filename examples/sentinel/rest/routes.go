package rest

import (
	"github.com/cosmos/cosmos-sdk/client/context"
	sentinel "github.com/cosmos/cosmos-sdk/examples/sentinel"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/gorilla/mux"
)

func RegisterRoutes(ctx context.CoreContext, r *mux.Router, cdc *wire.Codec, keeper sentinel.Keeper) {

	ServiceRoutes(ctx, r, cdc)
	QueryRoutes(ctx, r, cdc, keeper)

}
