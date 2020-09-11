package multisig

import (
	fmt "fmt"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/cosmos/cosmos-sdk/codec/types"
	crypto "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	tmcrypto "github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/crypto/types/multisig"
)

var cdc = codec.NewProtoCodec(types.NewInterfaceRegistry())

var _ multisig.PubKey = &LegacyAminoMultisigThresholdPubKey{}

// Address implements crypto.PubKey Address method
func (m *LegacyAminoMultisigThresholdPubKey) Address() crypto.Address {
	return tmcrypto.AddressHash(m.Bytes())
}

// Bytes returns the proto encoded version of the LegacyAminoMultisigThresholdPubKey
func (m *LegacyAminoMultisigThresholdPubKey) Bytes() []byte {
	return cdc.MustMarshalBinaryBare(m)
}

// VerifyMultisignature implements the multisig.PubKey VerifyMultisignature method
func (m *LegacyAminoMultisigThresholdPubKey) VerifyMultisignature(getSignBytes multisig.GetSignBytesFunc, sig *signing.MultiSignatureData) error {
	bitarray := sig.BitArray
	sigs := sig.Signatures
	size := bitarray.Count()
	pubKeys := m.GetPubKeys()
	// ensure bit array is the correct size
	if len(pubKeys) != size {
		return fmt.Errorf("bit array size is incorrect %d", len(pubKeys))
	}
	// ensure size of signature list
	if len(sigs) < int(m.K) || len(sigs) > size {
		return fmt.Errorf("signature size is incorrect %d", len(sigs))
	}
	// ensure at least k signatures are set
	if bitarray.NumTrueBitsBefore(size) < int(m.K) {
		return fmt.Errorf("minimum number of signatures not set, have %d, expected %d", bitarray.NumTrueBitsBefore(size), int(m.K))
	}
	// index in the list of signatures which we are concerned with.
	sigIndex := 0
	for i := 0; i < size; i++ {
		if bitarray.GetIndex(i) {
			si := sig.Signatures[sigIndex]
			switch si := si.(type) {
			case *signing.SingleSignatureData:
				msg, err := getSignBytes(si.SignMode)
				if err != nil {
					return err
				}
				if !pubKeys[i].VerifySignature(msg, si.Signature) {
					return err
				}
			case *signing.MultiSignatureData:
				nestedMultisigPk, ok := pubKeys[i].(multisig.PubKey)
				if !ok {
					return fmt.Errorf("unable to parse pubkey of index %d", i)
				}
				if err := nestedMultisigPk.VerifyMultisignature(getSignBytes, si); err != nil {
					return err
				}
			default:
				return fmt.Errorf("improper signature data type for index %d", sigIndex)
			}
			sigIndex++
		}
	}
	return nil
}

// VerifySignature implements crypto.PubKey VerifySignature method,
// it panics because it can't handle MultiSignatureData
// cf. https://github.com/cosmos/cosmos-sdk/issues/7109#issuecomment-686329936
func (m *LegacyAminoMultisigThresholdPubKey) VerifySignature(msg []byte, sig []byte) bool {
	panic("not implemented")
}

// GetPubKeys implements the PubKey.GetPubKeys method
func (m *LegacyAminoMultisigThresholdPubKey) GetPubKeys() []tmcrypto.PubKey {
	if m != nil {
		pubKeys := make([]tmcrypto.PubKey, len(m.PubKeys))
		for i := 0; i < len(m.PubKeys); i++ {
			pubKeys[i] = m.PubKeys[i].GetCachedValue().(tmcrypto.PubKey)
		}
		return pubKeys
	}

	return nil
}

// Equals returns true if m and other both have the same number of keys, and
// all constituent keys are the same, and in the same order.
func (m *LegacyAminoMultisigThresholdPubKey) Equals(key tmcrypto.PubKey) bool {
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
func (m *LegacyAminoMultisigThresholdPubKey) GetThreshold() uint {
	return uint(m.K)
}

// Type returns multisig type
func (m *LegacyAminoMultisigThresholdPubKey) Type() string {
	return "PubKeyMultisigThreshold"
}
