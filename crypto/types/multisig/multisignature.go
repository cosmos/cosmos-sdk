package multisig

import (
	"fmt"
	"strings"

	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/crypto/types"
)

// Multisignature is used to represent the signature object used in the multisigs.
// Sigs is a list of signatures, sorted by corresponding index.
type Multisignature struct {
	BitArray *types.CompactBitArray
	Sigs     [][]byte
}

// NewMultisig returns a new Multisignature of size n.
func NewMultisig(n int) *Multisignature {
	// Default the signature list to have a capacity of two, since we can
	// expect that most multisigs will require multiple signers.
	return &Multisignature{types.NewCompactBitArray(n), make([][]byte, 0, 2)}
}

// GetIndex returns the index of pk in keys. Returns -1 if not found
func getIndex(pk crypto.PubKey, keys []crypto.PubKey) int {
	for i := 0; i < len(keys); i++ {
		if pk.Equals(keys[i]) {
			return i
		}
	}
	return -1
}

// AddSignature adds a signature to the multisig, at the corresponding index.
// If the signature already exists, replace it.
func (mSig *Multisignature) AddSignature(sig []byte, index int) {
	newSigIndex := mSig.BitArray.NumTrueBitsBefore(index)
	// Signature already exists, just replace the value there
	if mSig.BitArray.GetIndex(index) {
		mSig.Sigs[newSigIndex] = sig
		return
	}
	mSig.BitArray.SetIndex(index, true)
	// Optimization if the index is the greatest index
	if newSigIndex == len(mSig.Sigs) {
		mSig.Sigs = append(mSig.Sigs, sig)
		return
	}
	// Expand slice by one with a dummy element, move all elements after i
	// over by one, then place the new signature in that gap.
	mSig.Sigs = append(mSig.Sigs, make([]byte, 0))
	copy(mSig.Sigs[newSigIndex+1:], mSig.Sigs[newSigIndex:])
	mSig.Sigs[newSigIndex] = sig
}

// AddSignatureFromPubKey adds a signature to the multisig, at the index in
// keys corresponding to the provided pubkey.
func (mSig *Multisignature) AddSignatureFromPubKey(sig []byte, pubkey crypto.PubKey, keys []crypto.PubKey) error {
	index := getIndex(pubkey, keys)
	if index == -1 {
		keysStr := make([]string, len(keys))
		for i, k := range keys {
			keysStr[i] = fmt.Sprintf("%X", k.Bytes())
		}

		return fmt.Errorf("provided key %X doesn't exist in pubkeys: \n%s", pubkey.Bytes(), strings.Join(keysStr, "\n"))
	}

	mSig.AddSignature(sig, index)
	return nil
}

// Marshal the multisignature with amino
func (mSig *Multisignature) Marshal() []byte {
	return cdc.MustMarshalBinaryBare(mSig)
}
