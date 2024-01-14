package tx

import (
	txsigning "cosmossdk.io/x/tx/signing"
	sdk "github.com/cosmos/cosmos-sdk/types"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
)

// TxEncodingConfig defines an interface that contains transaction
// encoders and decoders
type TxEncodingConfig interface {
	TxEncoder() sdk.TxV2Encoder
	TxDecoder() sdk.TxV2Decoder
	TxJSONEncoder() sdk.TxV2Encoder
	TxJSONDecoder() sdk.TxV2Decoder
	MarshalSignatureJSON([]signingtypes.SignatureV2) ([]byte, error)
	UnmarshalSignatureJSON([]byte) ([]signingtypes.SignatureV2, error)
}

// TxConfig defines an interface a client can utilize to generate an
// application-defined concrete transaction type. The type returned must
// implement TxBuilder.
type TxConfig interface {
	TxEncodingConfig

	NewTxBuilder() TxBuilder
	WrapTxBuilder(sdk.Tx) (TxBuilder, error)
	SignModeHandler() *txsigning.HandlerMap
	SigningContext() *txsigning.Context
}
