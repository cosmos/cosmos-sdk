package feegrant

// evidence module events
const (
	EventTypeUseFeeGrant    = "use_feegrant"
	EventTypeRevokeFeeGrant = "revoke_feegrant"
	EventTypeSetFeeGrant    = "set_feegrant"

	AttributeKeyGranter = "granter"
	AttributeKeyGrantee = "grantee"

	AttributeValueCategory = ModuleName
)
