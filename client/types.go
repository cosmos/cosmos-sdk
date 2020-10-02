package client

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
)

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
	SignModeHandler() signing.SignModeHandler
}

// TxBuilder defines an interface which an application-defined concrete transaction
// type must implement. Namely, it must be able to set messages, generate
// signatures, and provide canonical bytes to sign over. The transaction must
// also know how to encode itself.
type TxBuilder interface {
	GetTx() sdk.Tx

	SetMsgs(msgs ...sdk.Msg) error
	SetSignatures(signatures ...signing.SignatureV2) error
	SetMemo(memo string)
	SetFeeAmount(amount sdk.Coins)
	SetGasLimit(limit uint64)
	SetTimeoutHeight(height uint64)
}
