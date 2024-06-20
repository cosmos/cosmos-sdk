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

// SignatureDataToModeInfoAndSig converts SignatureData to ModeInfo and its corresponding raw signature.
func SignatureDataToModeInfoAndSig(data SignatureData) (*apitx.ModeInfo, []byte) {
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
			modeInfos[i], sigs[i] = SignatureDataToModeInfoAndSig(d)
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

// ModeInfoAndSigToSignatureData converts ModeInfo and a raw signature to SignatureData.
func ModeInfoAndSigToSignatureData(modeInfo *apitx.ModeInfo, sig []byte) (SignatureData, error) {
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
			sigsV2[i], err = ModeInfoAndSigToSignatureData(mi, sigs[i])
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
