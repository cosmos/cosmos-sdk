package tx

import (
	"google.golang.org/protobuf/encoding/protojson"
	protov2 "google.golang.org/protobuf/proto"

	apitx "cosmossdk.io/api/cosmos/tx/v1beta1"
)

var (
	marshalOption      = protov2.MarshalOptions{Deterministic: true}
	jsonMarshalOptions = protojson.MarshalOptions{
		Indent:         "",
		UseProtoNames:  true,
		UseEnumNumbers: false,
	}
)

// txApiDecoder unmarshals transaction bytes into API Tx type
type txApiDecoder func(txBytes []byte) (*apitx.Tx, error)

// txApiEncoder marshals transaction to bytes
type txApiEncoder func(tx *apitx.Tx) ([]byte, error)

func txDecoder(txBytes []byte) (*apitx.Tx, error) {
	var tx apitx.Tx
	return &tx, protov2.Unmarshal(txBytes, &tx)
}

func txEncoder(tx *apitx.Tx) ([]byte, error) {
	return marshalOption.Marshal(tx)
}

func txJsonDecoder(txBytes []byte) (*apitx.Tx, error) {
	var tx apitx.Tx
	return &tx, protojson.Unmarshal(txBytes, &tx)
}

func txJsonEncoder(tx *apitx.Tx) ([]byte, error) {
	return jsonMarshalOptions.Marshal(tx)
}
