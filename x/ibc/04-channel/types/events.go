package types

import (
	"fmt"

	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

// IBC channel events
const (
	AttributeKeyConnectionID       = "connection_id"
	AttributeKeyPortID             = "port_id"
	AttributeKeyChannelID          = "channel_id"
	AttributeCounterpartyPortID    = "counterparty_port_id"
	AttributeCounterpartyChannelID = "counterparty_channel_id"
)

// IBC channel events vars
var (
	EventTypeChannelOpenInit     = MsgChannelOpenInit{}.Type()
	EventTypeChannelOpenTry      = MsgChannelOpenTry{}.Type()
	EventTypeChannelOpenAck      = MsgChannelOpenAck{}.Type()
	EventTypeChannelOpenConfirm  = MsgChannelOpenConfirm{}.Type()
	EventTypeChannelCloseInit    = MsgChannelCloseInit{}.Type()
	EventTypeChannelCloseConfirm = MsgChannelCloseConfirm{}.Type()

	AttributeValueCategory = fmt.Sprintf("%s_%s", ibctypes.ModuleName, SubModuleName)
)
