package types

import (
	"fmt"

	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

// IBC channel events
const (
	EventTypeSendPacket          = "send_packet"
	EventTypeChannelOpenInit     = "channel_open_init"
	EventTypeChannelOpenTry      = "channel_open_try"
	EventTypeChannelOpenAck      = "channel_open_ack"
	EventTypeChannelOpenConfirm  = "channel_open_confirm"
	EventTypeChannelCloseInit    = "channel_close_init"
	EventTypeChannelCloseConfirm = "channel_close_confirm"

	AttributeKeySenderPort   = "sender_port"
	AttributeKeyReceiverPort = "receiver_port"
	AttributeKeyChannelID    = "channel_id"
	AttributeKeySequence     = "sequence"
)

// IBC channel events vars
var (
	AttributeValueCategory = fmt.Sprintf("%s_%s", ibctypes.ModuleName, SubModuleName)
)
