package tx

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
)

// SignatureDataToModeInfoAndSig converts a SignatureData to a ModeInfo and raw bytes signature
func SignatureDataToModeInfoAndSig(data signing.SignatureData) (*tx.ModeInfo, []byte) {
	if data == nil {
		return nil, nil
	}

	switch data := data.(type) {
	case *signing.SingleSignatureData:
		return &tx.ModeInfo{
			Sum: &tx.ModeInfo_Single_{
				Single: &tx.ModeInfo_Single{Mode: data.SignMode},
			},
		}, data.Signature
	case *signing.MultiSignatureData:
		n := len(data.Signatures)
		modeInfos := make([]*tx.ModeInfo, n)
		sigs := make([][]byte, n)

		for i, d := range data.Signatures {
			modeInfos[i], sigs[i] = SignatureDataToModeInfoAndSig(d)
		}

		multisig := types.MultiSignature{
			Signatures: sigs,
		}
		sig, err := multisig.Marshal()
		if err != nil {
			panic(err)
		}

		return &tx.ModeInfo{
			Sum: &tx.ModeInfo_Multi_{
				Multi: &tx.ModeInfo_Multi{
					Bitarray:  data.BitArray,
					ModeInfos: modeInfos,
				},
			},
		}, sig
	default:
		panic(fmt.Sprintf("unexpected signature data type %T", data))
	}
}

// ModeInfoAndSigToSignatureData converts a ModeInfo and raw bytes signature to a SignatureData or returns
// an error
func ModeInfoAndSigToSignatureData(modeInfo *tx.ModeInfo, sig []byte) (signing.SignatureData, error) {
	switch modeInfo := modeInfo.Sum.(type) {
	case *tx.ModeInfo_Single_:
		return &signing.SingleSignatureData{
			SignMode:  modeInfo.Single.Mode,
			Signature: sig,
		}, nil

	case *tx.ModeInfo_Multi_:
		multi := modeInfo.Multi

		sigs, err := decodeMultisignatures(sig)
		if err != nil {
			return nil, err
		}

		sigv2s := make([]signing.SignatureData, len(sigs))
		for i, mi := range multi.ModeInfos {
			sigv2s[i], err = ModeInfoAndSigToSignatureData(mi, sigs[i])
			if err != nil {
				return nil, err
			}
		}

		return &signing.MultiSignatureData{
			BitArray:   multi.Bitarray,
			Signatures: sigv2s,
		}, nil

	default:
		panic(fmt.Errorf("unexpected ModeInfo data type %T", modeInfo))
	}
}

// decodeMultisignatures safely decodes the the raw bytes as a MultiSignature protobuf message
func decodeMultisignatures(bz []byte) ([][]byte, error) {
	multisig := types.MultiSignature{}
	err := multisig.Unmarshal(bz)
	if err != nil {
		return nil, err
	}
	// NOTE: it is import to reject multi-signatures that contain unrecognized fields because this is an exploitable
	// malleability in the protobuf message. Basically an attacker could bloat a MultiSignature message with unknown
	// fields, thus bloating the transaction and causing it to fail.
	if len(multisig.XXX_unrecognized) > 0 {
		return nil, fmt.Errorf("rejecting unrecognized fields found in MultiSignature")
	}
	return multisig.Signatures, nil
}

func (g generator) MarshalSignatureJSON(sigs []signing.SignatureV2) ([]byte, error) {
	descs := make([]*tx.SignatureDescriptor, len(sigs))

	for i, sig := range sigs {
		publicKey, err := g.pubkeyCodec.Encode(sig.PubKey)
		if err != nil {
			return nil, err
		}

		descData := sigDataToSigDescData(sig.Data)

		descs[i] = &tx.SignatureDescriptor{
			PublicKey: publicKey,
			Data:      descData,
		}
	}

	toJson := &tx.SignatureDescriptors{Signatures: descs}

	return codec.ProtoMarshalJSON(toJson)
}

func sigDataToSigDescData(data signing.SignatureData) *tx.SignatureDescriptor_Data {
	switch data := data.(type) {
	case *signing.SingleSignatureData:
		return &tx.SignatureDescriptor_Data{
			Sum: &tx.SignatureDescriptor_Data_Single_{
				Single: &tx.SignatureDescriptor_Data_Single{
					Mode:      data.SignMode,
					Signature: data.Signature,
				},
			},
		}
	case *signing.MultiSignatureData:
		descDatas := make([]*tx.SignatureDescriptor_Data, len(data.Signatures))

		for j, d := range data.Signatures {
			descDatas[j] = sigDataToSigDescData(d)
		}

		return &tx.SignatureDescriptor_Data{
			Sum: &tx.SignatureDescriptor_Data_Multi_{
				Multi: &tx.SignatureDescriptor_Data_Multi{
					Bitarray:   data.BitArray,
					Signatures: descDatas,
				},
			},
		}
	default:
		panic(fmt.Errorf("unexpected case %+v", data))
	}
}

func (g generator) UnmarshalSignatureJSON(bz []byte) ([]signing.SignatureV2, error) {
	var sigDescs tx.SignatureDescriptors
	err := g.protoCodec.UnmarshalJSON(bz, &sigDescs)
	if err != nil {
		return nil, err
	}

	sigs := make([]signing.SignatureV2, len(sigDescs.Signatures))
	for i, desc := range sigDescs.Signatures {
		pubKey, err := g.pubkeyCodec.Decode(desc.PublicKey)
		if err != nil {
			return nil, err
		}

		data := sigDescDataToSigData(desc.Data)

		sigs[i] = signing.SignatureV2{
			PubKey: pubKey,
			Data:   data,
		}
	}

	return sigs, nil
}

func sigDescDataToSigData(descData *tx.SignatureDescriptor_Data) signing.SignatureData {
	switch descData := descData.Sum.(type) {
	case *tx.SignatureDescriptor_Data_Single_:
		return &signing.SingleSignatureData{
			SignMode:  descData.Single.Mode,
			Signature: descData.Single.Signature,
		}
	case *tx.SignatureDescriptor_Data_Multi_:
		multi := descData.Multi
		datas := make([]signing.SignatureData, len(multi.Signatures))

		for j, d := range multi.Signatures {
			datas[j] = sigDescDataToSigData(d)
		}

		return &signing.MultiSignatureData{
			BitArray:   multi.Bitarray,
			Signatures: datas,
		}
	default:
		panic(fmt.Errorf("unexpected case %+v", descData))
	}
}
