package types

// POA module event types
var (
	EventTypeCompleteUnbonding = "complete_unbonding"
	EventTypeCreateValidator   = "create_validator"
	EventTypeEditValidator     = "edit_validator"
	EventTypeUnbond            = "unbond"

	AttributeKeyValidator      = "validator"
	AttributeKeySrcValidator   = "source_validator"
	AttributeKeyCompletionTime = "completion_time"
	AttributeValueCategory     = ModuleName
)
