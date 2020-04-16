package types

import (
	"fmt"

	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

// IBC transfer events
const (
	EventTypeTimeout      = "timeout"
	EventTypePacket       = "fungible_token_packet"
	EventTypeChannelClose = "channel_closed"

	AttributeKeyReceiver       = "receiver"
	AttributeKeyValue          = "value"
	AttributeKeyRefundReceiver = "refund_receiver"
	AttributeKeyRefundValue    = "refund_value"
	AttributeKeyAckSuccess     = "success"
	AttributeKeyAckError       = "error"
)

// IBC transfer events vars
var (
	AttributeValueCategory = fmt.Sprintf("%s_%s", ibctypes.ModuleName, ModuleName)
)
