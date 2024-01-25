package tx

import (
	"google.golang.org/protobuf/encoding/protojson"

	txv1beta1 "cosmossdk.io/api/cosmos/tx/v1beta1"
	"cosmossdk.io/core/address"
	"cosmossdk.io/x/tx/decode"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DefaultJSONTxDecoder returns a default protobuf JSON TxDecoder using the provided Marshaler.
func DefaultJSONTxDecoder(addrCodec address.Codec, cdc codec.BinaryCodec, decoder *decode.Decoder) sdk.TxDecoder {
	jsonUnmarshaller := protojson.UnmarshalOptions{
		AllowPartial:   false,
		DiscardUnknown: false,
	}
	return func(txBytes []byte) (sdk.Tx, error) {
		jsonTx := new(txv1beta1.Tx)
		err := jsonUnmarshaller.Unmarshal(txBytes, jsonTx)
		if err != nil {
			return nil, err
		}

		// need to convert jsonTx into raw tx.
		bodyBytes, err := marshalOption.Marshal(jsonTx.Body)
		if err != nil {
			return nil, err
		}

		authInfoBytes, err := marshalOption.Marshal(jsonTx.AuthInfo)
		if err != nil {
			return nil, err
		}

		protoTxBytes, err := marshalOption.Marshal(&txv1beta1.TxRaw{
			BodyBytes:     bodyBytes,
			AuthInfoBytes: authInfoBytes,
			Signatures:    jsonTx.Signatures,
		})
		if err != nil {
			return nil, err
		}

		decodedTx, err := decoder.Decode(protoTxBytes)
		if err != nil {
			return nil, err
		}
		return newWrapperFromDecodedTx(addrCodec, cdc, decodedTx)
	}
}
