package rest

import (
	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	client "github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router, queryRoute string) {
	client.RegisterRESTRoutes(cliCtx, r, queryRoute)
	connection.RegisterRESTRoutes(cliCtx, r, queryRoute)
	channel.RegisterRESTRoutes(cliCtx, r, queryRoute)
}
