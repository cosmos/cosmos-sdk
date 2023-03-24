package testutil

import (
	errorsmod "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

var (
	_ sdk.Msg = &MsgCounter{}
	_ sdk.Msg = &MsgKeyValue{}
	_ sdk.Msg = &MsgCounter2{}
)

func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgCounter{},
		&MsgCounter2{},
		&MsgKeyValue{},
	)
	msgservice.RegisterMsgServiceDesc(registry, &_Counter_serviceDesc)
	msgservice.RegisterMsgServiceDesc(registry, &_Counter2_serviceDesc)
	msgservice.RegisterMsgServiceDesc(registry, &_KeyValue_serviceDesc)

	codec.RegisterInterfaces(registry)
}

func (*MsgCounter) GetSigners() []sdk.AccAddress { return []sdk.AccAddress{} }
func (msg *MsgCounter) ValidateBasic() error {
	if msg.Counter >= 0 {
		return nil
	}
	return errorsmod.Wrap(sdkerrors.ErrInvalidSequence, "counter should be a non-negative integer")
}

func (*MsgCounter2) GetSigners() []sdk.AccAddress { return []sdk.AccAddress{} }
func (msg *MsgCounter2) ValidateBasic() error {
	if msg.Counter >= 0 {
		return nil
	}
	return errorsmod.Wrap(sdkerrors.ErrInvalidSequence, "counter should be a non-negative integer")
}

func (msg *MsgKeyValue) GetSigners() []sdk.AccAddress {
	if len(msg.Signer) == 0 {
		return []sdk.AccAddress{}
	}

	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(msg.Signer)}
}

func (msg *MsgKeyValue) ValidateBasic() error {
	if msg.Key == nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "key cannot be nil")
	}
	if msg.Value == nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "value cannot be nil")
	}
	return nil
}
