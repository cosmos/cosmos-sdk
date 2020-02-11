package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// MsgGrantAuthorization grants the provided authorization to the grantee on the granter's
// account with the provided expiration time.
type MsgGrantAuthorization struct {
	Granter       sdk.AccAddress `json:"granter"`
	Grantee       sdk.AccAddress `json:"grantee"`
	Authorization Authorization  `json:"authorization"`
	// Expiration specifies the expiration time of the grant
	Expiration time.Time `json:"expiration"`
}

func NewMsgGrantAuthorization(granter sdk.AccAddress, grantee sdk.AccAddress, authorization Authorization, expiration time.Time) MsgGrantAuthorization {
	return MsgGrantAuthorization{
		Granter:       granter,
		Grantee:       grantee,
		Authorization: authorization,
		Expiration:    expiration,
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

func (msg MsgGrantAuthorization) ValidateBasic() error {
	if msg.Granter.Empty() {
		return sdkerrors.Wrap(ErrInvalidGranter, "missing granter address")
	}
	if msg.Grantee.Empty() {
		return sdkerrors.Wrap(ErrInvalidGranter, "missing grantee address")
	}
	if msg.Expiration.Unix() < time.Now().Unix() {
		return sdkerrors.Wrap(ErrInvalidGranter, "Time can't be in the past")
	}

	return nil
}

// MsgRevokeAuthorization revokes any authorization with the provided sdk.Msg type on the
// granter's account with that has been granted to the grantee.
type MsgRevokeAuthorization struct {
	Granter sdk.AccAddress `json:"granter"`
	Grantee sdk.AccAddress `json:"grantee"`
	// AuthorizationMsgType is the type of sdk.Msg that the revoked Authorization refers to.
	// i.e. this is what `Authorization.MsgType()` returns
	AuthorizationMsgType string `json:"authorization_msg_type"`
}

func NewMsgRevokeAuthorization(granter sdk.AccAddress, grantee sdk.AccAddress, authorizationMsgType string) MsgRevokeAuthorization {
	return MsgRevokeAuthorization{
		Granter:              granter,
		Grantee:              grantee,
		AuthorizationMsgType: authorizationMsgType,
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

func (msg MsgRevokeAuthorization) ValidateBasic() error {
	if msg.Granter.Empty() {
		return sdkerrors.Wrap(ErrInvalidGranter, "missing granter address")
	}
	if msg.Grantee.Empty() {
		return sdkerrors.Wrap(ErrInvalidGranter, "missing grantee address")
	}
	return nil
}

// MsgExecAuthorized attempts to execute the provided messages using
// authorizations granted to the grantee. Each message should have only
// one signer corresponding to the granter of the authorization.
type MsgExecAuthorized struct {
	Grantee sdk.AccAddress `json:"grantee"`
	Msgs    []sdk.Msg      `json:"msgs"`
}

func NewMsgExecAuthorized(grantee sdk.AccAddress, msg []sdk.Msg) MsgExecAuthorized {
	return MsgExecAuthorized{
		Grantee: grantee,
		Msgs:    msg,
	}
}

func (msg MsgExecAuthorized) Route() string { return RouterKey }
func (msg MsgExecAuthorized) Type() string  { return "exec_delegated" }

func (msg MsgExecAuthorized) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Grantee}
}

func (msg MsgExecAuthorized) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgExecAuthorized) ValidateBasic() error {
	if msg.Grantee.Empty() {
		return sdkerrors.Wrap(ErrInvalidGranter, "missing grantee address")
	}
	return nil
}
