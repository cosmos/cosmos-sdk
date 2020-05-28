package types

import "github.com/cosmos/cosmos-sdk/crypto/types"

type SignatureV2 interface {
	isSignatureV2()
}

type SingleSignature struct {
	SignMode  SignMode
	Signature []byte
}

type MultiSignature struct {
	BitArray   *types.CompactBitArray
	Signatures []SignatureV2
}

var _, _ SignatureV2 = &SingleSignature{}, &MultiSignature{}

func (m *SingleSignature) isSignatureV2() {}
func (m *MultiSignature) isSignatureV2()  {}

func ModeInfoToSignatureV2(modeInfo *ModeInfo, sig []byte) (SignatureV2, error) {
	switch modeInfo := modeInfo.Sum.(type) {
	case *ModeInfo_Single_:
		return &SingleSignature{
			SignMode:  modeInfo.Single.Mode,
			Signature: sig,
		}, nil

	case *ModeInfo_Multi_:
		multi := modeInfo.Multi

		sigs, err := types.DecodeMultisignatures(sig)
		if err != nil {
			return nil, err
		}

		sigv2s := make([]SignatureV2, len(sigs))
		for i, mi := range multi.ModeInfos {
			sigv2s[i], err = ModeInfoToSignatureV2(mi, sigs[i])
			if err != nil {
				return nil, err
			}
		}

		return &MultiSignature{
			BitArray:   multi.Bitarray,
			Signatures: sigv2s,
		}, nil

	default:
		panic("unexpected case")
	}
}

