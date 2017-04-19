package tx

import (
	"github.com/pkg/errors"
	crypto "github.com/tendermint/go-crypto"
	data "github.com/tendermint/go-wire/data"
)

// OneSig lets us wrap arbitrary data with a go-crypto signature
//
// TODO: rethink how we want to integrate this with KeyStore so it makes
// more sense (particularly the verify method)
type OneSig struct {
	Data data.Bytes
	Signed
}

var _ SigInner = &OneSig{}

func New(data []byte) Sig {
	return WrapSig(&OneSig{Data: data})
}

// SignBytes returns the original data passed into `NewSig`
func (s *OneSig) SignBytes() []byte {
	return s.Data
}

// Sign will add a signature and pubkey.
//
// Depending on the Signable, one may be able to call this multiple times for multisig
// Returns error if called with invalid data or too many times
func (s *OneSig) Sign(pubkey crypto.PubKey, sig crypto.Signature) error {
	if pubkey.Empty() || sig.Empty() {
		return errors.New("Signature or Key missing")
	}
	if !s.Sig.Empty() {
		return errors.New("Transaction can only be signed once")
	}

	// set the value once we are happy
	s.Signed = Signed{sig, pubkey}
	return nil
}

// Signers will return the public key(s) that signed if the signature
// is valid, or an error if there is any issue with the signature,
// including if there are no signatures
func (s *OneSig) Signers() ([]crypto.PubKey, error) {
	if s.Pubkey.Empty() || s.Sig.Empty() {
		return nil, errors.New("Never signed")
	}
	if !s.Pubkey.VerifyBytes(s.Data, s.Sig) {
		return nil, errors.New("Signature doesn't match")
	}
	return []crypto.PubKey{s.Pubkey}, nil
}
