package tx

import (
	"cosmossdk.io/core/transaction"
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/gogoproto/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"reflect"
	"strings"

	gogoany "github.com/cosmos/gogoproto/types/any"

	base "cosmossdk.io/api/cosmos/base/v1beta1"
	apitxsigning "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	apitx "cosmossdk.io/api/cosmos/tx/v1beta1"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
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

func NewGasConfig(gas uint64, gasAdjustment float64, gasPrices []*base.DecCoin) GasConfig {
	if gas == 0 {
		gas = 200000
	}
	return GasConfig{
		gas:           gas,
		gasAdjustment: gasAdjustment,
		gasPrices:     gasPrices,
	}
}

// FeeConfig defines the 'fee' related fields in a transaction.
type FeeConfig struct {
	fees       []*base.Coin
	feeGranter string
	feePayer   string
}

// ExecutionOptions defines the transaction execution options ran by the client
type ExecutionOptions struct {
	unordered          bool
	offline            bool
	offChain           bool
	generateOnly       bool
	simulateAndExecute bool
	preprocessTxHook   PreprocessTxFn
}

type ExtensionOptions struct {
	ExtOptions []*gogoany.Any
}

// GasEstimateResponse defines a response definition for tx gas estimation.
type GasEstimateResponse struct {
	GasEstimate uint64 `json:"gas_estimate" yaml:"gas_estimate"`
}

func (gr GasEstimateResponse) String() string {
	return fmt.Sprintf("gas estimate: %d", gr.GasEstimate)
}

type Tx interface {
	GetMsgs() ([]transaction.Msg, error)
	GetSigners() ([][]byte, error)
	GetPubKeys() ([]cryptotypes.PubKey, error)
	GetSignatures() ([]Signature, error)
}

type wrappedTx struct {
	tx  *apitx.Tx
	cdc codec.BinaryCodec
}

func (w wrappedTx) GetMsgs() ([]transaction.Msg, error) {
	panic("implement me")
}

func (w wrappedTx) GetSigners() ([][]byte, error) {
	panic("implement me")
}

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
