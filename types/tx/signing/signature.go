package signing

import (
	"fmt"

	"github.com/tendermint/tendermint/crypto"
)

// SignatureV2 is a convenience type that is easier to use in application logic
// than the protobuf SignerInfo's and raw signature bytes. It goes beyond the
// first sdk.Signature types by supporting sign modes and explicitly nested
// multi-signatures. It is intended to be used for both building and verifying
// signatures.
type SignatureV2 struct {
	// PubKey is the public key to use for verifying the signature
	PubKey crypto.PubKey

	// Data is the actual data of the signature which includes SignMode's and
	// the signatures themselves for either single or multi-signatures.
	Data SignatureData

	// Sequence is the sequence of this account. Only populated in
	// SIGN_MODE_DIRECT.
	Sequence uint64

	// Ugly flag to keep backwards-compatibility with Amino StdSignatures.
	// In SIGN_MODE_DIRECT, sequence is in AuthInfo, and will thus be populated
	// in the Sequence field above. The ante handler then checks this Sequence
	// with the actual sequence on-chain.
	// In SIGN_MODE_LEGACY_AMINO_JSON, sequence is signed via StdSignDoc, and
	// checked during signature verification. It's not populated in the
	// Sequence field above. This flag indicates that the Sequence field should
	// be skipped in ante handlers.
	// TLDR;
	// - false (by default) in SIGN_MODE_DIRECT
	// - true in SIGN_MODE_LEGACY_AMINO_JSON
	// ref: https://github.com/cosmos/cosmos-sdk/issues/7229
	SkipSequenceCheck bool
}

// SignatureDataToProto converts a SignatureData to SignatureDescriptor_Data.
// SignatureDescriptor_Data is considered an encoding type whereas SignatureData is used for
// business logic.
func SignatureDataToProto(data SignatureData) *SignatureDescriptor_Data {
	switch data := data.(type) {
	case *SingleSignatureData:
		return &SignatureDescriptor_Data{
			Sum: &SignatureDescriptor_Data_Single_{
				Single: &SignatureDescriptor_Data_Single{
					Mode:      data.SignMode,
					Signature: data.Signature,
				},
			},
		}
	case *MultiSignatureData:
		descDatas := make([]*SignatureDescriptor_Data, len(data.Signatures))

		for j, d := range data.Signatures {
			descDatas[j] = SignatureDataToProto(d)
		}

		return &SignatureDescriptor_Data{
			Sum: &SignatureDescriptor_Data_Multi_{
				Multi: &SignatureDescriptor_Data_Multi{
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
func SignatureDataFromProto(descData *SignatureDescriptor_Data) SignatureData {
	switch descData := descData.Sum.(type) {
	case *SignatureDescriptor_Data_Single_:
		return &SingleSignatureData{
			SignMode:  descData.Single.Mode,
			Signature: descData.Single.Signature,
		}
	case *SignatureDescriptor_Data_Multi_:
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
