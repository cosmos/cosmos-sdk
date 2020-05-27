package client

import (
	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/types"
)

type (
	// TxGenerator defines an interface a client can utilize to generate an
	// application-defined concrete transaction type. The type returned must
	// implement TxBuilder.
	TxGenerator interface {
		NewTx() TxBuilder
		NewFee() Fee
		NewSignature() Signature
		MarshalTx(tx types.Tx) ([]byte, error)
	}

	Fee interface {
		types.Fee
		SetGas(uint64)
		SetAmount(types.Coins)
	}

	Signature interface {
		types.Signature
		SetPubKey(crypto.PubKey) error
		SetSignature([]byte)
	}

	// TxBuilder defines an interface which an application-defined concrete transaction
	// type must implement. Namely, it must be able to set messages, generate
	// signatures, and provide canonical bytes to sign over. The transaction must
	// also know how to encode itself.
	TxBuilder interface {
		GetTx() types.Tx

		SetMsgs(...types.Msg) error
		GetSignatures() []types.Signature
		SetSignatures(...Signature) error
		GetFee() types.Fee
		SetFee(Fee) error
		GetMemo() string
		SetMemo(string)

		// CanonicalSignBytes returns the canonical sign bytes to sign over, given a
		// chain ID, along with an account and sequence number.
		CanonicalSignBytes(cid string, num, seq uint64) ([]byte, error)
	}
)
