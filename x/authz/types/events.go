package types

// authz module events
const (
	EventGrantAuthorization   = "grant-authorization"
	EventRevokeAuthorization  = "revoke-authorization"
	EventExecuteAuthorization = "execute-authorization"

	AttributeKeyGrantType      = "grant-type"
	AttributeKeyGranteeAddress = "grantee"
	AttributeKeyGranterAddress = "granter"

	AttributeValueCategory = ModuleName
)
