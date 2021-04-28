package authz

import (
	"time"

	"github.com/gogo/protobuf/proto"

	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ sdk.MsgRequest = &MsgGrant{}
	_ sdk.MsgRequest = &MsgRevoke{}
	_ sdk.MsgRequest = &MsgExec{}

	_ cdctypes.UnpackInterfacesMessage = &MsgGrant{}
	_ cdctypes.UnpackInterfacesMessage = &MsgExec{}
)

// NewMsgGrant creates a new MsgGrant
//nolint:interfacer
func NewMsgGrant(granter sdk.AccAddress, grantee sdk.AccAddress, a Authorization, expiration time.Time) (*MsgGrant, error) {
	m := &MsgGrant{
		Granter:    granter.String(),
		Grantee:    grantee.String(),
		Expiration: expiration,
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

	if msg.Expiration.Unix() < time.Now().Unix() {
		return sdkerrors.Wrap(ErrInvalidExpirationTime, "Time can't be in the past")
	}

	av := msg.Authorization.GetCachedValue()
	a, ok := av.(Authorization)
	if !ok {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "expected %T, got %T", (Authorization)(nil), av)
	}
	return a.ValidateBasic()
}

// GetAuthorization returns the cache value from the MsgGrant.Authorization if present.
func (msg *MsgGrant) GetAuthorization() Authorization {
	a, ok := msg.Authorization.GetCachedValue().(Authorization)
	if !ok {
		return nil
	}
	return a
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
	msg.Authorization = any
	return nil
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (msg MsgExec) UnpackInterfaces(unpacker cdctypes.AnyUnpacker) error {
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
func (msg MsgGrant) UnpackInterfaces(unpacker cdctypes.AnyUnpacker) error {
	var a Authorization
	return unpacker.UnpackAny(msg.Authorization, &a)
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

// NewMsgExec creates a new MsgExecAuthorized
//nolint:interfacer
func NewMsgExec(grantee sdk.AccAddress, msgs []sdk.ServiceMsg) MsgExec {
	msgsAny := make([]*cdctypes.Any, len(msgs))
	for i, msg := range msgs {
		bz, err := proto.Marshal(msg.Request)
		if err != nil {
			panic(err)
		}

		anyMsg := &cdctypes.Any{
			TypeUrl: msg.MethodName,
			Value:   bz,
		}

		msgsAny[i] = anyMsg
	}

	return MsgExec{
		Grantee: grantee.String(),
		Msgs:    msgsAny,
	}
}

// GetServiceMsgs returns the cache values from the MsgExecAuthorized.Msgs if present.
func (msg MsgExec) GetServiceMsgs() ([]sdk.ServiceMsg, error) {
	msgs := make([]sdk.ServiceMsg, len(msg.Msgs))
	for i, msgAny := range msg.Msgs {
		msgReq, ok := msgAny.GetCachedValue().(sdk.MsgRequest)
		if !ok {
			return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "messages contains %T which is not a sdk.MsgRequest", msgAny)
		}
		srvMsg := sdk.ServiceMsg{
			MethodName: msgAny.TypeUrl,
			Request:    msgReq,
		}

		msgs[i] = srvMsg
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
