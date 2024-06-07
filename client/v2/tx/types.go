package tx

import (
	"fmt"

	base "cosmossdk.io/api/cosmos/base/v1beta1"
	apitxsigning "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	apitx "cosmossdk.io/api/cosmos/tx/v1beta1"
	"cosmossdk.io/core/transaction"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// PreprocessTxFn defines a hook by which chains can preprocess transactions before broadcasting
type PreprocessTxFn func(chainID string, key uint, tx TxBuilder) error

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

// GasEstimateResponse defines a response definition for tx gas estimation.
type GasEstimateResponse struct {
	GasEstimate uint64 `json:"gas_estimate" yaml:"gas_estimate"`
}

func (gr GasEstimateResponse) String() string {
	return fmt.Sprintf("gas estimate: %d", gr.GasEstimate)
}

type TxWrapper struct {
	Tx *apitx.Tx
}

func (tx TxWrapper) GetMsgs() ([]transaction.Msg, error) {
	//TODO implement me
	panic("implement me")
}

func (tx TxWrapper) GetSignatures() ([]Signature, error) {
	//TODO implement me
	panic("implement me")
}
