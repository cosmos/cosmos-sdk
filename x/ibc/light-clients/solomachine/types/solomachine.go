package types

import (
	"github.com/tendermint/tendermint/crypto"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
)

var _, _, _, _ codectypes.UnpackInterfacesMessage = &ClientState{}, &ConsensusState{}, &Header{}, &HeaderData{}

// UnpackInterfaces implements the UnpackInterfaceMessages.UnpackInterfaces method
func (cs *ClientState) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	return cs.ConsensusState.UnpackInterfaces(unpacker)
}

// UnpackInterfaces implements the UnpackInterfaceMessages.UnpackInterfaces method
func (cs *ConsensusState) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var pubKey crypto.PubKey
	err := unpacker.UnpackAny(cs.PublicKey, &pubKey)
	if err != nil {
		return err
	}

	return nil
}

// UnpackInterfaces implements the UnpackInterfaceMessages.UnpackInterfaces method
func (h *Header) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var pubKey crypto.PubKey
	err := unpacker.UnpackAny(h.NewPublicKey, &pubKey)
	if err != nil {
		return err
	}

	return nil
}

// UnpackInterfaces implements the UnpackInterfaceMessages.UnpackInterfaces method
func (hd *HeaderData) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var pubKey crypto.PubKey
	err := unpacker.UnpackAny(hd.NewPubKey, &pubKey)
	if err != nil {
		return err
	}

	return nil
}
