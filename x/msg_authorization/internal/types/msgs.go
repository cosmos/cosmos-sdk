package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MsgGrant grants the provided capability to the grantee on the granter's
// account with the provided expiration time.
type MsgGrantAuthorization struct {
	Granter    sdk.AccAddress `json:"granter"`
	Grantee    sdk.AccAddress `json:"grantee"`
	Capability Capability     `json:"capability"`
	// Expiration specifies the expiration time of the grant
	Expiration time.Time `json:"expiration"`
}

func NewMsgGrantAuthorization(granter sdk.AccAddress, grantee sdk.AccAddress, capability Capability, expiration time.Time) MsgGrantAuthorization {
	return MsgGrantAuthorization{
		Granter:    granter,
		Grantee:    grantee,
		Capability: capability,
		Expiration: expiration,
	}
}

func (msg MsgGrantAuthorization) Route() string { return RouterKey }
func (msg MsgGrantAuthorization) Type() string  { return "grant_authorization" }

func (msg MsgGrantAuthorization) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Granter}
}

func (msg MsgGrantAuthorization) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgGrantAuthorization) ValidateBasic() sdk.Error {
	if msg.Granter.Empty() {
		return ErrInvalidGranter(DefaultCodespace)
	}
	if msg.Grantee.Empty() {
		return ErrInvalidGrantee(DefaultCodespace)
	}
	if msg.Expiration.Unix() < time.Now().Unix() {
		return ErrInvalidExpirationTime(DefaultCodespace)
	}

	return nil
}

// MsgRevokeAuthorization revokes any capability with the provided sdk.Msg type on the
// granter's account with that has been granted to the grantee.
type MsgRevokeAuthorization struct {
	Granter sdk.AccAddress `json:"granter"`
	Grantee sdk.AccAddress `json:"grantee"`
	// CapabilityMsgType is the type of sdk.Msg that the revoked Capability refers to.
	// i.e. this is what `Capability.MsgType()` returns
	CapabilityMsgType sdk.Msg `json:"capability_msg_type"`
}

func NewMsgRevokeAuthorization(granter sdk.AccAddress, grantee sdk.AccAddress, capabilityMsgType sdk.Msg) MsgRevokeAuthorization {
	return MsgRevokeAuthorization{
		Granter:           granter,
		Grantee:           grantee,
		CapabilityMsgType: capabilityMsgType,
	}
}

func (msg MsgRevokeAuthorization) Route() string { return RouterKey }
func (msg MsgRevokeAuthorization) Type() string  { return "revoke_authorization" }

func (msg MsgRevokeAuthorization) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Granter}
}

func (msg MsgRevokeAuthorization) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgRevokeAuthorization) ValidateBasic() sdk.Error {
	if msg.Granter.Empty() {
		return sdk.ErrInvalidAddress(msg.Granter.String())
	}
	if msg.Grantee.Empty() {
		return sdk.ErrInvalidAddress(msg.Grantee.String())
	}
	return nil
}

// MsgExecDelegated attempts to execute the provided messages using
// capabilities granted to the grantee. Each message should have only
// one signer corresponding to the granter of the capability.
type MsgExecDelegated struct {
	Grantee sdk.AccAddress `json:"grantee"`
	Msgs    []sdk.Msg      `json:"msg"`
}

func NewMsgExecDelegated(grantee sdk.AccAddress, msgs []sdk.Msg) MsgExecDelegated {
	return MsgExecDelegated{
		Grantee: grantee,
		Msgs:    msgs,
	}
}
func (msg MsgExecDelegated) Route() string { return RouterKey }
func (msg MsgExecDelegated) Type() string  { return "exec_delegated" }

func (msg MsgExecDelegated) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Grantee}
}

func (msg MsgExecDelegated) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgExecDelegated) ValidateBasic() sdk.Error {
	if msg.Grantee.Empty() {
		return sdk.ErrInvalidAddress(msg.Grantee.String())
	}
	return nil
}
