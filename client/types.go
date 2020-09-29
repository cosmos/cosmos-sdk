package client

import (
	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
)

// SigVerifiableTx defines a transaction interface for all signature verification
// handlers.
type SigVerifiableTx interface {
	types.Tx
	GetSigners() []types.AccAddress
	GetPubKeys() []crypto.PubKey // If signer already has pubkey in context, this list will have nil in its place
	GetSignaturesV2() ([]signing.SignatureV2, error)
}

// Tx defines a transaction interface that supports all standard message, signature
// fee, memo, and auxiliary interfaces.
type Tx interface {
	SigVerifiableTx

	types.TxWithMemo
	types.FeeTx
	types.TxWithTimeoutHeight
}

// TxEncodingConfig defines an interface that contains transaction
// encoders and decoders
type TxEncodingConfig interface {
	TxEncoder() sdk.TxEncoder
	TxDecoder() sdk.TxDecoder
	TxJSONEncoder() sdk.TxEncoder
	TxJSONDecoder() sdk.TxDecoder
	MarshalSignatureJSON([]signing.SignatureV2) ([]byte, error)
	UnmarshalSignatureJSON([]byte) ([]signing.SignatureV2, error)
}

// TxConfig defines an interface a client can utilize to generate an
// application-defined concrete transaction type. The type returned must
// implement TxBuilder.
type TxConfig interface {
	TxEncodingConfig

	NewTxBuilder() TxBuilder
	WrapTxBuilder(sdk.Tx) (TxBuilder, error)
	SignModeHandler() SignModeHandler
}

// SignModeHandler defines a interface to be implemented by types which will handle
// SignMode's by generating sign bytes from a Tx and SignerData
type SignModeHandler interface {
	// DefaultMode is the default mode that is to be used with this handler if no
	// other mode is specified. This can be useful for testing and CLI usage
	DefaultMode() signing.SignMode

	// Modes is the list of modes supporting by this handler
	Modes() []signing.SignMode

	// GetSignBytes returns the sign bytes for the provided SignMode, SignerData and Tx,
	// or an error
	GetSignBytes(mode signing.SignMode, data SignerData, tx sdk.Tx) ([]byte, error)
}

// TxBuilder defines an interface which an application-defined concrete transaction
// type must implement. Namely, it must be able to set messages, generate
// signatures, and provide canonical bytes to sign over. The transaction must
// also know how to encode itself.
type TxBuilder interface {
	GetTx() Tx

	SetMsgs(msgs ...sdk.Msg) error
	SetSignatures(signatures ...signing.SignatureV2) error
	SetMemo(memo string)
	SetFeeAmount(amount sdk.Coins)
	SetGasLimit(limit uint64)
	SetTimeoutHeight(height uint64)
}

// SignerData is the specific information needed to sign a transaction that generally
// isn't included in the transaction body itself
type SignerData struct {
	// ChainID is the chain that this transaction is targeted
	ChainID string

	// AccountNumber is the account number of the signer
	AccountNumber uint64

	// Sequence is the account sequence number of the signer that is used
	// for replay protection. This field is only useful for Legacy Amino signing,
	// since in SIGN_MODE_DIRECT the account sequence is already in the signer
	// info.
	Sequence uint64
}
