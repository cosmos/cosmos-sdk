package authz

import (
	"time"

	"github.com/gogo/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/codec/legacy"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"
)

var (
	_ sdk.Msg = &MsgGrant{}
	_ sdk.Msg = &MsgRevoke{}
	_ sdk.Msg = &MsgExec{}

	// For amino support.
	_ legacytx.LegacyMsg = &MsgGrant{}
	_ legacytx.LegacyMsg = &MsgRevoke{}
	_ legacytx.LegacyMsg = &MsgExec{}

	_ cdctypes.UnpackInterfacesMessage = &MsgGrant{}
	_ cdctypes.UnpackInterfacesMessage = &MsgExec{}
)

// NewMsgGrant creates a new MsgGrant
//nolint:interfacer
func NewMsgGrant(granter sdk.AccAddress, grantee sdk.AccAddress, a Authorization, expiration time.Time) (*MsgGrant, error) {
	m := &MsgGrant{
		Granter: granter.String(),
		Grantee: grantee.String(),
		Grant:   Grant{Expiration: expiration},
	}
	err := m.SetAuthorization(a)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// GetSigners implements Msg
func (msg MsgGrant) GetSigners() []sdk.AccAddress {
	granter, err := sdk.AccAddressFromBech32(msg.Granter)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{granter}
}

// ValidateBasic implements Msg
func (msg MsgGrant) ValidateBasic() error {
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
	return msg.Grant.ValidateBasic()
}

// Type implements the LegacyMsg.Type method.
func (msg MsgGrant) Type() string {
	return sdk.MsgTypeURL(&msg)
}

// Route implements the LegacyMsg.Route method.
func (msg MsgGrant) Route() string {
	return sdk.MsgTypeURL(&msg)
}

// GetSignBytes implements the LegacyMsg.GetSignBytes method.
func (msg MsgGrant) GetSignBytes() []byte {
	return sdk.MustSortJSON(legacy.Cdc.MustMarshalJSON(&msg))
}

// GetAuthorization returns the cache value from the MsgGrant.Authorization if present.
func (msg *MsgGrant) GetAuthorization() Authorization {
	return msg.Grant.GetAuthorization()
}

// SetAuthorization converts Authorization to any and adds it to MsgGrant.Authorization.
func (msg *MsgGrant) SetAuthorization(a Authorization) error {
	m, ok := a.(proto.Message)
	if !ok {
		return sdkerrors.Wrapf(sdkerrors.ErrPackAny, "can't proto marshal %T", m)
	}
	any, err := cdctypes.NewAnyWithValue(m)
	if err != nil {
		return err
	}
	msg.Grant.Authorization = any
	return nil
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (msg MsgExec) UnpackInterfaces(unpacker cdctypes.AnyUnpacker) error {
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
func (msg MsgGrant) UnpackInterfaces(unpacker cdctypes.AnyUnpacker) error {
	return msg.Grant.UnpackInterfaces(unpacker)
}

// NewMsgRevoke creates a new MsgRevoke
//nolint:interfacer
func NewMsgRevoke(granter sdk.AccAddress, grantee sdk.AccAddress, msgTypeURL string) MsgRevoke {
	return MsgRevoke{
		Granter:    granter.String(),
		Grantee:    grantee.String(),
		MsgTypeUrl: msgTypeURL,
	}
}

// GetSigners implements Msg
func (msg MsgRevoke) GetSigners() []sdk.AccAddress {
	granter, err := sdk.AccAddressFromBech32(msg.Granter)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{granter}
}

// ValidateBasic implements MsgRequest.ValidateBasic
func (msg MsgRevoke) ValidateBasic() error {
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

	if msg.MsgTypeUrl == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "missing method name")
	}

	return nil
}

// Type implements the LegacyMsg.Type method.
func (msg MsgRevoke) Type() string {
	return sdk.MsgTypeURL(&msg)
}

// Route implements the LegacyMsg.Route method.
func (msg MsgRevoke) Route() string {
	return sdk.MsgTypeURL(&msg)
}

// GetSignBytes implements the LegacyMsg.GetSignBytes method.
func (msg MsgRevoke) GetSignBytes() []byte {
	return sdk.MustSortJSON(legacy.Cdc.MustMarshalJSON(&msg))
}

// NewMsgExec creates a new MsgExecAuthorized
//nolint:interfacer
func NewMsgExec(grantee sdk.AccAddress, msgs []sdk.Msg) MsgExec {
	msgsAny := make([]*cdctypes.Any, len(msgs))
	for i, msg := range msgs {
		any, err := cdctypes.NewAnyWithValue(msg)
		if err != nil {
			panic(err)
		}

		msgsAny[i] = any
	}

	return MsgExec{
		Grantee: grantee.String(),
		Msgs:    msgsAny,
	}
}

// GetMessages returns the cache values from the MsgExecAuthorized.Msgs if present.
func (msg MsgExec) GetMessages() ([]sdk.Msg, error) {
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
func (msg MsgExec) GetSigners() []sdk.AccAddress {
	grantee, err := sdk.AccAddressFromBech32(msg.Grantee)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{grantee}
}

// ValidateBasic implements Msg
func (msg MsgExec) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Grantee)
	if err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "invalid grantee address")
	}

	if len(msg.Msgs) == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "messages cannot be empty")
	}

	return nil
}

// Type implements the LegacyMsg.Type method.
func (msg MsgExec) Type() string {
	return sdk.MsgTypeURL(&msg)
}

// Route implements the LegacyMsg.Route method.
func (msg MsgExec) Route() string {
	return sdk.MsgTypeURL(&msg)
}

// GetSignBytes implements the LegacyMsg.GetSignBytes method.
func (msg MsgExec) GetSignBytes() []byte {
	return sdk.MustSortJSON(legacy.Cdc.MustMarshalJSON(&msg))
}
