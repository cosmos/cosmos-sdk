package tx

import (
	"fmt"

	apitxsigning "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	typestx "cosmossdk.io/api/cosmos/tx/v1beta1"
	"cosmossdk.io/core/transaction"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

// PreprocessTxFn defines a hook by which chains can preprocess transactions before broadcasting
type PreprocessTxFn func(chainID string, key uint, tx TxBuilder) error

// TxApiDecoder unmarshals transaction bytes into API Tx type
type TxApiDecoder func(txBytes []byte) (typestx.Tx, error)

// TxApiEncoder marshals transaction to bytes
type TxApiEncoder func(tx typestx.Tx) ([]byte, error)

// GasEstimateResponse defines a response definition for tx gas estimation.
type GasEstimateResponse struct {
	GasEstimate uint64 `json:"gas_estimate" yaml:"gas_estimate"`
}

func (gr GasEstimateResponse) String() string {
	return fmt.Sprintf("gas estimate: %d", gr.GasEstimate)
}

type TxWrapper struct {
	Tx *typestx.Tx
}

func (tx TxWrapper) GetMsgs() ([]transaction.Msg, error) {
	//TODO implement me
	panic("implement me")
}

func (tx TxWrapper) GetSignatures() ([]Signature, error) {
	//TODO implement me
	panic("implement me")
}

type Signature struct {
	// PubKey is the public key to use for verifying the signature
	PubKey cryptotypes.PubKey

	// Data is the actual data of the signature which includes SignMode's and
	// the signatures themselves for either single or multi-signatures.
	Data SignatureData

	// Sequence is the sequence of this account. Only populated in
	// SIGN_MODE_DIRECT.
	Sequence uint64
}

type SignatureData interface {
	isSignatureData()
}

// SingleSignatureData represents the signature and SignMode of a single (non-multisig) signer
type SingleSignatureData struct {
	// SignMode represents the SignMode of the signature
	SignMode apitxsigning.SignMode

	// Signature is the raw signature.
	Signature []byte
}

// MultiSignatureData represents the nested SignatureData of a multisig signature
type MultiSignatureData struct {
	// BitArray is a compact way of indicating which signers from the multisig key
	// have signed
	BitArray *cryptotypes.CompactBitArray

	// Signatures is the nested SignatureData's for each signer
	Signatures []SignatureData
}

func (m *SingleSignatureData) isSignatureData() {}
func (m *MultiSignatureData) isSignatureData()  {}
