/*
Package auth contains generic Credential implementations that can be used
by your application or tests to handle authentication needs.

It currently supports transaction data as opaque bytes and either single
or multiple private key signatures using straightforward algorithms.
It currently does not support N-of-M key share signing of other more
complex algorithms (although it would be great to add them).

This can be embedded in another structure along with the data to be
signed and easily allow you to build a custom Signable implementation.
Please see example usage of Credential.
*/
package auth

import (
	crypto "github.com/tendermint/go-crypto"

	"github.com/cosmos/cosmos-sdk/errors"
)

//////////////////////////////////////////
// Interface

// Credential can be combined with message data
// to create a keys.Signable
type Credential interface {
	Empty() bool
	Sign(pubkey crypto.PubKey, sig crypto.Signature) error
	Signers(signBytes []byte) ([]crypto.PubKey, error)
}

// Signable is data along with credentials, which can be verified
type Signable interface {
	Signers() ([]crypto.PubKey, error)
}

/////////////////////////////////////////
// NamedSig - one signature

// NamedSig holds one signature of the data
type NamedSig struct {
	Sig    crypto.Signature
	Pubkey crypto.PubKey
}

var _ Credential = &NamedSig{}

func NewSig() *NamedSig {
	return new(NamedSig)
}

// Empty returns true if there is not enough signature info
func (s *NamedSig) Empty() bool {
	return s.Sig.Empty() || s.Pubkey.Empty()
}

// Sign will add a signature and pubkey.
func (s *NamedSig) Sign(pubkey crypto.PubKey, sig crypto.Signature) error {
	if !s.Empty() {
		return ErrTooManySignatures()
	}
	s.Sig = sig
	s.Pubkey = pubkey
	if s.Empty() {
		return errors.ErrMissingSignature()
	}
	return nil
}

// signer will return a pubkey and a possible error.
// building block to combine
func (s *NamedSig) signer(signBytes []byte) (crypto.PubKey, error) {
	key := s.Pubkey
	if s.Empty() {
		return key, errors.ErrMissingSignature()
	}
	if !s.Pubkey.VerifyBytes(signBytes, s.Sig) {
		return key, ErrInvalidSignature()
	}
	return key, nil
}

// Signers will return the public key that signed if the signature
// is valid, or an error if there is any issue with the signature,
// including if there are no signatures
func (s *NamedSig) Signers(signBytes []byte) ([]crypto.PubKey, error) {
	key, err := s.signer(signBytes)
	if err != nil {
		return nil, err
	}
	return []crypto.PubKey{key}, nil
}

// // TxBytes returns the full data with signatures
// func (s *OneSig) TxBytes() ([]byte, error) {
// 	return data.ToWire(s.Wrap())
// }

// // SignBytes returns the original data passed into `NewSig`
// func (s *OneSig) SignBytes() []byte {
// 	res, err := data.ToWire(s.Tx)
// 	if err != nil {
// 		panic(err)
// 	}
// 	return res
// }

/////////////////////////////////////////
// NamedSigs - multiple signatures

// NamedSigs is a list of signatures
// and fulfils the same interface as NamedSig
type NamedSigs []NamedSig

var _ Credential = &NamedSigs{}

func NewMultiSig() *NamedSigs {
	// pre-allocate space of two, as we expect multiple signatures
	s := make(NamedSigs, 0, 2)
	return &s
}

// Empty returns true iff no signatures were ever added
func (s *NamedSigs) Empty() bool {
	return len(*s) == 0
}

// Sign will add a signature and pubkey.
//
// Depending on the Signable, one may be able to call this multiple times for multisig
// Returns error if called with invalid data or too many times
func (s *NamedSigs) Sign(pubkey crypto.PubKey, sig crypto.Signature) error {
	// optimize for success case - append and store signature
	l := len(*s)
	*s = append(*s, NamedSig{})
	err := (*s)[l].Sign(pubkey, sig)

	// if there is an error, remove from the list
	if err != nil {
		*s = (*s)[:l]
	}
	return err
}

// Signers will return the public key(s) that signed if the signature
// is valid, or an error if there is any issue with the signature,
// including if there are no signatures
func (s *NamedSigs) Signers(signBytes []byte) (res []crypto.PubKey, err error) {
	if s.Empty() {
		return nil, errors.ErrMissingSignature()
	}

	l := len(*s)
	res = make([]crypto.PubKey, l)
	for i := 0; i < l; i++ {
		res[i], err = (*s)[i].signer(signBytes)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

// Sign - sign the given data with private key and store
// the result in the credentil
func Sign(msg []byte, key crypto.PrivKey, cred Credential) error {
	pubkey := key.PubKey()
	sig := key.Sign(msg)
	return cred.Sign(pubkey, sig)
}
