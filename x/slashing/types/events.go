package types

// Slashing module event types
var (
	EventTypeSlash  = "slash"
	EventTypeUnjail = "unjail"

	AttributeKeyAddress = "address"
	AttributeKeyPower   = "power"
	AttributeKeyReason  = "reason"
	AttributeKeyJailed  = "jailed"

	AttributeValueDoubleSign       = "double_sign"
	AttributeValueMissingSignature = "missing_signature"
	AttributeValueCategory         = "slashing"
)
