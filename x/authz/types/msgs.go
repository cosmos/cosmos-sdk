package types

import (
	"fmt"
	"time"

	"github.com/gogo/protobuf/proto"

	types "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ sdk.MsgRequest = &MsgGrantAuthorizationRequest{}
	_ sdk.MsgRequest = &MsgRevokeAuthorizationRequest{}
	_ sdk.MsgRequest = &MsgExecAuthorizedRequest{}

	_ types.UnpackInterfacesMessage = &MsgGrantAuthorizationRequest{}
	_ types.UnpackInterfacesMessage = &MsgExecAuthorizedRequest{}
)

// NewMsgGrantAuthorization creates a new MsgGrantAuthorization
//nolint:interfacer
func NewMsgGrantAuthorization(granter sdk.AccAddress, grantee sdk.AccAddress, authorization Authorization, expiration time.Time) (*MsgGrantAuthorizationRequest, error) {
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
	if msg.Granter == "" {
		return sdkerrors.Wrap(ErrInvalidGranter, "missing granter address")
	}
	if msg.Grantee == "" {
		return sdkerrors.Wrap(ErrInvalidGrantee, "missing grantee address")
	}
	if msg.Expiration.Unix() < time.Now().Unix() {
		return sdkerrors.Wrap(ErrInvalidExpirationTime, "Time can't be in the past")
	}

	return nil
}

// GetGrantAuthorization returns the cache value from the MsgGrantAuthorization.Authorization if present.
func (msg *MsgGrantAuthorizationRequest) GetGrantAuthorization() Authorization {
	authorization, ok := msg.Authorization.GetCachedValue().(Authorization)
	if !ok {
		return nil
	}
	return authorization
}

// SetAuthorization converts Authorization to any and adds it to MsgGrantAuthorization.Authorization.
func (msg *MsgGrantAuthorizationRequest) SetAuthorization(authorization Authorization) error {
	m, ok := authorization.(proto.Message)
	if !ok {
		return fmt.Errorf("can't proto marshal %T", m)
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
		var msgExecAuthorized sdk.MsgRequest
		err := unpacker.UnpackAny(x, &msgExecAuthorized)
		if err != nil {
			return err
		}
	}

	return nil
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (msg MsgGrantAuthorizationRequest) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	var authorization Authorization
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
	if msg.Granter == "" {
		return sdkerrors.Wrap(ErrInvalidGranter, "missing granter address")
	}
	if msg.Grantee == "" {
		return sdkerrors.Wrap(ErrInvalidGrantee, "missing grantee address")
	}
	return nil
}

// NewMsgExecAuthorized creates a new MsgExecAuthorized
//nolint:interfacer
func NewMsgExecAuthorized(grantee sdk.AccAddress, msgs []sdk.ServiceMsg) MsgExecAuthorizedRequest {
	msgsAny := make([]*types.Any, len(msgs))
	for i, msg := range msgs {
		bz, err := proto.Marshal(msg.Request)
		if err != nil {
			panic(err)
		}

		anyMsg := &types.Any{
			TypeUrl: msg.MethodName,
			Value:   bz,
		}

		msgsAny[i] = anyMsg
	}

	return MsgExecAuthorizedRequest{
		Grantee: grantee.String(),
		Msgs:    msgsAny,
	}
}

// GetServiceMsgs returns the cache values from the MsgExecAuthorized.Msgs if present.
func (msg MsgExecAuthorizedRequest) GetServiceMsgs() ([]sdk.ServiceMsg, error) {
	msgs := make([]sdk.ServiceMsg, len(msg.Msgs))
	for i, msgAny := range msg.Msgs {
		msg1 := sdk.ServiceMsg{
			MethodName: msgAny.TypeUrl,
			Request:    msgAny.GetCachedValue().(sdk.MsgRequest),
		}

		msgs[i] = msg1
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
	if msg.Grantee == "" {
		return sdkerrors.Wrap(ErrInvalidGranter, "missing grantee address")
	}
	return nil
}
