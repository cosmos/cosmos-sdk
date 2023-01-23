package signing

import (
	"fmt"

	multisigv1beta1 "cosmossdk.io/api/cosmos/crypto/multisig/v1beta1"
	signingv1beta1 "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	"google.golang.org/protobuf/types/known/anypb"
)

// Signature is a convenience type that is easier to use in application logic
// than the protobuf SignerInfo's and raw signature bytes. It goes beyond the
// first sdk.Signature types by supporting sign modes and explicitly nested
// multi-signatures. It is intended to be used for both building and verifying
// signatures.
type Signature struct {
	// PubKey is the public key to use for verifying the signature
	PubKey *anypb.Any

	// Data is the actual data of the signature which includes SignMode's and
	// the signatures themselves for either single or multi-signatures.
	Data SignatureData

	// Sequence is the sequence of this account. Only populated in
	// SIGN_MODE_DIRECT.
	Sequence uint64
}

// SignatureDataToProto converts a SignatureData to SignatureDescriptor_Data.
// SignatureDescriptor_Data is considered an encoding type whereas SignatureData is used for
// business logic.
func SignatureDataToProto(data SignatureData) *signingv1beta1.SignatureDescriptor_Data {
	switch data := data.(type) {
	case *SingleSignatureData:
		return &signingv1beta1.SignatureDescriptor_Data{
			Sum: &signingv1beta1.SignatureDescriptor_Data_Single_{
				Single: &signingv1beta1.SignatureDescriptor_Data_Single{
					Mode:      data.SignMode,
					Signature: data.Signature,
				},
			},
		}
	case *MultiSignatureData:
		descDatas := make([]*signingv1beta1.SignatureDescriptor_Data, len(data.Signatures))

		for j, d := range data.Signatures {
			descDatas[j] = SignatureDataToProto(d)
		}

		return &signingv1beta1.SignatureDescriptor_Data{
			Sum: &signingv1beta1.SignatureDescriptor_Data_Multi_{
				Multi: &signingv1beta1.SignatureDescriptor_Data_Multi{
					Bitarray:   data.BitArray,
					Signatures: descDatas,
				},
			},
		}
	default:
		panic(fmt.Errorf("unexpected case %+v", data))
	}
}

// SignatureDataFromProto converts a SignatureDescriptor_Data to SignatureData.
// SignatureDescriptor_Data is considered an encoding type whereas SignatureData is used for
// business logic.
func SignatureDataFromProto(descData *signingv1beta1.SignatureDescriptor_Data) SignatureData {
	switch descData := descData.Sum.(type) {
	case *signingv1beta1.SignatureDescriptor_Data_Single_:
		return &SingleSignatureData{
			SignMode:  descData.Single.Mode,
			Signature: descData.Single.Signature,
		}
	case *signingv1beta1.SignatureDescriptor_Data_Multi_:
		multi := descData.Multi
		datas := make([]SignatureData, len(multi.Signatures))

		for j, d := range multi.Signatures {
			datas[j] = SignatureDataFromProto(d)
		}

		return &MultiSignatureData{
			BitArray:   multi.Bitarray,
			Signatures: datas,
		}
	default:
		panic(fmt.Errorf("unexpected case %+v", descData))
	}
}

// SignatureData represents either a *SingleSignatureData or *MultiSignatureData.
// It is a convenience type that is easier to use in business logic than the encoded
// protobuf ModeInfo's and raw signatures.
type SignatureData interface {
	isSignatureData()
}

// SingleSignatureData represents the signature and SignMode of a single (non-multisig) signer
type SingleSignatureData struct {
	// SignMode represents the SignMode of the signature
	SignMode signingv1beta1.SignMode

	// Signature is the raw signature.
	Signature []byte
}

// MultiSignatureData represents the nested SignatureData of a multisig signature
type MultiSignatureData struct {
	// BitArray is a compact way of indicating which signers from the multisig key
	// have signed
	BitArray *multisigv1beta1.CompactBitArray

	// Signatures is the nested SignatureData's for each signer
	Signatures []SignatureData
}

var _, _ SignatureData = &SingleSignatureData{}, &MultiSignatureData{}

func (m *SingleSignatureData) isSignatureData() {}
func (m *MultiSignatureData) isSignatureData()  {}
