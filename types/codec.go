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

	var genericJSON interface{}

	// decode canonical proto encoding into a generic map
	if err := jsonc.Unmarshal(buf.Bytes(), &genericJSON); err != nil {
		return nil, err
	}

	// strip default values, i.e. 0, true, "", [], {}, null
	genericJSONNoDefaults := JSONStripDefaults(genericJSON)

	// finally, return the canonical JSON encoding via JSON Canonical Form
	return jsonc.Marshal(genericJSONNoDefaults)
}

func JSONStripDefaults(val interface{}) interface{} {
	switch val := val.(type) {
	case bool:
		if !val {
			return nil
		}
	case string:
		if val == "" {
			return nil
		}
	case float64:
		if val == 0 {
			return nil
		}
	case []interface{}:
		n := len(val)
		if n == 0 {
			return nil
		}
		res := make([]interface{}, n)
		for i, x := range val {
			res[i] = JSONStripDefaults(x)
		}
		return res
	case map[string]interface{}:
		res := make(map[string]interface{})
		for k, v := range val {
			v2 := JSONStripDefaults(v)
			if v2 != nil {
				res[k] = v2
			}
		}
		if len(res) == 0 {
			return nil
		}
		return res
	}
	return val
}
