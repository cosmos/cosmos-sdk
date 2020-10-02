package signing

import (
	"fmt"

	"github.com/tendermint/tendermint/crypto"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
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

var _, _ codectypes.UnpackInterfacesMessage = &SignatureDescriptors{}, &SignatureDescriptor{}

// UnpackInterfaces implements the UnpackInterfaceMessages.UnpackInterfaces method
func (sds *SignatureDescriptors) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	for _, sig := range sds.Signatures {
		err := sig.UnpackInterfaces(unpacker)

		if err != nil {
			return err
		}
	}

	return nil
}

// UnpackInterfaces implements the UnpackInterfaceMessages.UnpackInterfaces method
// [Deprecated]
func (sd *SignatureDescriptor) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	return unpacker.UnpackAny(sd.PublicKey, new(crypto.PubKey))
}

// SignModeHandler defines a interface to be implemented by types which will handle
// SignMode's by generating sign bytes from a Tx and SignerData/
type SignModeHandler interface {
	// DefaultMode is the default mode that is to be used with this handler if no
	// other mode is specified. This can be useful for testing and CLI usage
	DefaultMode() SignMode

	// Modes is the list of modes supporting by this handler
	Modes() []SignMode

	// GetSignBytes returns the sign bytes for the provided SignMode, SignerData and Tx,
	// or an error
	GetSignBytes(mode SignMode, data SignerData, tx sdk.Tx) ([]byte, error)
}
