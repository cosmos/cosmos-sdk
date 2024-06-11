package tx

import (
	"errors"
	"fmt"

	apicrypto "cosmossdk.io/api/cosmos/crypto/multisig/v1beta1"
	apitxsigning "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	apitx "cosmossdk.io/api/cosmos/tx/v1beta1"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

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
	BitArray *apicrypto.CompactBitArray

	// Signatures is the nested SignatureData's for each signer
	Signatures []SignatureData
}

func (m *SingleSignatureData) isSignatureData() {}
func (m *MultiSignatureData) isSignatureData()  {}

// SignatureDataToModeInfoAndSig converts a SignatureData to a ModeInfo and raw bytes signature
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
		n := len(data.Signatures)
		modeInfos := make([]*apitx.ModeInfo, n)
		sigs := make([][]byte, n)

		for i, d := range data.Signatures {
			modeInfos[i], sigs[i] = SignatureDataToModeInfoAndSig(d)
		}

		multisig := cryptotypes.MultiSignature{
			Signatures: sigs,
		}
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

			return &MultiSignatureData{
				BitArray: &apicrypto.CompactBitArray{
					ExtraBitsStored: multi.Bitarray.GetExtraBitsStored(),
					Elems:           multi.Bitarray.GetElems(),
				},
				Signatures: sigsV2,
			}, nil
		}
	}

	return nil, fmt.Errorf("unsupported ModeInfo type %T", modeInfo)
}

func decodeMultiSignatures(bz []byte) ([][]byte, error) {
	multisig := cryptotypes.MultiSignature{}

	err := multisig.Unmarshal(bz)
	if err != nil {
		return nil, err
	}

	if len(multisig.XXX_unrecognized) > 0 {
		return nil, errors.New("rejecting unrecognized fields found in MultiSignature")
	}
	return multisig.Signatures, nil
}
