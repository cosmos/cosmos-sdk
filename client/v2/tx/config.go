package tx

import (
	base "cosmossdk.io/api/cosmos/base/v1beta1"
	apitxsigning "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	"cosmossdk.io/x/tx/signing"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TxConfig defines an interface a client can utilize to generate an
// application-defined concrete transaction type. The type returned must
// implement TxBuilder.
type TxConfig interface {
	TxEncodingConfig
	TxSigningConfig
	TxBuilderProvider
}

// TxEncodingConfig defines an interface that contains transaction
// encoders and decoders
type TxEncodingConfig interface {
	TxEncoder() TxApiEncoder
	TxDecoder() TxApiDecoder
	TxJSONEncoder() TxApiEncoder
	TxJSONDecoder() TxApiDecoder
}

type TxSigningConfig interface {
	SignModeHandler() *signing.HandlerMap
	SigningContext() *signing.Context
	MarshalSignatureJSON([]Signature) ([]byte, error)
	UnmarshalSignatureJSON([]byte) ([]Signature, error)
}

type TxParameters struct {
	timeoutHeight uint64
	chainID       string
	memo          string
	signMode      apitxsigning.SignMode

	AccountConfig
	GasConfig
	FeeConfig
	ExecutionOptions
	ExtensionOptions
}

// AccountConfig defines the 'account' related fields in a transaction.
type AccountConfig struct {
	accountNumber uint64
	sequence      uint64
	fromName      string
	fromAddress   sdk.AccAddress
}

// GasConfig defines the 'gas' related fields in a transaction.
type GasConfig struct {
	gas           uint64
	gasAdjustment float64
	gasPrices     []*base.DecCoin
}

// FeeConfig defines the 'fee' related fields in a transaction.
type FeeConfig struct {
	fees       []*base.Coin
	feeGranter sdk.AccAddress
	feePayer   sdk.AccAddress
}

// ExecutionOptions defines the transaction execution options ran by the client
type ExecutionOptions struct {
	unordered          bool
	offline            bool
	generateOnly       bool
	simulateAndExecute bool
	preprocessTxHook   PreprocessTxFn
}

type ExtensionOptions struct {
	ExtOptions []*codectypes.Any
}
