package rest

import (
	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
)

const (
	RestChannelID = "channel-id"
	RestPortID    = "port-id"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(clientCtx client.Context, r *mux.Router) {
	registerQueryRoutes(clientCtx, r)
	registerTxRoutes(clientCtx, r)
}

// TransferTxReq defines the properties of a transfer tx request's body.
type TransferTxReq struct {
	BaseReq    rest.BaseReq `json:"base_req" yaml:"base_req"`
	DestHeight uint64       `json:"dest_height" yaml:"dest_height"`
	Amount     sdk.Coins    `json:"amount" yaml:"amount"`
	Receiver   string       `json:"receiver" yaml:"receiver"`
}
