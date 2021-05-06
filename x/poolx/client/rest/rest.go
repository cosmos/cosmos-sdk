package rest

import (
	"fmt"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client"
	clientrest "github.com/cosmos/cosmos-sdk/client/rest"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/poolx/types"
	"github.com/gorilla/mux"
)

const (
	MethodGet = "GET"
)

// RegisterRoutes registers poolx-related REST handlers to a router
func RegisterRoutes(clientCtx client.Context, r *mux.Router) {
}

func registerQueryRoutes(clientCtx client.Context, r *mux.Router) {
}

func registerTxHandlers(clientCtx client.Context, r *mux.Router) {
}
