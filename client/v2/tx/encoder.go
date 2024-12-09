package tx

import (
	"fmt"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/encoding/prototext"
	protov2 "google.golang.org/protobuf/proto"

	txv1beta1 "cosmossdk.io/api/cosmos/tx/v1beta1"
	txdecode "cosmossdk.io/x/tx/decode"

	"github.com/cosmos/cosmos-sdk/codec"
)

var (
	// marshalOption configures protobuf marshaling to be deterministic.
	marshalOption = protov2.MarshalOptions{Deterministic: true}

	// jsonMarshalOptions configures JSON marshaling for protobuf messages.
	jsonMarshalOptions = protojson.MarshalOptions{
		Indent:          "",
		UseProtoNames:   true,
		UseEnumNumbers:  false,
		EmitUnpopulated: true,
	}

	// textMarshalOptions
	textMarshalOptions = prototext.MarshalOptions{
		Indent: "",
	}
)

// Decoder defines the interface for decoding transaction bytes into a DecodedTx.
type Decoder interface {
	Decode(txBytes []byte) (*txdecode.DecodedTx, error)
}

// txDecoder is a function type that unmarshals transaction bytes into an API Tx type.
type txDecoder func(txBytes []byte) (Tx, error)

// txEncoder is a function type that marshals a transaction into bytes.
type txEncoder func(tx Tx) ([]byte, error)

// decodeTx decodes transaction bytes into an apitx.Tx structure.
func decodeTx(cdc codec.BinaryCodec, decoder Decoder) txDecoder {
	return func(txBytes []byte) (Tx, error) {
		tx := new(txv1beta1.Tx)
		err := protov2.Unmarshal(txBytes, tx)
		if err != nil {
			return nil, err
		}

		pTxBytes, err := protoTxBytes(tx)
		if err != nil {
			return nil, err
		}

		decodedTx, err := decoder.Decode(pTxBytes)
		if err != nil {
			return nil, err
		}
		return newWrapperTx(cdc, decodedTx), nil
	}
}

// encodeTx encodes an apitx.Tx into bytes using protobuf marshaling options.
func encodeTx(tx Tx) ([]byte, error) {
	wTx, ok := tx.(*wrappedTx)
	if !ok {
		return nil, fmt.Errorf("unexpected tx type: %T", tx)
	}
	return marshalOption.Marshal(wTx.Tx)
}

// decodeJsonTx decodes transaction bytes into an apitx.Tx structure using JSON format.
func decodeJsonTx(cdc codec.BinaryCodec, decoder Decoder) txDecoder {
	return func(txBytes []byte) (Tx, error) {
		jsonTx := new(txv1beta1.Tx)
		err := protojson.UnmarshalOptions{
			AllowPartial:   false,
			DiscardUnknown: false,
		}.Unmarshal(txBytes, jsonTx)
		if err != nil {
			return nil, err
		}

		pTxBytes, err := protoTxBytes(jsonTx)
		if err != nil {
			return nil, err
		}

		decodedTx, err := decoder.Decode(pTxBytes)
		if err != nil {
			return nil, err
		}
		return newWrapperTx(cdc, decodedTx), nil
	}
}

// encodeJsonTx encodes an apitx.Tx into bytes using JSON marshaling options.
func encodeJsonTx(tx Tx) ([]byte, error) {
	wTx, ok := tx.(*wrappedTx)
	if !ok {
		return nil, fmt.Errorf("unexpected tx type: %T", tx)
	}
	return jsonMarshalOptions.Marshal(wTx.Tx)
}

func encodeTextTx(tx Tx) ([]byte, error) {
	wTx, ok := tx.(*wrappedTx)
	if !ok {
		return nil, fmt.Errorf("unexpected tx type: %T", tx)
	}
	return textMarshalOptions.Marshal(wTx.Tx)
}

// decodeJsonTx decodes transaction bytes into an apitx.Tx structure using JSON format.
func decodeTextTx(cdc codec.BinaryCodec, decoder Decoder) txDecoder {
	return func(txBytes []byte) (Tx, error) {
		jsonTx := new(txv1beta1.Tx)
		err := prototext.UnmarshalOptions{
			AllowPartial:   false,
			DiscardUnknown: false,
		}.Unmarshal(txBytes, jsonTx)
		if err != nil {
			return nil, err
		}

		pTxBytes, err := protoTxBytes(jsonTx)
		if err != nil {
			return nil, err
		}

		decodedTx, err := decoder.Decode(pTxBytes)
		if err != nil {
			return nil, err
		}
		return newWrapperTx(cdc, decodedTx), nil
	}
}

func protoTxBytes(tx *txv1beta1.Tx) ([]byte, error) {
	bodyBytes, err := marshalOption.Marshal(tx.Body)
	if err != nil {
		return nil, err
	}

	authInfoBytes, err := marshalOption.Marshal(tx.AuthInfo)
	if err != nil {
		return nil, err
	}

	return marshalOption.Marshal(&txv1beta1.TxRaw{
		BodyBytes:     bodyBytes,
		AuthInfoBytes: authInfoBytes,
		Signatures:    tx.Signatures,
	})
}
