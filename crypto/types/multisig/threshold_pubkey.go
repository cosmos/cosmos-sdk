package multisig

import (
	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/types/tx/signing"
)

// PubKeyMultisigThreshold implements a K of N threshold multisig.
type PubKeyMultisigThreshold struct {
	K       uint            `json:"threshold"`
	PubKeys []crypto.PubKey `json:"pubkeys"`
}

func (pk PubKeyMultisigThreshold) Threshold() uint {
	return pk.K
}

var _ MultisigPubKey = PubKeyMultisigThreshold{}

// NewPubKeyMultisigThreshold returns a new PubKeyMultisigThreshold.
// Panics if len(pubkeys) < k or 0 >= k.
func NewPubKeyMultisigThreshold(k int, pubkeys []crypto.PubKey) MultisigPubKey {
	if k <= 0 {
		panic("threshold k of n multisignature: k <= 0")
	}
	if len(pubkeys) < k {
		panic("threshold k of n multisignature: len(pubkeys) < k")
	}
	for _, pubkey := range pubkeys {
		if pubkey == nil {
			panic("nil pubkey")
		}
	}
	return PubKeyMultisigThreshold{uint(k), pubkeys}
}

// VerifyBytes should not be used with this PubKeyMultisigThreshold, instead VerifyMultisignature
// must be used
func (pk PubKeyMultisigThreshold) VerifyBytes([]byte, []byte) bool {
	return false
}

func (pk PubKeyMultisigThreshold) VerifyMultisignature(getSignBytes GetSignBytesFunc, sig *signing.MultiSignatureData) bool {
	bitarray := sig.BitArray
	sigs := sig.Signatures
	size := bitarray.Size()
	// ensure bit array is the correct size
	if len(pk.PubKeys) != size {
		return false
	}
	// ensure size of signature list
	if len(sigs) < int(pk.K) || len(sigs) > size {
		return false
	}
	// ensure at least k signatures are set
	if bitarray.NumTrueBitsBefore(size) < int(pk.K) {
		return false
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
					return false
				}
				if !pk.PubKeys[i].VerifyBytes(msg, si.Signature) {
					return false
				}
			case *signing.MultiSignatureData:
				nestedMultisigPk, ok := pk.PubKeys[i].(MultisigPubKey)
				if !ok {
					return false
				}
				if !nestedMultisigPk.VerifyMultisignature(getSignBytes, si) {
					return false
				}
			default:
				return false
			}
			sigIndex++
		}
	}
	return true
}

func (pk PubKeyMultisigThreshold) GetPubKeys() []crypto.PubKey {
	return pk.PubKeys
}

// Bytes returns the amino encoded version of the PubKeyMultisigThreshold
func (pk PubKeyMultisigThreshold) Bytes() []byte {
	return cdc.MustMarshalBinaryBare(pk)
}

// Address returns tmhash(PubKeyMultisigThreshold.Bytes())
func (pk PubKeyMultisigThreshold) Address() crypto.Address {
	return crypto.AddressHash(pk.Bytes())
}

// Equals returns true iff pk and other both have the same number of keys, and
// all constituent keys are the same, and in the same order.
func (pk PubKeyMultisigThreshold) Equals(other crypto.PubKey) bool {
	otherKey, sameType := other.(PubKeyMultisigThreshold)
	if !sameType {
		return false
	}
	if pk.K != otherKey.K || len(pk.PubKeys) != len(otherKey.PubKeys) {
		return false
	}
	for i := 0; i < len(pk.PubKeys); i++ {
		if !pk.PubKeys[i].Equals(otherKey.PubKeys[i]) {
			return false
		}
	}
	return true
}
