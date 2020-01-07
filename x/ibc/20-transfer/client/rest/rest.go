package rest

import (
	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	channelexported "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

const (
	RestChannelID = "channel-id"
	RestPortID    = "port-id"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router) {
	registerQueryRoutes(cliCtx, r)
	registerTxRoutes(cliCtx, r)
}

// TransferTxReq defines the properties of a transfer tx request's body.
type TransferTxReq struct {
	BaseReq  rest.BaseReq   `json:"base_req" yaml:"base_req"`
	Amount   sdk.Coins      `json:"amount" yaml:"amount"`
	Receiver sdk.AccAddress `json:"receiver" yaml:"receiver"`
	Source   bool           `json:"source" yaml:"source"`
}

// RecvPacketReq defines the properties of a receive packet request's body.
type RecvPacketReq struct {
	BaseReq rest.BaseReq            `json:"base_req" yaml:"base_req"`
	Packet  channelexported.PacketI `json:"packet" yaml:"packet"`
	Proofs  []commitment.Proof      `json:"proofs" yaml:"proofs"`
	Height  uint64                  `json:"height" yaml:"height"`
}
