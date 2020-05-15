package context

import (
	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/types"
)

// AccountRetriever defines the interfaces required by transactions to
// ensure an account exists and to be able to query for account fields necessary
// for signing.
type AccountRetriever interface {
	EnsureExists(addr types.AccAddress) error
	GetAccountNumberSequence(addr types.AccAddress) (uint64, uint64, error)
}

type (
	// TxGenerator defines an interface a client can utilize to generate an
	// application-defined concrete transaction type. The type returned must
	// implement ClientTx.
	TxGenerator interface {
		NewTx() ClientTx
		NewFee() ClientFee
		NewSignature() ClientSignature
		MarshalTx(tx ClientTx) ([]byte, error)
	}

	ClientFee interface {
		types.Fee
		SetGas(uint64)
		SetAmount(types.Coins)
	}

	ClientSignature interface {
		types.Signature
		SetPubKey(crypto.PubKey) error
		SetSignature([]byte)
	}

	// ClientTx defines an interface which an application-defined concrete transaction
	// type must implement. Namely, it must be able to set messages, generate
	// signatures, and provide canonical bytes to sign over. The transaction must
	// also know how to encode itself.
	ClientTx interface {
		types.Tx

		SetMsgs(...types.Msg) error
		GetSignatures() []types.Signature
		SetSignatures(...ClientSignature) error
		GetFee() types.Fee
		SetFee(ClientFee) error
		GetMemo() string
		SetMemo(string)

		// CanonicalSignBytes returns the canonical JSON bytes to sign over, given a
		// chain ID, along with an account and sequence number. The JSON encoding
		// ensures all field names adhere to their proto definition, default values
		// are omitted, and follows the JSON Canonical Form.
		CanonicalSignBytes(cid string, num, seq uint64) ([]byte, error)
	}
)
