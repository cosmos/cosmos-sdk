package sanction

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/sanction/errors"
)

var _ sdk.Msg = &MsgSanction{}

func NewMsgSanction(authority string, addrs ...sdk.AccAddress) *MsgSanction {
	rv := &MsgSanction{
		Authority: authority,
	}
	for _, addr := range addrs {
		rv.Addresses = append(rv.Addresses, addr.String())
	}
	return rv
}

func (m MsgSanction) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("authority, %q: %v", m.Authority, err)
	}
	for i, addr := range m.Addresses {
		_, err = sdk.AccAddressFromBech32(addr)
		if err != nil {
			return sdkerrors.ErrInvalidAddress.Wrapf("addresses[%d], %q: %v", i, addr, err)
		}
	}
	return nil
}

func (m MsgSanction) GetSigners() []sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(m.Authority)
	return []sdk.AccAddress{addr}
}

var _ sdk.Msg = &MsgUnsanction{}

func NewMsgUnsanction(authority string, addrs ...sdk.AccAddress) *MsgUnsanction {
	rv := &MsgUnsanction{
		Authority: authority,
	}
	for _, addr := range addrs {
		rv.Addresses = append(rv.Addresses, addr.String())
	}
	return rv
}

func (m MsgUnsanction) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("authority, %q: %v", m.Authority, err)
	}
	for i, addr := range m.Addresses {
		_, err = sdk.AccAddressFromBech32(addr)
		if err != nil {
			return sdkerrors.ErrInvalidAddress.Wrapf("addresses[%d], %q: %v", i, addr, err)
		}
	}
	return nil
}

func (m MsgUnsanction) GetSigners() []sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(m.Authority)
	return []sdk.AccAddress{addr}
}

var _ sdk.Msg = &MsgUpdateParams{}

func NewMsgUpdateParams(authority string, minDepSanction, minDepUnsanction sdk.Coins) *MsgUpdateParams {
	rv := &MsgUpdateParams{
		Authority: authority,
		Params: &Params{
			ImmediateSanctionMinDeposit:   minDepSanction,
			ImmediateUnsanctionMinDeposit: minDepUnsanction,
		},
	}
	return rv
}

func (m MsgUpdateParams) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("authority, %q: %v", m.Authority, err)
	}
	if m.Params != nil {
		err = m.Params.ValidateBasic()
		if err != nil {
			return errors.ErrInvalidParams.Wrap(err.Error())
		}
	}
	return nil
}

func (m MsgUpdateParams) GetSigners() []sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(m.Authority)
	return []sdk.AccAddress{addr}
}
