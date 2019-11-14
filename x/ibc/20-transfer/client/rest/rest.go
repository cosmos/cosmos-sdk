package rest

import (
	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
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

/*
// RecvPacketReq defines the properties of a receive packet request's body.
type RecvPacketReq struct {
	BaseReq rest.BaseReq            `json:"base_req" yaml:"base_req"`
	Packet  channelexported.PacketI `json:"packet" yaml:"packet"`
	Proofs  []commitment.Proof      `json:"proofs" yaml:"proofs"`
	Height  uint64                  `json:"height" yaml:"height"`
}
*/

// RecvPacketReq defines the properties of a receive packet request's body.
type RecvPacketReq struct {
	BaseReq         rest.BaseReq `json:"base_req" yaml:"base_req"`
	SourceNode      string       `json:"source_node" yaml:"source_node"`
	Sequence        string       `json:"sequence" yaml:"sequence"`
	Timeout         string       `json:"timeout" yaml:"timeout"`
	SourceChainID   string       `json:"source_chain_id" yaml:"source_chain_id"`
	SourcePortID    string       `json:"source_port_id" yaml:"source_port_id"`
	SourceChannelID string       `json:"source_channel_id" yaml:"source_channel_id"`
	ClientID        string       `json:"client_id" yaml:"client_id"`
}
