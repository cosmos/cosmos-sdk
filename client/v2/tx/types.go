package tx

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/cosmos/gogoproto/proto"
	gogoany "github.com/cosmos/gogoproto/types/any"
	"google.golang.org/protobuf/types/known/anypb"

	base "cosmossdk.io/api/cosmos/base/v1beta1"
	apitxsigning "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	apitx "cosmossdk.io/api/cosmos/tx/v1beta1"
	"cosmossdk.io/client/v2/internal/coins"
	"cosmossdk.io/core/transaction"

	"github.com/cosmos/cosmos-sdk/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

const defaultGas = 200000

// PreprocessTxFn defines a hook by which chains can preprocess transactions before broadcasting
type PreprocessTxFn func(chainID string, key uint, tx TxBuilder) error

// HasValidateBasic is a copy of types.HasValidateBasic to avoid sdk import.
type HasValidateBasic interface {
	// ValidateBasic does a simple validation check that
	// doesn't require access to any other information.
	ValidateBasic() error
}

// TxParameters defines the parameters required for constructing a transaction.
type TxParameters struct {
	timeoutHeight uint64                // timeoutHeight indicates the block height after which the transaction is no longer valid.
	chainID       string                // chainID specifies the unique identifier of the blockchain where the transaction will be processed.
	memo          string                // memo contains any arbitrary memo to be attached to the transaction.
	signMode      apitxsigning.SignMode // signMode determines the signing mode to be used for the transaction.

	AccountConfig    // AccountConfig includes information about the transaction originator's account.
	GasConfig        // GasConfig specifies the gas settings for the transaction.
	FeeConfig        // FeeConfig details the fee associated with the transaction.
	ExecutionOptions // ExecutionOptions includes settings that modify how the transaction is executed.
	ExtensionOptions // ExtensionOptions allows for additional features or data to be included in the transaction.
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
	unordered          bool           // unordered indicates if the transaction execution order is not guaranteed.
	offline            bool           // offline specifies if the transaction should be prepared for offline signing.
	offChain           bool           // offChain indicates if the transaction should be executed off the blockchain.
	generateOnly       bool           // generateOnly specifies if the transaction should only be generated and not executed.
	simulateAndExecute bool           // simulateAndExecute indicates if the transaction should be simulated before execution.
	preprocessTxHook   PreprocessTxFn // preprocessTxHook is a function hook for preprocessing the transaction.
}

// ExtensionOptions holds a slice of Any protocol buffer messages that can be used to extend the functionality
// of a transaction with additional data. This is typically used to include non-standard or experimental features.
type ExtensionOptions struct {
	ExtOptions []*gogoany.Any // ExtOptions are the extension options in the form of Any protocol buffer messages.
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
	// GetMsgs retrieves the messages included in the transaction.
	GetMsgs() ([]transaction.Msg, error)
	// GetSigners fetches the addresses of the signers of the transaction.
	GetSigners() ([][]byte, error)
	// GetPubKeys retrieves the public keys of the signers of the transaction.
	GetPubKeys() ([]cryptotypes.PubKey, error)
	// GetSignatures fetches the signatures attached to the transaction.
	GetSignatures() ([]Signature, error)
}

// wrappedTx wraps a transaction and provides a codec for binary encoding/decoding.
type wrappedTx struct {
	tx  *apitx.Tx         // tx is the transaction being wrapped.
	cdc codec.BinaryCodec // cdc is the codec used for binary encoding/decoding.
}

// GetMsgs retrieves the messages included in the transaction.
func (w wrappedTx) GetMsgs() ([]transaction.Msg, error) {
	return nil, errors.New("not implemented")
}

// GetSigners fetches the addresses of the signers of the transaction.
func (w wrappedTx) GetSigners() ([][]byte, error) {
	return nil, errors.New("not implemented")
}

// GetPubKeys retrieves the public keys of the signers from the transaction's SignerInfos.
func (w wrappedTx) GetPubKeys() ([]cryptotypes.PubKey, error) {
	signerInfos := w.tx.AuthInfo.SignerInfos
	pks := make([]cryptotypes.PubKey, len(signerInfos))

	for i, si := range signerInfos {
		// NOTE: it is okay to leave this nil if there is no PubKey in the SignerInfo.
		// PubKey's can be left unset in SignerInfo.
		if si.PublicKey == nil {
			continue
		}
		maybePk, err := w.decodeAny(si.PublicKey)
		if err != nil {
			return nil, err
		}
		pk, ok := maybePk.(cryptotypes.PubKey)
		if !ok {
			return nil, fmt.Errorf("invalid public key type: %T", maybePk)
		}
		pks[i] = pk
	}

	return pks, nil
}

// GetSignatures fetches the signatures attached to the transaction.
func (w wrappedTx) GetSignatures() ([]Signature, error) {
	signerInfos := w.tx.AuthInfo.SignerInfos
	sigs := w.tx.Signatures

	pubKeys, err := w.GetPubKeys()
	if err != nil {
		return nil, err
	}
	signatures := make([]Signature, len(sigs))

	for i, si := range signerInfos {
		if si.ModeInfo == nil || si.ModeInfo.Sum == nil {
			signatures[i] = Signature{
				PubKey: pubKeys[i],
			}
		} else {
			sigData, err := ModeInfoAndSigToSignatureData(si.ModeInfo, sigs[i])
			if err != nil {
				return nil, err
			}
			signatures[i] = Signature{
				PubKey:   pubKeys[i],
				Data:     sigData,
				Sequence: si.GetSequence(),
			}
		}
	}

	return signatures, nil
}

// decodeAny decodes a protobuf Any message into a concrete proto.Message.
func (w wrappedTx) decodeAny(anyPb *anypb.Any) (proto.Message, error) {
	name := anyPb.GetTypeUrl()
	if i := strings.LastIndexByte(name, '/'); i >= 0 {
		name = name[i+len("/"):]
	}
	typ := proto.MessageType(name)
	if typ == nil {
		return nil, fmt.Errorf("unknown type: %s", name)
	}
	v1 := reflect.New(typ.Elem()).Interface().(proto.Message)
	err := w.cdc.Unmarshal(anyPb.GetValue(), v1)
	if err != nil {
		return nil, err
	}
	return v1, nil
}
