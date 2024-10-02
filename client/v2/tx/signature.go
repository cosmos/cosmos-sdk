package tx

import (
	"errors"
	"fmt"

	apicrypto "cosmossdk.io/api/cosmos/crypto/multisig/v1beta1"
	apitxsigning "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	apitx "cosmossdk.io/api/cosmos/tx/v1beta1"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

// Signature holds the necessary components to verify transaction signatures.
type Signature struct {
	PubKey   cryptotypes.PubKey // Public key for signature verification.
	Data     SignatureData      // Signature data containing the actual signatures.
	Sequence uint64             // Account sequence, relevant for SIGN_MODE_DIRECT.
}

// SignatureData defines an interface for different signature data types.
type SignatureData interface {
	isSignatureData()
}

// SingleSignatureData stores a single signer's signature and its mode.
type SingleSignatureData struct {
	SignMode  apitxsigning.SignMode // Mode of the signature.
	Signature []byte                // Actual binary signature.
}

// MultiSignatureData encapsulates signatures from a multisig transaction.
type MultiSignatureData struct {
	BitArray   *apicrypto.CompactBitArray // Bitmap of signers.
	Signatures []SignatureData            // Individual signatures.
}

func (m *SingleSignatureData) isSignatureData() {}
func (m *MultiSignatureData) isSignatureData()  {}

// signatureDataToModeInfoAndSig converts SignatureData to ModeInfo and its corresponding raw signature.
func signatureDataToModeInfoAndSig(data SignatureData) (*apitx.ModeInfo, []byte) {
	if data == nil {
		return nil, nil
	}

	switch data := data.(type) {
	case *SingleSignatureData:
		return &apitx.ModeInfo{
			Sum: &apitx.ModeInfo_Single_{
				Single: &apitx.ModeInfo_Single{Mode: data.SignMode},
			},
		}, data.Signature
	case *MultiSignatureData:
		modeInfos := make([]*apitx.ModeInfo, len(data.Signatures))
		sigs := make([][]byte, len(data.Signatures))

		for i, d := range data.Signatures {
			modeInfos[i], sigs[i] = signatureDataToModeInfoAndSig(d)
		}

		multisig := cryptotypes.MultiSignature{Signatures: sigs}
		sig, err := multisig.Marshal()
		if err != nil {
			panic(err)
		}

		return &apitx.ModeInfo{
			Sum: &apitx.ModeInfo_Multi_{
				Multi: &apitx.ModeInfo_Multi{
					Bitarray:  data.BitArray,
					ModeInfos: modeInfos,
				},
			},
		}, sig
	default:
		panic(fmt.Sprintf("unexpected signature data type %T", data))
	}
}

// modeInfoAndSigToSignatureData converts ModeInfo and a raw signature to SignatureData.
func modeInfoAndSigToSignatureData(modeInfo *apitx.ModeInfo, sig []byte) (SignatureData, error) {
	switch mi := modeInfo.Sum.(type) {
	case *apitx.ModeInfo_Single_:
		return &SingleSignatureData{
			SignMode:  mi.Single.Mode,
			Signature: sig,
		}, nil

	case *apitx.ModeInfo_Multi_:
		multi := mi.Multi

		sigs, err := decodeMultiSignatures(sig)
		if err != nil {
			return nil, err
		}

		sigsV2 := make([]SignatureData, len(sigs))
		for i, mi := range multi.ModeInfos {
			sigsV2[i], err = modeInfoAndSigToSignatureData(mi, sigs[i])
			if err != nil {
				return nil, err
			}
		}
		return &MultiSignatureData{
			BitArray:   multi.Bitarray,
			Signatures: sigsV2,
		}, nil
	}

	return nil, fmt.Errorf("unsupported ModeInfo type %T", modeInfo)
}

// decodeMultiSignatures decodes a byte array into individual signatures.
func decodeMultiSignatures(bz []byte) ([][]byte, error) {
	multisig := cryptotypes.MultiSignature{}

	err := multisig.Unmarshal(bz)
	if err != nil {
		return nil, err
	}

	if len(multisig.XXX_unrecognized) > 0 {
		return nil, errors.New("unrecognized fields in MultiSignature")
	}
	return multisig.Signatures, nil
}

// signatureDataToProto converts a SignatureData interface to a protobuf SignatureDescriptor_Data.
// This function supports both SingleSignatureData and MultiSignatureData types.
// For SingleSignatureData, it directly maps the signature mode and signature bytes to the protobuf structure.
// For MultiSignatureData, it recursively converts each signature in the collection to the corresponding protobuf structure.
func signatureDataToProto(data SignatureData) (*apitxsigning.SignatureDescriptor_Data, error) {
	switch data := data.(type) {
	case *SingleSignatureData:
		// Handle single signature data conversion.
		return &apitxsigning.SignatureDescriptor_Data{
			Sum: &apitxsigning.SignatureDescriptor_Data_Single_{
				Single: &apitxsigning.SignatureDescriptor_Data_Single{
					Mode:      data.SignMode,
					Signature: data.Signature,
				},
			},
		}, nil
	case *MultiSignatureData:
		var err error
		descDatas := make([]*apitxsigning.SignatureDescriptor_Data, len(data.Signatures))

		for i, j := range data.Signatures {
			descDatas[i], err = signatureDataToProto(j)
			if err != nil {
				return nil, err
			}
		}
		return &apitxsigning.SignatureDescriptor_Data{
			Sum: &apitxsigning.SignatureDescriptor_Data_Multi_{
				Multi: &apitxsigning.SignatureDescriptor_Data_Multi{
					Bitarray:   data.BitArray,
					Signatures: descDatas,
				},
			},
		}, nil
	}

	// Return an error if the data type is not supported.
	return nil, fmt.Errorf("unexpected signature data type %T", data)
}

// SignatureDataFromProto converts a protobuf SignatureDescriptor_Data to a SignatureData interface.
// This function supports both Single and Multi signature data types.
func SignatureDataFromProto(descData *apitxsigning.SignatureDescriptor_Data) (SignatureData, error) {
	switch descData := descData.Sum.(type) {
	case *apitxsigning.SignatureDescriptor_Data_Single_:
		return &SingleSignatureData{
			SignMode:  descData.Single.Mode,
			Signature: descData.Single.Signature,
		}, nil
	case *apitxsigning.SignatureDescriptor_Data_Multi_:
		var err error
		multi := descData.Multi
		data := make([]SignatureData, len(multi.Signatures))

		for i, j := range multi.Signatures {
			data[i], err = SignatureDataFromProto(j)
			if err != nil {
				return nil, err
			}
		}

		return &MultiSignatureData{
			BitArray:   multi.Bitarray,
			Signatures: data,
		}, nil
	}

	return nil, fmt.Errorf("unexpected signature data type %T", descData)
}
