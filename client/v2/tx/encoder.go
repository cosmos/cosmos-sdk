package tx

import (
	"google.golang.org/protobuf/encoding/protojson"
	protov2 "google.golang.org/protobuf/proto"

	apitx "cosmossdk.io/api/cosmos/tx/v1beta1"
	txdecode "cosmossdk.io/x/tx/decode"
)

var (
	// marshalOption configures protobuf marshaling to be deterministic.
	marshalOption = protov2.MarshalOptions{Deterministic: true}

	// jsonMarshalOptions configures JSON marshaling for protobuf messages.
	jsonMarshalOptions = protojson.MarshalOptions{
		Indent:         "",
		UseProtoNames:  true,
		UseEnumNumbers: false,
	}
)

// Decoder defines the interface for decoding transaction bytes into a DecodedTx.
type Decoder interface {
	Decode(txBytes []byte) (*txdecode.DecodedTx, error)
}

// txApiDecoder is a function type that unmarshals transaction bytes into an API Tx type.
type txApiDecoder func(txBytes []byte) (*apitx.Tx, error)

// txApiEncoder is a function type that marshals a transaction into bytes.
type txApiEncoder func(tx *apitx.Tx) ([]byte, error)

// txDecoder decodes transaction bytes into an apitx.Tx structure.
func txDecoder(txBytes []byte) (*apitx.Tx, error) {
	var tx apitx.Tx
	return &tx, protov2.Unmarshal(txBytes, &tx)
}

// txEncoder encodes an apitx.Tx into bytes using protobuf marshaling options.
func txEncoder(tx *apitx.Tx) ([]byte, error) {
	return marshalOption.Marshal(tx)
}

// txJsonDecoder decodes transaction bytes into an apitx.Tx structure using JSON format.
func txJsonDecoder(txBytes []byte) (*apitx.Tx, error) {
	var tx apitx.Tx
	return &tx, protojson.Unmarshal(txBytes, &tx)
}

// txJsonEncoder encodes an apitx.Tx into bytes using JSON marshaling options.
func txJsonEncoder(tx *apitx.Tx) ([]byte, error) {
	return jsonMarshalOptions.Marshal(tx)
}
