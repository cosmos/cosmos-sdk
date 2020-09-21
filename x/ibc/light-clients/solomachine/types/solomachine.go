package types

import (
	"github.com/tendermint/tendermint/crypto"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
)

// Interface implementation checks.
var _, _, _, _ codectypes.UnpackInterfacesMessage = &ClientState{}, &ConsensusState{}, &Header{}, &HeaderData{}

// UnpackInterfaces implements the UnpackInterfaceMessages.UnpackInterfaces method
func (cs ClientState) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	return cs.ConsensusState.UnpackInterfaces(unpacker)
}

// UnpackInterfaces implements the UnpackInterfaceMessages.UnpackInterfaces method
func (cs ConsensusState) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	return unpacker.UnpackAny(cs.PublicKey, new(crypto.PubKey))
}

// UnpackInterfaces implements the UnpackInterfaceMessages.UnpackInterfaces method
func (h Header) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	return unpacker.UnpackAny(h.NewPublicKey, new(crypto.PubKey))
}

// UnpackInterfaces implements the UnpackInterfaceMessages.UnpackInterfaces method
func (hd HeaderData) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	return unpacker.UnpackAny(hd.NewPubKey, new(crypto.PubKey))
}
