package types

import (
	"github.com/tendermint/tendermint/crypto"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
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
		var msg Msg
		err := unpacker.UnpackAny(any, &msg)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *SignerInfo) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var pubkey crypto.PubKey
	return unpacker.UnpackAny(m.PublicKey, &pubkey)
}

func (m *AuthInfo) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	for _, si := range m.SignerInfos {
		err := si.UnpackInterfaces(unpacker)
		if err != nil {
			return err
		}
	}
	return nil
}
