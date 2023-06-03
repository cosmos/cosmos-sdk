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

// Route Implements Msg.
func (m MsgAuthorizeCircuitBreaker) Route() string { return sdk.MsgTypeURL(&m) }

// Type Implements Msg.
func (m MsgAuthorizeCircuitBreaker) Type() string { return sdk.MsgTypeURL(&m) }

// GetSigners returns the expected signers for a MsgAuthorizeCircuitBreaker.
func (m MsgAuthorizeCircuitBreaker) GetSigners() []sdk.AccAddress {
	granter := sdk.MustAccAddressFromBech32(m.Granter)

	return []sdk.AccAddress{granter}
}

// NewMsgTripCircuitBreaker creates a new MsgTripCircuitBreaker instance.
func NewMsgTripCircuitBreaker(authority string, urls []string) *MsgTripCircuitBreaker {
	return &MsgTripCircuitBreaker{
		Authority:   authority,
		MsgTypeUrls: urls,
	}
}

// Route Implements Msg.
func (m MsgTripCircuitBreaker) Route() string { return sdk.MsgTypeURL(&m) }

// Type Implements Msg.
func (m MsgTripCircuitBreaker) Type() string { return sdk.MsgTypeURL(&m) }

// GetSigners returns the expected signers for a MsgTripCircuitBreaker.
func (m MsgTripCircuitBreaker) GetSigners() []sdk.AccAddress {
	granter := sdk.MustAccAddressFromBech32(m.Authority)

	return []sdk.AccAddress{granter}
}

// NewMsgResetCircuitBreaker creates a new MsgResetCircuitBreaker instance.
func NewMsgResetCircuitBreaker(authority string, urls []string) *MsgResetCircuitBreaker {
	return &MsgResetCircuitBreaker{
		Authority:   authority,
		MsgTypeUrls: urls,
	}
}

// Route Implements Msg.
func (m MsgResetCircuitBreaker) Route() string { return sdk.MsgTypeURL(&m) }

// Type Implements Msg.
func (m MsgResetCircuitBreaker) Type() string { return sdk.MsgTypeURL(&m) }

// GetSigners returns the expected signers for a MsgResetCircuitBreaker.
func (m MsgResetCircuitBreaker) GetSigners() []sdk.AccAddress {
	granter := sdk.MustAccAddressFromBech32(m.Authority)

	return []sdk.AccAddress{granter}
}
