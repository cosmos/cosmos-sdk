package tx

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto"
)

var _, _, _, _ codectypes.UnpackInterfacesMessage = &Tx{}, &TxBody{}, &AuthInfo{}, &SignerInfo{}

// UnpackInterfaces implements the UnpackInterfaceMessages.UnpackInterfaces method
func (m *Tx) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	if m.Body != nil {
		if err := m.Body.UnpackInterfaces(unpacker); err != nil {
			return err
		}
	}

	if m.AuthInfo != nil {
		return m.AuthInfo.UnpackInterfaces(unpacker)
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

// UnpackInterfaces implements the UnpackInterfaceMessages.UnpackInterfaces method
func (m *AuthInfo) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	for _, signerInfo := range m.SignerInfos {
		err := signerInfo.UnpackInterfaces(unpacker)
		if err != nil {
			return err
		}
	}
	return nil
}

// UnpackInterfaces implements the UnpackInterfaceMessages.UnpackInterfaces method
func (m *SignerInfo) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var pubKey crypto.PubKey
	err := unpacker.UnpackAny(m.PublicKey, &pubKey)
	if err != nil {
		return err
	}

	return nil
}
