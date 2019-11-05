package types

import (
	"fmt"

	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

// IBC channel events
const (
	AttributeKeySenderPort   = "sender_port"
	AttributeKeyReceiverPort = "receiver_port"
	AttributeKeyChannelID    = "channel_id"
	AttributeKeySequence     = "sequence"
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
