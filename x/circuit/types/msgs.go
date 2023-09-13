package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	_ sdk.Msg = &MsgAuthorizeCircuitBreaker{}
	_ sdk.Msg = &MsgTripCircuitBreaker{}
	_ sdk.Msg = &MsgResetCircuitBreaker{}
)

// NewMsgAuthorizeCircuitBreaker creates a new MsgAuthorizeCircuitBreaker instance.
func NewMsgAuthorizeCircuitBreaker(granter, grantee string, permission *Permissions) *MsgAuthorizeCircuitBreaker {
	return &MsgAuthorizeCircuitBreaker{
		Granter:     granter,
		Grantee:     grantee,
		Permissions: permission,
	}
}

// NewMsgTripCircuitBreaker creates a new MsgTripCircuitBreaker instance.
func NewMsgTripCircuitBreaker(authority string, urls []string) *MsgTripCircuitBreaker {
	return &MsgTripCircuitBreaker{
		Authority:   authority,
		MsgTypeUrls: urls,
	}
}

// NewMsgResetCircuitBreaker creates a new MsgResetCircuitBreaker instance.
func NewMsgResetCircuitBreaker(authority string, urls []string) *MsgResetCircuitBreaker {
	return &MsgResetCircuitBreaker{
		Authority:   authority,
		MsgTypeUrls: urls,
	}
}
