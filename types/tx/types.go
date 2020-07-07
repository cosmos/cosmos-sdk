package tx

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

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
