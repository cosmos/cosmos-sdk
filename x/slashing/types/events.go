package types

// Slashing module event types
const (
	EventTypeSlash    = "slash"
	EventTypeJailed   = "jail"
	EventTypeLiveness = "liveness"

	AttributeKeyAddress       = "address"
	AttributeKeyHeight        = "height"
	AttributeKeyPower         = "power"
	AttributeKeyMissedBlocks  = "missed_blocks"
	AttributeKeyAmountSlashed = "amount_slashed"

	AttributeValueCategory = ModuleName
)
