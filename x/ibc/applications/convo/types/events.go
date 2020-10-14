package types

// IBC convo events
const (
	EventTypeConvo        = "convo"
	EventTypeTimeout      = "timeout"
	EventTypePacket       = "convo_packet"
	EventTypeChannelClose = "channel_closed"
	EventTypeDenomTrace   = "denomination_trace"

	AttributeKeyReceiver   = "receiver"
	AttributeKeyAckSuccess = "success"
	AttributeKeyAck        = "acknowledgement"
	AttributeKeyAckError   = "error"
)
