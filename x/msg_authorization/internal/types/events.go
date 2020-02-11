package types

// msg_authorization module events
const (
	EventGrantAuthorization   = "grant-authorization"
	EventRevokeAuthorization  = "revoke-authorization"
	EventExecuteAuthorization = "execute-authorization"

	AttributeGranttype         = "grant-type"
	AttributeKeyGranteeAddress = "grantee"
	AttributeKeyGranterAddress = "granter"

	AttributeValueCategory = ModuleName
)
