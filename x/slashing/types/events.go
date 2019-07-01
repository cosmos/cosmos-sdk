package types

// Slashing module event types
var (
	EventTypeSlash    = "slash"
	EventTypeLiveness = "liveness"

	AttributeKeyAddress      = "address"
	AttributeKeyHeight       = "height"
	AttributeKeyPower        = "power"
	AttributeKeyReason       = "reason"
	AttributeKeyJailed       = "jailed"
	AttributeKeyMissedBlocks = "missed_blocks"

	AttributeValueDoubleSign       = "double_sign"
	AttributeValueMissingSignature = "missing_signature"
	AttributeValueCategory         = ModuleName
)
