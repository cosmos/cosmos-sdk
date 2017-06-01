/*
package tx contains generic Signable implementations that can be used
by your application or tests to handle authentication needs.

It currently supports transaction data as opaque bytes and either single
or multiple private key signatures using straightforward algorithms.
It currently does not support N-of-M key share signing of other more
complex algorithms (although it would be great to add them).

You can create them with NewSig() and NewMultiSig(), and they fulfill
the keys.Signable interface. You can then .Wrap() them to create
a basecoin.Tx.
*/
package txs

import (
	crypto "github.com/tendermint/go-crypto"
	"github.com/tendermint/go-crypto/keys"
	"github.com/tendermint/go-wire/data"

	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/errors"
)

// Signed holds one signature of the data
type Signed struct {
	Sig    crypto.Signature
	Pubkey crypto.PubKey
}

func (s Signed) Empty() bool {
	return s.Sig.Empty() || s.Pubkey.Empty()
}

/**** Registration ****/

func init() {
	basecoin.TxMapper.
		RegisterImplementation(&OneSig{}, TypeSig, ByteSig).
		RegisterImplementation(&MultiSig{}, TypeMultiSig, ByteMultiSig)
}

/**** One Sig ****/

// OneSig lets us wrap arbitrary data with a go-crypto signature
type OneSig struct {
	Tx     basecoin.Tx `json:"tx"`
	Signed `json:"signature"`
}

var _ keys.Signable = &OneSig{}

func NewSig(tx basecoin.Tx) *OneSig {
	return &OneSig{Tx: tx}
}

func (s *OneSig) Wrap() basecoin.Tx {
	return basecoin.Tx{s}
}

func (s *OneSig) ValidateBasic() error {
	// TODO: VerifyBytes here, we do it in Signers?
	if s.Empty() || !s.Pubkey.VerifyBytes(s.SignBytes(), s.Sig) {
		return errors.Unauthorized()
	}
	return s.Tx.ValidateBasic()
}

// TxBytes returns the full data with signatures
func (s *OneSig) TxBytes() ([]byte, error) {
	return data.ToWire(s)
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
		return errors.MissingSignature()
	}
	if !s.Empty() {
		return errors.TooManySignatures()
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
		return nil, errors.MissingSignature()
	}
	if !s.Pubkey.VerifyBytes(s.SignBytes(), s.Sig) {
		return nil, errors.InvalidSignature()
	}
	return []crypto.PubKey{s.Pubkey}, nil
}

/**** MultiSig ****/

// MultiSig lets us wrap arbitrary data with a go-crypto signature
type MultiSig struct {
	Tx   basecoin.Tx `json:"tx"`
	Sigs []Signed    `json:"signatures"`
}

var _ keys.Signable = &MultiSig{}

func NewMulti(tx basecoin.Tx) *MultiSig {
	return &MultiSig{Tx: tx}
}

func (s *MultiSig) Wrap() basecoin.Tx {
	return basecoin.Tx{s}
}

func (s *MultiSig) ValidateBasic() error {
	// TODO: more efficient
	_, err := s.Signers()
	if err != nil {
		return err
	}
	return s.Tx.ValidateBasic()
}

// TxBytes returns the full data with signatures
func (s *MultiSig) TxBytes() ([]byte, error) {
	return data.ToWire(s)
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
		return errors.MissingSignature()
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
		return nil, errors.MissingSignature()
	}
	// verify all the signatures before returning them
	keys := make([]crypto.PubKey, len(s.Sigs))
	data := s.SignBytes()
	for i := range s.Sigs {
		ms := s.Sigs[i]
		if !ms.Pubkey.VerifyBytes(data, ms.Sig) {
			return nil, errors.InvalidSignature()
		}
		keys[i] = ms.Pubkey
	}

	return keys, nil
}
