package types

import (
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/gogo/protobuf/proto"

	"github.com/tendermint/tendermint/crypto"
)

type (
	// Msg defines the interface a transaction message must fulfill.
	Msg interface {
		proto.Message

		// Return the message type.
		// Must be alphanumeric or empty.
		Route() string

		// Returns a human-readable string for the message, intended for utilization
		// within tags
		Type() string

		// ValidateBasic does a simple validation check that
		// doesn't require access to any other information.
		ValidateBasic() error

		// Get the canonical byte representation of the Msg.
		GetSignBytes() []byte

		// Signers returns the addrs of signers that must sign.
		// CONTRACT: All signatures must be present to be valid.
		// CONTRACT: Returns addrs in some deterministic order.
		GetSigners() []AccAddress
	}

	// Fee defines an interface for an application application-defined concrete
	// transaction type to be able to set and return the transaction fee.
	Fee interface {
		GetGas() uint64
		GetAmount() Coins
	}

	// Signature defines an interface for an application application-defined
	// concrete transaction type to be able to set and return transaction signatures.
	Signature interface {
		GetPubKey() crypto.PubKey
		GetSignature() []byte
	}

	// Tx defines the interface a transaction must fulfill.
	Tx interface {
		// Gets the all the transaction's messages.
		GetMsgs() []Msg

		// ValidateBasic does a simple and lightweight validation check that doesn't
		// require access to any other information.
		ValidateBasic() error
	}

	// FeeTx defines the interface to be implemented by Tx to use the FeeDecorators
	FeeTx interface {
		Tx
		GetGas() uint64
		GetFee() Coins
		FeePayer() AccAddress
	}

	// Tx must have GetMemo() method to use ValidateMemoDecorator
	TxWithMemo interface {
		Tx
		GetMemo() string // TODO: check if we can move it to `Tx`
	}

	// TxWithTimeoutHeight extends the Tx interface by allowing a transaction to
	// set a height timeout.
	TxWithTimeoutHeight interface {
		Tx
		GetTimeoutHeight() uint64
	}

	// SigVerifiableTx defines a transaction interface for all signature verification
	// handlers. It extends the sdk.Tx type.
	// FIXME: moved from x/auth/signing/sig_verifiable_tx.go
	SigVerifiableTx interface {
		Tx
		GetSigners() []AccAddress
		GetPubKeys() []crypto.PubKey // If signer already has pubkey in context, this list will have nil in its place
		GetSignaturesV2() ([]signing.SignatureV2, error)
	}

	// Tx defines a transaction interface that supports all standard message, signature
	// fee, memo, and auxiliary interfaces.
	// It extends the sdk.Tx type (through tx.WithMemo)
	// FIXME: moved from x/auth/signing/sig_verifiable_tx.go
	FullTx interface {
		SigVerifiableTx

		TxWithMemo
		FeeTx
		TxWithTimeoutHeight
	}
)

// TxDecoder unmarshals transaction bytes
type TxDecoder func(txBytes []byte) (Tx, error)

// TxEncoder marshals transaction to bytes
type TxEncoder func(tx Tx) ([]byte, error)
