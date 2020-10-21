package types

import (
	"fmt"
	"time"

	types "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/gogo/protobuf/proto"
	"gopkg.in/yaml.v2"
)

// msg_authorization message types
const (
	TypeMsgGrantAuthorization  = "grant_authorization"
	TypeMsgRevokeAuthorization = "revoke_authorization"
	TypeMsgExecDelegated       = "exec_delegated"
)

var (
	_ sdk.Msg = &MsgGrantAuthorization{}
	_ sdk.Msg = &MsgRevokeAuthorization{}
	_ sdk.Msg = &MsgExecAuthorized{}

	_ types.UnpackInterfacesMessage = &MsgGrantAuthorization{}
	_ types.UnpackInterfacesMessage = &MsgExecAuthorized{}
)

// NewMsgGrantAuthorization creates a new MsgGrantAuthorization
//nolint:interfacer
func NewMsgGrantAuthorization(granter sdk.AccAddress, grantee sdk.AccAddress, authorization Authorization, expiration time.Time) (*MsgGrantAuthorization, error) {
	m := &MsgGrantAuthorization{
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

// Route implements Msg
func (msg MsgGrantAuthorization) Route() string { return RouterKey }

// Type implements Msg
func (msg MsgGrantAuthorization) Type() string { return TypeMsgGrantAuthorization }

// GetSigners implements Msg
func (msg MsgGrantAuthorization) GetSigners() []sdk.AccAddress {
	granter, err := sdk.AccAddressFromBech32(msg.Granter)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{granter}
}

// GetSignBytes implements Msg
func (msg MsgGrantAuthorization) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements Msg
func (msg MsgGrantAuthorization) ValidateBasic() error {
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

func (msg *MsgGrantAuthorization) GetAuthorization() Authorization {
	authorization, ok := msg.Authorization.GetCachedValue().(Authorization)
	if !ok {
		return nil
	}
	return authorization
}

func (msg *MsgGrantAuthorization) SetAuthorization(authorization Authorization) error {
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

func (msg MsgExecAuthorized) UnpackInterfaces(unpacker types.AnyUnpacker) error {
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
func (msg MsgGrantAuthorization) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	var authorization Authorization
	return unpacker.UnpackAny(msg.Authorization, &authorization)
}

// String implements the Stringer interface
func (msg MsgGrantAuthorization) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

// NewMsgRevokeAuthorization creates a new MsgRevokeAuthorization
//nolint:interfacer
func NewMsgRevokeAuthorization(granter sdk.AccAddress, grantee sdk.AccAddress, authorizationMsgType string) MsgRevokeAuthorization {
	return MsgRevokeAuthorization{
		Granter:              granter.String(),
		Grantee:              grantee.String(),
		AuthorizationMsgType: authorizationMsgType,
	}
}

// Route implements Msg
func (msg MsgRevokeAuthorization) Route() string { return RouterKey }

// Type implements Msg
func (msg MsgRevokeAuthorization) Type() string { return TypeMsgRevokeAuthorization }

// GetSigners implements Msg
func (msg MsgRevokeAuthorization) GetSigners() []sdk.AccAddress {
	granter, err := sdk.AccAddressFromBech32(msg.Granter)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{granter}
}

// GetSignBytes implements Msg
func (msg MsgRevokeAuthorization) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements Msg
func (msg MsgRevokeAuthorization) ValidateBasic() error {
	if msg.Granter == "" {
		return sdkerrors.Wrap(ErrInvalidGranter, "missing granter address")
	}
	if msg.Grantee == "" {
		return sdkerrors.Wrap(ErrInvalidGrantee, "missing grantee address")
	}
	return nil
}

// String implements the Stringer interface
func (msg MsgRevokeAuthorization) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

// NewMsgExecAuthorized creates a new MsgExecAuthorized
//nolint:interfacer
func NewMsgExecAuthorized(grantee sdk.AccAddress, msgs []sdk.ServiceMsg) MsgExecAuthorized {
	msgsAny := make([]*types.Any, len(msgs))
	for i, msg := range msgs {
		// msg1, ok := msg.(proto.Message)
		// if !ok {
		// 	panic(fmt.Errorf("cannot proto marshal %T", msg1))
		// }
		any, err := types.NewAnyWithValue(&msg)
		if err != nil {
			panic(err)
		}

		msgsAny[i] = any
	}
	return MsgExecAuthorized{
		Grantee: grantee.String(),
		Msgs:    msgsAny,
	}
}

// GetMsgs Unpacks any messages
func (msg MsgExecAuthorized) GetMsgs() ([]sdk.ServiceMsg, error) {
	msgs := make([]sdk.ServiceMsg, len(msg.Msgs))
	for i, msgAny := range msg.Msgs {
		msg1, ok := msgAny.GetCachedValue().(sdk.ServiceMsg)
		if !ok {
			return nil, fmt.Errorf("cannot proto marshal %T", msg1)
		}
		msgs[i] = msg1
	}
	return msgs, nil
}

// Route implements Msg
func (msg MsgExecAuthorized) Route() string { return RouterKey }

// Type implements Msg
func (msg MsgExecAuthorized) Type() string { return TypeMsgExecDelegated }

// GetSigners implements Msg
func (msg MsgExecAuthorized) GetSigners() []sdk.AccAddress {
	grantee, err := sdk.AccAddressFromBech32(msg.Grantee)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{grantee}
}

// GetSignBytes implements Msg
func (msg MsgExecAuthorized) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements Msg
func (msg MsgExecAuthorized) ValidateBasic() error {
	if msg.Grantee == "" {
		return sdkerrors.Wrap(ErrInvalidGranter, "missing grantee address")
	}
	return nil
}

// String implements the Stringer interface
func (msg MsgExecAuthorized) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}
