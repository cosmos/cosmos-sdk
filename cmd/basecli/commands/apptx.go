package commands

import (
	"github.com/pkg/errors"

	crypto "github.com/tendermint/go-crypto"
	keys "github.com/tendermint/go-crypto/keys"
	wire "github.com/tendermint/go-wire"

	bc "github.com/tendermint/basecoin/types"
)

type AppTx struct {
	chainID string
	signers []crypto.PubKey
	Tx      *bc.AppTx
}

var _ keys.Signable = &AppTx{}

// SignBytes returned the unsigned bytes, needing a signature
func (s *AppTx) SignBytes() []byte {
	return s.Tx.SignBytes(s.chainID)
}

// Sign will add a signature and pubkey.
//
// Depending on the Signable, one may be able to call this multiple times for multisig
// Returns error if called with invalid data or too many times
func (s *AppTx) Sign(pubkey crypto.PubKey, sig crypto.Signature) error {
	if len(s.signers) > 0 {
		return errors.New("AppTx already signed")
	}
	s.Tx.SetSignature(sig)
	s.signers = []crypto.PubKey{pubkey}
	return nil
}

// Signers will return the public key(s) that signed if the signature
// is valid, or an error if there is any issue with the signature,
// including if there are no signatures
func (s *AppTx) Signers() ([]crypto.PubKey, error) {
	if len(s.signers) == 0 {
		return nil, errors.New("No signatures on AppTx")
	}
	return s.signers, nil
}

// TxBytes returns the transaction data as well as all signatures
// It should return an error if Sign was never called
func (s *AppTx) TxBytes() ([]byte, error) {
	// TODO: verify it is signed

	// Code and comment from: basecoin/cmd/commands/tx.go
	// Don't you hate having to do this?
	// How many times have I lost an hour over this trick?!
	txBytes := wire.BinaryBytes(bc.TxS{s.Tx})
	return txBytes, nil
}

// AddSigner sets address and pubkey info on the tx based on the key that
// will be used for signing
func (a *AppTx) AddSigner(pk crypto.PubKey) {
	// get addr if available
	var addr []byte
	if !pk.Empty() {
		addr = pk.Address()
	}

	// set the send address, and pubkey if needed
	in := &a.Tx.Input
	in.Address = addr
	if in.Sequence == 1 {
		in.PubKey = pk
	}
}

// TODO: this should really be in the basecoin.types SendTx,
// but that code is too ugly now, needs refactor..
func (a *AppTx) ValidateBasic() error {
	if a.chainID == "" {
		return errors.New("No chain-id specified")
	}
	in := a.Tx.Input
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
	return nil
}
