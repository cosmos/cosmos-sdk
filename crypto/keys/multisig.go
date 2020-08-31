package keys

import (
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	tmcrypto "github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/crypto/types/multisig"
)

var cdc = codec.NewProtoCodec(types.NewInterfaceRegistry())

var _ multisig.PubKey = &MultisigThresholdPubKey{}

// Address implements crypto.PubKey Address method
// TODO
func (m *MultisigThresholdPubKey) Address() crypto.Address {
	return nil
}

// Bytes returns the proto encoded version of the MultisigThresholdPubKey
func (m *MultisigThresholdPubKey) Bytes() []byte {
	return cdc.MustMarshalBinaryBare(m)
}

// VerifyMultisignature implements the multisig.PubKey VerifyMultisignature method
// TODO
func (m *MultisigThresholdPubKey) VerifyMultisignature(getSignBytes multisig.GetSignBytesFunc, sig *signing.MultiSignatureData) error {
	return nil
}

// VerifySignature implements crypto.PubKey VerifySignature method
func (m *MultisigThresholdPubKey) VerifySignature(msg []byte, sig []byte) bool {
	// var sigData signing.SignatureDescriptor_Data_Multi // is this expected?
	// err := cdc.UnmarshalBinaryBare(sig, &sigData)
	// if err != nil {
	// 	return false
	// }
	// size := sigData.Bitarray.Count()
	// // ensure bit array is the correct size
	// if len(m.PubKeys) != size {
	// 	return false
	// }
	// // ensure size of signature list
	// if len(sigData.Signatures) < int(m.K) || len(sigData.Signatures) > size {
	// 	return false
	// }
	// // ensure at least k signatures are set
	// if sigData.Bitarray.NumTrueBitsBefore(size) < int(m.K) {
	// 	return false
	// }
	// // index in the list of signatures which we are concerned with.
	// sigIndex := 0
	// for i := 0; i < size; i++ {
	// 	if sigData.Bitarray.GetIndex(i) {
	// 		pk, ok := m.PubKeys[i].GetCachedValue().(tmcrypto.PubKey)
	// 		if !ok {
	// 			return false
	// 		}
	// 		if !pk.VerifySignature(msg, sigData.Signatures[sigIndex]) { // TODO fix
	// 			return false
	// 		}
	// 		sigIndex++
	// 	}
	// }
	return true
}

// GetPubKeys implements the PubKey.GetPubKeys method
func (m *MultisigThresholdPubKey) GetPubKeys() []crypto.PubKey {
	pubKeys := make([]crypto.PubKey, len(m.PubKeys))
	for i := 0; i < len(m.PubKeys); i++ {
		pubKeys[i] = m.PubKeys[i].GetCachedValue().(tmcrypto.PubKey)
	}
	return pubKeys
}

// Equals returns true if m and other both have the same number of keys, and
// all constituent keys are the same, and in the same order.
func (m *MultisigThresholdPubKey) Equals(key crypto.PubKey) bool {
	otherKey, ok := key.(multisig.PubKey)
	if !ok {
		return false
	}
	pubKeys := m.GetPubKeys()
	otherPubKeys := otherKey.GetPubKeys()
	if m.GetThreshold() != otherKey.GetThreshold() || len(pubKeys) != len(otherPubKeys) {
		return false
	}

	for i := 0; i < len(pubKeys); i++ {
		if !pubKeys[i].Equals(otherPubKeys[i]) {
			return false
		}
	}
	return true
}

// GetThreshold implements the PubKey.GetThreshold method
func (m *MultisigThresholdPubKey) GetThreshold() uint {
	return uint(m.K)
}

// Type returns multisig type
func (m *MultisigThresholdPubKey) Type() string {
	return "PubKeyMultisigThreshold"
}
