package testutil

import (
	errorsmod "cosmossdk.io/errors"

	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgCounter{},
		&MsgCounter2{},
		&MsgKeyValue{},
		&MsgNestedMessages{},
		&MsgSend{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Counter_serviceDesc)
	msgservice.RegisterMsgServiceDesc(registry, &_Counter2_serviceDesc)
	msgservice.RegisterMsgServiceDesc(registry, &_KeyValue_serviceDesc)
	msgservice.RegisterMsgServiceDesc(registry, &_NestedMessages_serviceDesc)
	msgservice.RegisterMsgServiceDesc(registry, &_Send_serviceDesc)

	codec.RegisterInterfaces(registry)
}

var _ sdk.Msg = &MsgCounter{}

func (msg *MsgCounter) GetSigners() []sdk.AccAddress { return []sdk.AccAddress{} }
func (msg *MsgCounter) ValidateBasic() error {
	if msg.Counter >= 0 {
		return nil
	}
	return errorsmod.Wrap(sdkerrors.ErrInvalidSequence, "counter should be a non-negative integer")
}

var _ sdk.Msg = &MsgCounter2{}

func (msg *MsgCounter2) GetSigners() []sdk.AccAddress { return []sdk.AccAddress{} }
func (msg *MsgCounter2) ValidateBasic() error {
	if msg.Counter >= 0 {
		return nil
	}
	return errorsmod.Wrap(sdkerrors.ErrInvalidSequence, "counter should be a non-negative integer")
}

var _ sdk.Msg = &MsgKeyValue{}

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

func (msg *MsgNestedMessages) GetMsgs() ([]sdk.Msg, error) {
	cdc := codectestutil.CodecOptions{}.NewCodec()
	RegisterInterfaces(cdc.InterfaceRegistry())
	msgs := make([]sdk.Msg, len(msg.GetMessages()))
	for i, m := range msg.GetMessages() {
		mm, err := cdc.InterfaceRegistry().Resolve(m.TypeUrl)
		if err != nil {
			return nil, err
		}
		err = cdc.UnpackAny(m, &mm)
		if err != nil {
			return nil, err
		}
		msgs[i] = mm
	}
	return msgs, nil
}
