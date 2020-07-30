package types

import (
	jsonc "github.com/gibson042/canonicaljson-go"

	"github.com/cosmos/cosmos-sdk/codec/types"

	"github.com/cosmos/cosmos-sdk/codec"
)

// Register the sdk message type
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*Msg)(nil), nil)
	cdc.RegisterInterface((*Tx)(nil), nil)
}

// Register the sdk message type
func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterInterface("cosmos_sdk.v1.Msg", (*Msg)(nil))
}

// CanonicalSignBytes returns a canonical JSON encoding of a Proto message that
// can be signed over. The JSON encoding ensures all field names adhere to their
// Proto definition, default values are omitted, and follows the JSON Canonical
// Form.
func CanonicalSignBytes(msg codec.ProtoMarshaler) ([]byte, error) {
	// first, encode via canonical Proto3 JSON
	bz, err := codec.ProtoMarshalJSON(msg)
	if err != nil {
		return nil, err
	}

	genericJSON := make(map[string]interface{})

	// decode canonical proto encoding into a generic map
	if err := jsonc.Unmarshal(bz, &genericJSON); err != nil {
		return nil, err
	}

	// finally, return the canonical JSON encoding via JSON Canonical Form
	return jsonc.Marshal(genericJSON)
}
