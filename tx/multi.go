package tx

import (
	"github.com/pkg/errors"
	crypto "github.com/tendermint/go-crypto"
	data "github.com/tendermint/go-data"
)

// MultiSig lets us wrap arbitrary data with a go-crypto signature
//
// TODO: rethink how we want to integrate this with KeyStore so it makes
// more sense (particularly the verify method)
type MultiSig struct {
	Data data.Bytes
	Sigs []Signed
}

type Signed struct {
	Sig    crypto.SignatureS
	Pubkey crypto.PubKeyS
}

var _ SigInner = &MultiSig{}

func NewMulti(data []byte) Sig {
	return Sig{&MultiSig{Data: data}}
}

// SignBytes returns the original data passed into `NewSig`
func (s *MultiSig) SignBytes() []byte {
	return s.Data
}

// Sign will add a signature and pubkey.
//
// Depending on the Signable, one may be able to call this multiple times for multisig
// Returns error if called with invalid data or too many times
func (s *MultiSig) Sign(pubkey crypto.PubKey, sig crypto.Signature) error {
	if pubkey == nil || sig == nil {
		return errors.New("Signature or Key missing")
	}

	// set the value once we are happy
	x := Signed{crypto.SignatureS{sig}, crypto.PubKeyS{pubkey}}
	s.Sigs = append(s.Sigs, x)
	return nil
}

// Signers will return the public key(s) that signed if the signature
// is valid, or an error if there is any issue with the signature,
// including if there are no signatures
func (s *MultiSig) Signers() ([]crypto.PubKey, error) {
	if len(s.Sigs) == 0 {
		return nil, errors.New("Never signed")
	}

	keys := make([]crypto.PubKey, len(s.Sigs))
	for i := range s.Sigs {
		ms := s.Sigs[i]
		if !ms.Pubkey.VerifyBytes(s.Data, ms.Sig) {
			return nil, errors.Errorf("Signature %d doesn't match (key: %X)", i, ms.Pubkey.Bytes())
		}
		keys[i] = ms.Pubkey
	}

	return keys, nil
}
