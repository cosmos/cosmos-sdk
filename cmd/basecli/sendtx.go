package main

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
