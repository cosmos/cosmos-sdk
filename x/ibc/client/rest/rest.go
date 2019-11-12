package rest

import (
	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"

	ics02 "github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	ics03 "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	ics04 "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
)

func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router) {
	ics02.RegisterRESTRoutes(cliCtx, r)
	ics03.RegisterRESTRoutes(cliCtx, r)
	ics04.RegisterRESTRoutes(cliCtx, r)
}
