package multisig

import (
	"fmt"

	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/crypto/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
)

// PubKey implements a K of N threshold multisig.
type PubKey struct {
	K       uint32          `json:"threshold"`
	PubKeys []crypto.PubKey `json:"pubkeys"`
}

var _ crypto.PubKey = PubKey{}

// NewPubKeyMultisigThreshold returns a new PubKeyMultisigThreshold.
// Panics if len(pubkeys) < k or 0 >= k.
func NewPubKeyMultisigThreshold(k uint32, pubkeys []crypto.PubKey) crypto.PubKey {
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
	return PubKey{uint(k), pubkeys}
}

// VerifyBytes expects sig to be an amino encoded version of a MultiSignature.
// Returns true iff the multisignature contains k or more signatures
// for the correct corresponding keys,
// and all signatures are valid. (Not just k of the signatures)
// The multisig uses a bitarray, so multiple signatures for the same key is not
// a concern.
func (pk PubKey) VerifyBytes(msg []byte, marshalledSig []byte) bool {
	var sig Multisignature
	err := cdc.UnmarshalBinaryBare(marshalledSig, &sig)
	if err != nil {
		return false
	}
	size := sig.BitArray.Size()
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

type DecodedMultisignature struct {
	ModeInfo   *txtypes.ModeInfo_Multi
	Signatures [][]byte
}

type GetSignBytesFunc func(single *txtypes.ModeInfo_Single) ([]byte, error)

type MultisigPubKey interface {
	VerifyMultisignature(getSignBytes GetSignBytesFunc, sig DecodedMultisignature) bool
}

func DecodeMultisignatures(bz []byte) ([][]byte, error) {
	multisig := types.MultiSignature{}
	err := multisig.Unmarshal(bz)
	if err != nil {
		return nil, err
	}
	if len(multisig.XXX_unrecognized) > 0 {
		return nil, fmt.Errorf("rejecting unrecognized fields found in MultiSignature")
	}
	return multisig.Sigs, nil
}

func (pk PubKey) VerifyMultisignature(getSignBytes GetSignBytesFunc, sig DecodedMultisignature) bool {
	bitarray := sig.ModeInfo.Bitarray
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
			mi := sig.ModeInfo.ModeInfos[sigIndex]
			switch mi := mi.Sum.(type) {
			case *txtypes.ModeInfo_Single_:
				msg, err := getSignBytes(mi.Single)
				if err != nil {
					return false
				}
				if !pk.PubKeys[i].VerifyBytes(msg, sigs[sigIndex]) {
					return false
				}
			case *txtypes.ModeInfo_Multi_:
				nestedMultisigPk, ok := pk.PubKeys[i].(MultisigPubKey)
				if !ok {
					return false
				}
				nestedSigs, err := DecodeMultisignatures(sigs[sigIndex])
				if err != nil {
					return false
				}
				if !nestedMultisigPk.VerifyMultisignature(getSignBytes, DecodedMultisignature{
					ModeInfo:   mi.Multi,
					Signatures: nestedSigs,
				}) {
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

// Bytes returns the amino encoded version of the PubKey
func (pk PubKey) Bytes() []byte {
	return cdc.MustMarshalBinaryBare(pk)
}

// Address returns tmhash(PubKey.Bytes())
func (pk PubKey) Address() crypto.Address {
	return crypto.AddressHash(pk.Bytes())
}

// Equals returns true iff pk and other both have the same number of keys, and
// all constituent keys are the same, and in the same order.
func (pk PubKey) Equals(other crypto.PubKey) bool {
	otherKey, sameType := other.(PubKey)
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
