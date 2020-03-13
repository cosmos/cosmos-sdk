package types

import (
	"bytes"

	jsonc "github.com/gibson042/canonicaljson-go"
	"github.com/gogo/protobuf/jsonpb"

	"github.com/cosmos/cosmos-sdk/codec"
)

// Register the sdk message type
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*Msg)(nil), nil)
	cdc.RegisterInterface((*Tx)(nil), nil)
}

// CanonicalSignBytes returns a canonical JSON encoding of a Proto message that
// can be signed over. The JSON encoding ensures all field names adhere to their
// Proto definition, default values are omitted, and follows the JSON Canonical
// Form.
func CanonicalSignBytes(m codec.ProtoMarshaler) ([]byte, error) {
	jm := &jsonpb.Marshaler{EmitDefaults: false, OrigName: false}
	buf := new(bytes.Buffer)

	// first, encode via canonical Protocol Buffer JSON
	if err := jm.Marshal(buf, m); err != nil {
		return nil, err
	}

	genericJSON := make(map[string]interface{})

	// decode canonical proto encoding into a generic map
	if err := jsonc.Unmarshal(buf.Bytes(), &genericJSON); err != nil {
		return nil, err
	}

	// finally, return the canonical JSON encoding via JSON Canonical Form
	return jsonc.Marshal(genericJSON)
}
