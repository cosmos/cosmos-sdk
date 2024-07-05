package tx

import (
	"errors"
	"fmt"

	"google.golang.org/protobuf/encoding/protojson"
	protov2 "google.golang.org/protobuf/proto"

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
type txDecoder func(txBytes []byte) (Tx, error)

// txApiEncoder is a function type that marshals a transaction into bytes.
type txEncoder func(tx Tx) ([]byte, error)

// decodeTx decodes transaction bytes into an apitx.Tx structure.
func decodeTx(txBytes []byte) (Tx, error) {
	return nil, errors.New("not implemented")
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
func decodeJsonTx(txBytes []byte) (Tx, error) {
	return nil, errors.New("not implemented")
}

// encodeJsonTx encodes an apitx.Tx into bytes using JSON marshaling options.
func encodeJsonTx(tx Tx) ([]byte, error) {
	wTx, ok := tx.(*wrappedTx)
	if !ok {
		return nil, fmt.Errorf("unexpected tx type: %T", tx)
	}
	return jsonMarshalOptions.Marshal(wTx.Tx)
}
