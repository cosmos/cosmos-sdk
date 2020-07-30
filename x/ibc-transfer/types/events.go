package types

// IBC transfer events
const (
	EventTypeTimeout      = "timeout"
	EventTypePacket       = "fungible_token_packet"
	EventTypeTransfer     = "ibc_transfer"
	EventTypeChannelClose = "channel_closed"

	AttributeKeyReceiver       = "receiver"
	AttributeKeyValue          = "value"
	AttributeKeyRefundReceiver = "refund_receiver"
	AttributeKeyRefundValue    = "refund_value"
	AttributeKeyAckSuccess     = "success"
	AttributeKeyAckError       = "error"
)
