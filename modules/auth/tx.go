/*
Package auth contains generic Signable implementations that can be used
by your application or tests to handle authentication needs.

It currently supports transaction data as opaque bytes and either single
or multiple private key signatures using straightforward algorithms.
It currently does not support N-of-M key share signing of other more
complex algorithms (although it would be great to add them).

You can create them with NewSig() and NewMultiSig(), and they fulfill
the keys.Signable interface. You can then .Wrap() them to create
a basecoin.Tx.
*/
package auth

import (
	crypto "github.com/tendermint/go-crypto"
	"github.com/tendermint/go-crypto/keys"
	"github.com/tendermint/go-wire/data"

	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/errors"
)

// nolint
const (
	// for signatures
	ByteSingleTx = 0x16
	ByteMultiSig = 0x17
)

// nolint
const (
	// for signatures
	TypeSingleTx = NameSigs + "/one"
	TypeMultiSig = NameSigs + "/multi"
)

// Signed holds one signature of the data
type Signed struct {
	Sig    crypto.Signature
	Pubkey crypto.PubKey
}

// Empty returns true if there is not enough signature info
func (s Signed) Empty() bool {
	return s.Sig.Empty() || s.Pubkey.Empty()
}

/**** Registration ****/

func init() {
	basecoin.TxMapper.
		RegisterImplementation(&OneSig{}, TypeSingleTx, ByteSingleTx).
		RegisterImplementation(&MultiSig{}, TypeMultiSig, ByteMultiSig)
}

/**** One Sig ****/

// OneSig lets us wrap arbitrary data with a go-crypto signature
type OneSig struct {
	Tx     basecoin.Tx `json:"tx"`
	Signed `json:"signature"`
}

//Interfaces to fulfill
var _ keys.Signable = &OneSig{}
var _ basecoin.TxLayer = &OneSig{}

// NewSig wraps the tx with a Signable that accepts exactly one signature
func NewSig(tx basecoin.Tx) *OneSig {
	return &OneSig{Tx: tx}
}

//nolint
func (s *OneSig) Wrap() basecoin.Tx {
	return basecoin.Tx{s}
}
func (s *OneSig) Next() basecoin.Tx {
	return s.Tx
}
func (s *OneSig) ValidateBasic() error {
	return s.Tx.ValidateBasic()
}

// TxBytes returns the full data with signatures
func (s *OneSig) TxBytes() ([]byte, error) {
	return data.ToWire(s.Wrap())
}

// SignBytes returns the original data passed into `NewSig`
func (s *OneSig) SignBytes() []byte {
	res, err := data.ToWire(s.Tx)
	if err != nil {
		panic(err)
	}
	return res
}

// Sign will add a signature and pubkey.
//
// Depending on the Signable, one may be able to call this multiple times for multisig
// Returns error if called with invalid data or too many times
func (s *OneSig) Sign(pubkey crypto.PubKey, sig crypto.Signature) error {
	signed := Signed{sig, pubkey}
	if signed.Empty() {
		return errors.ErrMissingSignature()
	}
	if !s.Empty() {
		return errors.ErrTooManySignatures()
	}
	// set the value once we are happy
	s.Signed = signed
	return nil
}

// Signers will return the public key(s) that signed if the signature
// is valid, or an error if there is any issue with the signature,
// including if there are no signatures
func (s *OneSig) Signers() ([]crypto.PubKey, error) {
	if s.Empty() {
		return nil, errors.ErrMissingSignature()
	}
	if !s.Pubkey.VerifyBytes(s.SignBytes(), s.Sig) {
		return nil, errors.ErrInvalidSignature()
	}
	return []crypto.PubKey{s.Pubkey}, nil
}

/**** MultiSig ****/

// MultiSig lets us wrap arbitrary data with a go-crypto signature
type MultiSig struct {
	Tx   basecoin.Tx `json:"tx"`
	Sigs []Signed    `json:"signatures"`
}

//Interfaces to fulfill
var _ keys.Signable = &MultiSig{}
var _ basecoin.TxLayer = &MultiSig{}

// NewMulti wraps the tx with a Signable that accepts arbitrary numbers of signatures
func NewMulti(tx basecoin.Tx) *MultiSig {
	return &MultiSig{Tx: tx}
}

// nolint
func (s *MultiSig) Wrap() basecoin.Tx {
	return basecoin.Tx{s}
}
func (s *MultiSig) Next() basecoin.Tx {
	return s.Tx
}
func (s *MultiSig) ValidateBasic() error {
	return s.Tx.ValidateBasic()
}

// TxBytes returns the full data with signatures
func (s *MultiSig) TxBytes() ([]byte, error) {
	return data.ToWire(s.Wrap())
}

// SignBytes returns the original data passed into `NewSig`
func (s *MultiSig) SignBytes() []byte {
	res, err := data.ToWire(s.Tx)
	if err != nil {
		panic(err)
	}
	return res
}

// Sign will add a signature and pubkey.
//
// Depending on the Signable, one may be able to call this multiple times for multisig
// Returns error if called with invalid data or too many times
func (s *MultiSig) Sign(pubkey crypto.PubKey, sig crypto.Signature) error {
	signed := Signed{sig, pubkey}
	if signed.Empty() {
		return errors.ErrMissingSignature()
	}
	// set the value once we are happy
	s.Sigs = append(s.Sigs, signed)
	return nil
}

// Signers will return the public key(s) that signed if the signature
// is valid, or an error if there is any issue with the signature,
// including if there are no signatures
func (s *MultiSig) Signers() ([]crypto.PubKey, error) {
	if len(s.Sigs) == 0 {
		return nil, errors.ErrMissingSignature()
	}
	// verify all the signatures before returning them
	keys := make([]crypto.PubKey, len(s.Sigs))
	data := s.SignBytes()
	for i := range s.Sigs {
		ms := s.Sigs[i]
		if !ms.Pubkey.VerifyBytes(data, ms.Sig) {
			return nil, errors.ErrInvalidSignature()
		}
		keys[i] = ms.Pubkey
	}

	return keys, nil
}

// Sign - sign the transaction with private key
func Sign(tx keys.Signable, key crypto.PrivKey) error {
	msg := tx.SignBytes()
	pubkey := key.PubKey()
	sig := key.Sign(msg)
	return tx.Sign(pubkey, sig)
}
