package rest

import (
	"github.com/gorilla/mux"

	"github.com/KiraCore/cosmos-sdk/client"
	sdk "github.com/KiraCore/cosmos-sdk/types"
	"github.com/KiraCore/cosmos-sdk/types/rest"
)

const (
	restChannelID = "channel-id"
	restPortID    = "port-id"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(clientCtx client.Context, r *mux.Router) {
	registerTxRoutes(clientCtx, r)
}

// TransferTxReq defines the properties of a transfer tx request's body.
type TransferTxReq struct {
	BaseReq          rest.BaseReq `json:"base_req" yaml:"base_req"`
	Amount           sdk.Coins    `json:"amount" yaml:"amount"`
	Receiver         string       `json:"receiver" yaml:"receiver"`
	TimeoutHeight    uint64       `json:"timeout_height" yaml:"timeout_height"`
	TimeoutTimestamp uint64       `json:"timeout_timestamp" yaml:"timeout_timestamp"`
}
