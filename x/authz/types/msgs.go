package types

import (
	"time"

	"github.com/gogo/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz/exported"
)

var (
	_ sdk.Msg = &MsgGrantAuthorizationRequest{}
	_ sdk.Msg = &MsgRevokeAuthorizationRequest{}
	_ sdk.Msg = &MsgExecAuthorizedRequest{}

	_ types.UnpackInterfacesMessage = &MsgGrantAuthorizationRequest{}
	_ types.UnpackInterfacesMessage = &MsgExecAuthorizedRequest{}
)

// NewMsgGrantAuthorization creates a new MsgGrantAuthorization
//nolint:interfacer
func NewMsgGrantAuthorization(granter sdk.AccAddress, grantee sdk.AccAddress, authorization exported.Authorization, expiration time.Time) (*MsgGrantAuthorizationRequest, error) {
	m := &MsgGrantAuthorizationRequest{
		Granter:    granter.String(),
		Grantee:    grantee.String(),
		Expiration: expiration,
	}
	err := m.SetAuthorization(authorization)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// GetSigners implements Msg
func (msg MsgGrantAuthorizationRequest) GetSigners() []sdk.AccAddress {
	granter, err := sdk.AccAddressFromBech32(msg.Granter)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{granter}
}

// ValidateBasic implements Msg
func (msg MsgGrantAuthorizationRequest) ValidateBasic() error {
	granter, err := sdk.AccAddressFromBech32(msg.Granter)
	if err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "invalid granter address")
	}
	grantee, err := sdk.AccAddressFromBech32(msg.Grantee)
	if err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "invalid granter address")
	}

	if granter.Equals(grantee) {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "granter and grantee cannot be same")
	}

	if msg.Expiration.Unix() < time.Now().Unix() {
		return sdkerrors.Wrap(ErrInvalidExpirationTime, "Time can't be in the past")
	}

	authorization, ok := msg.Authorization.GetCachedValue().(exported.Authorization)
	if !ok {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "expected %T, got %T", (exported.Authorization)(nil), msg.Authorization.GetCachedValue())
	}
	return authorization.ValidateBasic()
}

// GetGrantAuthorization returns the cache value from the MsgGrantAuthorization.Authorization if present.
func (msg *MsgGrantAuthorizationRequest) GetGrantAuthorization() exported.Authorization {
	authorization, ok := msg.Authorization.GetCachedValue().(exported.Authorization)
	if !ok {
		return nil
	}
	return authorization
}

// SetAuthorization converts Authorization to any and adds it to MsgGrantAuthorization.Authorization.
func (msg *MsgGrantAuthorizationRequest) SetAuthorization(authorization exported.Authorization) error {
	m, ok := authorization.(proto.Message)
	if !ok {
		return sdkerrors.Wrapf(sdkerrors.ErrPackAny, "can't proto marshal %T", m)
	}
	any, err := types.NewAnyWithValue(m)
	if err != nil {
		return err
	}
	msg.Authorization = any
	return nil
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (msg MsgExecAuthorizedRequest) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	for _, x := range msg.Msgs {
		var msgExecAuthorized sdk.Msg
		err := unpacker.UnpackAny(x, &msgExecAuthorized)
		if err != nil {
			return err
		}
	}

	return nil
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (msg MsgGrantAuthorizationRequest) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	var authorization exported.Authorization
	return unpacker.UnpackAny(msg.Authorization, &authorization)
}

// NewMsgRevokeAuthorization creates a new MsgRevokeAuthorization
//nolint:interfacer
func NewMsgRevokeAuthorization(granter sdk.AccAddress, grantee sdk.AccAddress, methodName string) MsgRevokeAuthorizationRequest {
	return MsgRevokeAuthorizationRequest{
		Granter:    granter.String(),
		Grantee:    grantee.String(),
		MethodName: methodName,
	}
}

// GetSigners implements Msg
func (msg MsgRevokeAuthorizationRequest) GetSigners() []sdk.AccAddress {
	granter, err := sdk.AccAddressFromBech32(msg.Granter)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{granter}
}

// ValidateBasic implements MsgRequest.ValidateBasic
func (msg MsgRevokeAuthorizationRequest) ValidateBasic() error {
	granter, err := sdk.AccAddressFromBech32(msg.Granter)
	if err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "invalid granter address")
	}
	grantee, err := sdk.AccAddressFromBech32(msg.Grantee)
	if err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "invalid grantee address")
	}

	if granter.Equals(grantee) {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "granter and grantee cannot be same")
	}

	if msg.MethodName == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "missing method name")
	}

	return nil
}

// NewMsgExecAuthorized creates a new MsgExecAuthorized
//nolint:interfacer
func NewMsgExecAuthorized(grantee sdk.AccAddress, msgs []sdk.Msg) MsgExecAuthorizedRequest {
	msgsAny := make([]*types.Any, len(msgs))
	for i, msg := range msgs {
		any, err := types.NewAnyWithValue(msg)
		if err != nil {
			panic(err)
		}

		msgsAny[i] = any
	}

	return MsgExecAuthorizedRequest{
		Grantee: grantee.String(),
		Msgs:    msgsAny,
	}
}

// GetMessages returns the cache values from the MsgExecAuthorized.Msgs if present.
func (msg MsgExecAuthorizedRequest) GetMessages() ([]sdk.Msg, error) {
	msgs := make([]sdk.Msg, len(msg.Msgs))
	for i, msgAny := range msg.Msgs {
		msg, ok := msgAny.GetCachedValue().(sdk.Msg)
		if !ok {
			return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "messages contains %T which is not a sdk.MsgRequest", msgAny)
		}
		msgs[i] = msg
	}

	return msgs, nil
}

// GetSigners implements Msg
func (msg MsgExecAuthorizedRequest) GetSigners() []sdk.AccAddress {
	grantee, err := sdk.AccAddressFromBech32(msg.Grantee)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{grantee}
}

// ValidateBasic implements Msg
func (msg MsgExecAuthorizedRequest) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Grantee)
	if err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "invalid grantee address")
	}

	if len(msg.Msgs) == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "messages cannot be empty")
	}

	return nil
}
