package multisig

import (
	"fmt"

	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/types/tx/signing"
)

// PubKeyMultisigThreshold implements a K of N threshold multisig.
type PubKeyMultisigThreshold struct {
	K       uint            `json:"threshold"`
	PubKeys []crypto.PubKey `json:"pubkeys"`
}

var _ PubKey = PubKeyMultisigThreshold{}

// NewPubKeyMultisigThreshold returns a new PubKeyMultisigThreshold.
// Panics if len(pubkeys) < k or 0 >= k.
func NewPubKeyMultisigThreshold(k int, pubkeys []crypto.PubKey) PubKey {
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

// VerifyBytes expects sig to be an amino encoded version of a MultiSignature.
// Returns true iff the multisignature contains k or more signatures
// for the correct corresponding keys,
// and all signatures are valid. (Not just k of the signatures)
// The multisig uses a bitarray, so multiple signatures for the same key is not
// a concern.
//
// NOTE: VerifyMultisignature should preferred to VerifyBytes which only works
// with amino multisignatures.
func (pk PubKeyMultisigThreshold) VerifyBytes(msg []byte, marshalledSig []byte) bool {
	var sig AminoMultisignature
	err := Cdc.UnmarshalBinaryBare(marshalledSig, &sig)
	if err != nil {
		return false
	}
	size := sig.BitArray.Count()
	// ensure bit array is the correct size
	if len(pk.PubKeys) != size {
		return false
	}
	// ensure size of signature list
	if len(sig.Sigs) < int(pk.K) || len(sig.Sigs) > size {
		return false
	}
	// ensure at least k signatures are set
	if sig.BitArray.NumTrueBitsBefore(size) < int(pk.K) {
		return false
	}
	// index in the list of signatures which we are concerned with.
	sigIndex := 0
	for i := 0; i < size; i++ {
		if sig.BitArray.GetIndex(i) {
			if !pk.PubKeys[i].VerifyBytes(msg, sig.Sigs[sigIndex]) {
				return false
			}
			sigIndex++
		}
	}
	return true
}

// VerifyMultisignature implements the PubKey.VerifyMultisignature method
func (pk PubKeyMultisigThreshold) VerifyMultisignature(getSignBytes GetSignBytesFunc, sig *signing.MultiSignatureData) error {
	bitarray := sig.BitArray
	sigs := sig.Signatures
	size := bitarray.Count()
	// ensure bit array is the correct size
	if len(pk.PubKeys) != size {
		return fmt.Errorf("bit array size is incorrect %d", len(pk.PubKeys))
	}
	// ensure size of signature list
	if len(sigs) < int(pk.K) || len(sigs) > size {
		return fmt.Errorf("signature size is incorrect %d", len(sigs))
	}
	// ensure at least k signatures are set
	if bitarray.NumTrueBitsBefore(size) < int(pk.K) {
		return fmt.Errorf("minimum number of signatures not set, have %d, expected %d", bitarray.NumTrueBitsBefore(size), int(pk.K))
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
				if !pk.PubKeys[i].VerifyBytes(msg, si.Signature) {
					return err
				}
			case *signing.MultiSignatureData:
				nestedMultisigPk, ok := pk.PubKeys[i].(PubKey)
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

// GetPubKeys implements the PubKey.GetPubKeys method
func (pk PubKeyMultisigThreshold) GetPubKeys() []crypto.PubKey {
	return pk.PubKeys
}

// Bytes returns the amino encoded version of the PubKeyMultisigThreshold
func (pk PubKeyMultisigThreshold) Bytes() []byte {
	return Cdc.MustMarshalBinaryBare(pk)
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

// GetThreshold implements the PubKey.GetThreshold method
func (pk PubKeyMultisigThreshold) GetThreshold() uint {
	return pk.K
}

func (pk PubKeyMultisigThreshold) Type() string { return "PubKeyMultisigThreshold" }
