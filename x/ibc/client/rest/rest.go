package rest

import (
	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client"
	ibcclient "github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	tendermint "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint"
	localhost "github.com/cosmos/cosmos-sdk/x/ibc/09-localhost"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(clientCtx client.Context, r *mux.Router, queryRoute string) {
	ibcclient.RegisterRESTRoutes(clientCtx, r)
	tendermint.RegisterRESTRoutes(clientCtx, r)
	localhost.RegisterRESTRoutes(clientCtx, r)
	connection.RegisterRESTRoutes(clientCtx, r)
	channel.RegisterRESTRoutes(clientCtx, r)
}
