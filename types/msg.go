package types

import crypto "github.com/tendermint/go-crypto"

// The parsed tx bytes is called a Msg.
type Msg interface {
	Get(key interface{}) (value interface{})
	Origin() (tx []byte)

	// Signers() returns the crypto.PubKey of signers
	// responsible for signing the Msg.
	// CONTRACT: All signatures must be present to be valid.
	// CONTRACT: Returns pubkeys in some deterministic order
	// CONTRACT: Get(MsgKeySigners) compatible.
	Signers() []crypto.PubKey

	// Signatures() returns the crypto.Signature of sigenrs
	// who signed the Msg.
	// CONTRACT: Length returned is same as length of
	// pubkeys returned from MsgKeySigners, and the order
	// matches.
	// CONTRACT: If the signature is missing (ie the Msg is
	// invalid), then the corresponding signature is
	// .Empty().
	// CONTRACT: Get(MsgKeySignatures) compatible.
	Signatures() []crypto.Signature
}
