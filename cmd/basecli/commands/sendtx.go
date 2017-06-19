package commands

import (
	"github.com/pkg/errors"
	bc "github.com/tendermint/basecoin/types"
	crypto "github.com/tendermint/go-crypto"
	keys "github.com/tendermint/go-crypto/keys"
	wire "github.com/tendermint/go-wire"
)

type SendTx struct {
	chainID string
	signers []crypto.PubKey
	Tx      *bc.SendTx
}

var _ keys.Signable = &SendTx{}

// SignBytes returned the unsigned bytes, needing a signature
func (s *SendTx) SignBytes() []byte {
	return s.Tx.SignBytes(s.chainID)
}

// Sign will add a signature and pubkey.
//
// Depending on the Signable, one may be able to call this multiple times for multisig
// Returns error if called with invalid data or too many times
func (s *SendTx) Sign(pubkey crypto.PubKey, sig crypto.Signature) error {
	addr := pubkey.Address()
	set := s.Tx.SetSignature(addr, sig)
	if !set {
		return errors.Errorf("Cannot add signature for address %X", addr)
	}
	s.signers = append(s.signers, pubkey)
	return nil
}

// Signers will return the public key(s) that signed if the signature
// is valid, or an error if there is any issue with the signature,
// including if there are no signatures
func (s *SendTx) Signers() ([]crypto.PubKey, error) {
	if len(s.signers) == 0 {
		return nil, errors.New("No signatures on SendTx")
	}
	return s.signers, nil
}

// TxBytes returns the transaction data as well as all signatures
// It should return an error if Sign was never called
func (s *SendTx) TxBytes() ([]byte, error) {
	// TODO: verify it is signed

	// Code and comment from: basecoin/cmd/commands/tx.go
	// Don't you hate having to do this?
	// How many times have I lost an hour over this trick?!
	txBytes := wire.BinaryBytes(struct {
		bc.Tx `json:"unwrap"`
	}{s.Tx})
	return txBytes, nil
}

// AddSigner sets address and pubkey info on the tx based on the key that
// will be used for signing
func (s *SendTx) AddSigner(pk crypto.PubKey) {
	// get addr if available
	var addr []byte
	if !pk.Empty() {
		addr = pk.Address()
	}

	// set the send address, and pubkey if needed
	in := s.Tx.Inputs
	in[0].Address = addr
	if in[0].Sequence == 1 {
		in[0].PubKey = pk
	}
}

// TODO: this should really be in the basecoin.types SendTx,
// but that code is too ugly now, needs refactor..
func (s *SendTx) ValidateBasic() error {
	if s.chainID == "" {
		return errors.New("No chain-id specified")
	}
	for _, in := range s.Tx.Inputs {
		if len(in.Address) != 20 {
			return errors.Errorf("Invalid input address length: %d", len(in.Address))
		}
		if !in.Coins.IsValid() {
			return errors.Errorf("Invalid input coins %v", in.Coins)
		}
		if in.Coins.IsZero() {
			return errors.New("Input coins cannot be zero")
		}
		if in.Sequence <= 0 {
			return errors.New("Sequence must be greater than 0")
		}
	}
	for _, out := range s.Tx.Outputs {
		// we now allow chain/addr, so it can be more than 20 bytes
		if len(out.Address) < 20 {
			return errors.Errorf("Invalid output address length: %d", len(out.Address))
		}
		if !out.Coins.IsValid() {
			return errors.Errorf("Invalid output coins %v", out.Coins)
		}
		if out.Coins.IsZero() {
			return errors.New("Output coins cannot be zero")
		}
	}

	return nil
}
