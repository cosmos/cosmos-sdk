package signing

import (
	"github.com/cosmos/cosmos-sdk/crypto/types"
)

// SignatureData represents either a *SingleSignatureData or *MultiSignatureData.
// It is a convenience type that is easier to use in business logic than the encoded
// protobuf ModeInfo's and raw signatures.
// The interface here is a hack being used to for type casting.
// TODO: Make a correct interface instead
type SignatureData interface {
	isSignatureData()
}

// SingleSignatureData represents the signature and SignMode of a single (non-multisig) signer
type SingleSignatureData struct {
	// SignMode represents the SignMode of the signature
	SignMode SignMode

	// SignMode represents the SignMode of the signature
	Signature []byte
}

// MultiSignatureData represents the nested SignatureData of a multisig signature
type MultiSignatureData struct {
	// BitArray is a compact way of indicating which signers from the multisig key
	// have signed
	BitArray *types.CompactBitArray

	// Signatures is the nested SignatureData's for each signer
	Signatures []SignatureData
}

var _, _ SignatureData = &SingleSignatureData{}, &MultiSignatureData{}

func (m *SingleSignatureData) isSignatureData() {}
func (m *MultiSignatureData) isSignatureData()  {}

// SignerData is the specific information needed to sign a transaction that generally
// isn't included in the transaction body itself
// TODO: renamed from SignerData  -> SignatureMetadata
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
