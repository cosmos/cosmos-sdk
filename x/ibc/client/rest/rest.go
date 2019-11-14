package rest

import (
	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	ics02 "github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	ics03 "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	ics04 "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	ics20 "github.com/cosmos/cosmos-sdk/x/ibc/20-transfer"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router, queryRoute string) {
	ics02.RegisterRESTRoutes(cliCtx, r, queryRoute)
	ics03.RegisterRESTRoutes(cliCtx, r, queryRoute)
	ics04.RegisterRESTRoutes(cliCtx, r, queryRoute)
	ics20.RegisterRESTRoutes(cliCtx, r)
}
