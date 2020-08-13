package tx

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Tx = Tx{}

// GetMsgs implements sdk.Tx for Tx.
func (m Tx) GetMsgs() []sdk.Msg {
	anys := m.Body.Messages
	res := make([]sdk.Msg, len(anys))
	for i, any := range anys {
		msg := any.GetCachedValue().(sdk.Msg)
		res[i] = msg
	}

	return res
}

// ValidateBasic implements sdk.Tx for Tx.
func (m Tx) ValidateBasic() error {
	msgs := m.GetMsgs()

	if len(msgs) == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "must contain at least one message")
	}

	for _, msg := range msgs {
		err := msg.ValidateBasic()
		if err != nil {
			return err
		}
	}

	return nil
}

var _, _ codectypes.UnpackInterfacesMessage = &Tx{}, &TxBody{}

// UnpackInterfaces implements the UnpackInterfaceMessages.UnpackInterfaces method
func (m *Tx) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	if m.Body != nil {
		return m.Body.UnpackInterfaces(unpacker)
	}
	return nil
}

// UnpackInterfaces implements the UnpackInterfaceMessages.UnpackInterfaces method
func (m *TxBody) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	for _, any := range m.Messages {
		var msg sdk.Msg
		err := unpacker.UnpackAny(any, &msg)
		if err != nil {
			return err
		}
	}
	return nil
}
