package tx

import (
	keyring2 "cosmossdk.io/client/v2/autocli/keyring"
	"cosmossdk.io/core/address"
	"fmt"
	flags2 "github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/pflag"
	"time"

	base "cosmossdk.io/api/cosmos/base/v1beta1"
	apitxsigning "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	"cosmossdk.io/client/v2/internal/coins"
	"cosmossdk.io/core/transaction"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

const defaultGas = 200000

// HasValidateBasic is a copy of types.HasValidateBasic to avoid sdk import.
type HasValidateBasic interface {
	// ValidateBasic does a simple validation check that
	// doesn't require access to any other information.
	ValidateBasic() error
}

// TxParameters defines the parameters required for constructing a transaction.
type TxParameters struct {
	timeoutTimestamp time.Time             // timeoutTimestamp indicates a timestamp after which the transaction is no longer valid.
	chainID          string                // chainID specifies the unique identifier of the blockchain where the transaction will be processed.
	memo             string                // memo contains any arbitrary memo to be attached to the transaction.
	signMode         apitxsigning.SignMode // signMode determines the signing mode to be used for the transaction.

	AccountConfig    // AccountConfig includes information about the transaction originator's account.
	GasConfig        // GasConfig specifies the gas settings for the transaction.
	FeeConfig        // FeeConfig details the fee associated with the transaction.
	ExecutionOptions // ExecutionOptions includes settings that modify how the transaction is executed.
}

// AccountConfig defines the 'account' related fields in a transaction.
type AccountConfig struct {
	// accountNumber is the unique identifier for the account.
	accountNumber uint64
	// sequence is the sequence number of the transaction.
	sequence uint64
	// fromName is the name of the account sending the transaction.
	fromName string
	// fromAddress is the address of the account sending the transaction.
	fromAddress string
	// address is the byte representation of the account address.
	address []byte
}

// GasConfig defines the 'gas' related fields in a transaction.
// GasConfig defines the gas-related settings for a transaction.
type GasConfig struct {
	gas           uint64          // gas is the amount of gas requested for the transaction.
	gasAdjustment float64         // gasAdjustment is the factor by which the estimated gas is multiplied to calculate the final gas limit.
	gasPrices     []*base.DecCoin // gasPrices is a list of denominations of DecCoin used to calculate the fee paid for the gas.
}

// NewGasConfig creates a new GasConfig with the specified gas, gasAdjustment, and gasPrices.
// If the provided gas value is zero, it defaults to a predefined value (defaultGas).
// The gasPrices string is parsed into a slice of DecCoin.
func NewGasConfig(gas uint64, gasAdjustment float64, gasPrices string) (GasConfig, error) {
	if gas == 0 {
		gas = defaultGas
	}

	parsedGasPrices, err := coins.ParseDecCoins(gasPrices)
	if err != nil {
		return GasConfig{}, err
	}

	return GasConfig{
		gas:           gas,
		gasAdjustment: gasAdjustment,
		gasPrices:     parsedGasPrices,
	}, nil
}

// FeeConfig holds the fee details for a transaction.
type FeeConfig struct {
	fees       []*base.Coin // fees are the amounts paid for the transaction.
	feePayer   string       // feePayer is the account responsible for paying the fees.
	feeGranter string       // feeGranter is the account granting the fee payment if different from the payer.
}

// NewFeeConfig creates a new FeeConfig with the specified fees, feePayer, and feeGranter.
// It parses the fees string into a slice of Coin, handling normalization.
func NewFeeConfig(fees, feePayer, feeGranter string) (FeeConfig, error) {
	parsedFees, err := coins.ParseCoinsNormalized(fees)
	if err != nil {
		return FeeConfig{}, err
	}

	return FeeConfig{
		fees:       parsedFees,
		feePayer:   feePayer,
		feeGranter: feeGranter,
	}, nil
}

// ExecutionOptions defines the transaction execution options ran by the client
// ExecutionOptions defines the settings for transaction execution.
type ExecutionOptions struct {
	unordered          bool // unordered indicates if the transaction execution order is not guaranteed.
	offline            bool // offline specifies if the transaction should be prepared for offline signing.
	simulateAndExecute bool // simulateAndExecute indicates if the transaction should be simulated before execution.
}

// GasEstimateResponse defines a response definition for tx gas estimation.
type GasEstimateResponse struct {
	GasEstimate uint64 `json:"gas_estimate" yaml:"gas_estimate"`
}

func (gr GasEstimateResponse) String() string {
	return fmt.Sprintf("gas estimate: %d", gr.GasEstimate)
}

// Tx defines the interface for transaction operations.
type Tx interface {
	transaction.Tx

	// GetSigners fetches the addresses of the signers of the transaction.
	GetSigners() ([][]byte, error)
	// GetPubKeys retrieves the public keys of the signers of the transaction.
	GetPubKeys() ([]cryptotypes.PubKey, error)
	// GetSignatures fetches the signatures attached to the transaction.
	GetSignatures() ([]Signature, error)
}

// txParamsFromFlagSet extracts the transaction parameters from the provided FlagSet.
func txParamsFromFlagSet(flags *pflag.FlagSet, keybase keyring2.Keyring, ac address.Codec) (params TxParameters, err error) {
	timestampUnix, _ := flags.GetInt64(flags2.FlagTimeoutTimestamp)
	timeoutTimestamp := time.Unix(timestampUnix, 0)
	chainID, _ := flags.GetString(flags2.FlagChainID)
	memo, _ := flags.GetString(flags2.FlagNote)
	signMode, _ := flags.GetString(flags2.FlagSignMode)

	accNumber, _ := flags.GetUint64(flags2.FlagAccountNumber)
	sequence, _ := flags.GetUint64(flags2.FlagSequence)
	from, _ := flags.GetString(flags2.FlagFrom)

	var fromName, fromAddress string
	var addr []byte
	isDryRun, _ := flags.GetBool(flags2.FlagDryRun)
	if isDryRun {
		addr, err = ac.StringToBytes(from)
	} else {
		fromName, fromAddress, _, err = keybase.KeyInfo(from)
		if err == nil {
			addr, err = ac.StringToBytes(fromAddress)
		}
	}
	if err != nil {
		return params, err
	}

	gas, _ := flags.GetString(flags2.FlagGas)
	gasSetting, _ := flags2.ParseGasSetting(gas)
	gasAdjustment, _ := flags.GetFloat64(flags2.FlagGasAdjustment)
	gasPrices, _ := flags.GetString(flags2.FlagGasPrices)

	fees, _ := flags.GetString(flags2.FlagFees)
	feePayer, _ := flags.GetString(flags2.FlagFeePayer)
	feeGrater, _ := flags.GetString(flags2.FlagFeeGranter)

	unordered, _ := flags.GetBool(flags2.FlagUnordered)
	offline, _ := flags.GetBool(flags2.FlagOffline)

	gasConfig, err := NewGasConfig(gasSetting.Gas, gasAdjustment, gasPrices)
	if err != nil {
		return params, err
	}
	feeConfig, err := NewFeeConfig(fees, feePayer, feeGrater)
	if err != nil {
		return params, err
	}

	txParams := TxParameters{
		timeoutTimestamp: timeoutTimestamp,
		chainID:          chainID,
		memo:             memo,
		signMode:         getSignMode(signMode),
		AccountConfig: AccountConfig{
			accountNumber: accNumber,
			sequence:      sequence,
			fromName:      fromName,
			fromAddress:   fromAddress,
			address:       addr,
		},
		GasConfig: gasConfig,
		FeeConfig: feeConfig,
		ExecutionOptions: ExecutionOptions{
			unordered:          unordered,
			offline:            offline,
			simulateAndExecute: gasSetting.Simulate,
		},
	}

	return txParams, nil
}
