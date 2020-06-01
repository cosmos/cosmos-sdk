package rest

import (
	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client"
	client2 "github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	tendermint "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint"
	localhost "github.com/cosmos/cosmos-sdk/x/ibc/09-localhost"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(clientCtx client.Context, r *mux.Router, queryRoute string) {
	client2.RegisterRESTRoutes(clientCtx, r, queryRoute)
	tendermint.RegisterRESTRoutes(clientCtx, r, queryRoute)
	localhost.RegisterRESTRoutes(clientCtx, r, queryRoute)
	connection.RegisterRESTRoutes(clientCtx, r, queryRoute)
	channel.RegisterRESTRoutes(clientCtx, r, queryRoute)
}
