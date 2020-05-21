package types

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ codectypes.UnpackInterfacesMessage = &Tx{}
var _ codectypes.UnpackInterfacesMessage = &TxBody{}
var _ codectypes.UnpackInterfacesMessage = &SignDoc{}

func (m *Tx) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	return m.Body.UnpackInterfaces(unpacker)
}

func (m *SignDoc) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	return m.Body.UnpackInterfaces(unpacker)
}

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
