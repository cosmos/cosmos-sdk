package types

import (
	"bytes"
	"encoding/json"
	"github.com/cosmos/cosmos-sdk/codec"
	jsonc "github.com/gibson042/canonicaljson-go"
	"github.com/gogo/protobuf/jsonpb"
	"reflect"
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
	if err := json.Unmarshal(buf.Bytes(), &genericJSON); err != nil {
		return nil, err
	}

	// strip default values, i.e. 0, true, "", [], {}, null
	genericJSONNoDefaults := jsonStripDefaults(genericJSON)

	// finally, return the canonical JSON encoding via JSON Canonical Form
	return jsonc.Marshal(genericJSONNoDefaults)
}

// jsonStripDefaults removes default and nil values from maps within JSON
// that has been parsed into a generic interface{} using json.Unmarshal.
// The returned interface{} value can then be marshaled back to JSON.
func jsonStripDefaults(val interface{}) interface{} {
	switch val := val.(type) {
	case []interface{}:
		n := len(val)
		if n == 0 {
			return nil
		}
		res := make([]interface{}, n)
		for i, x := range val {
			switch x := x.(type) {
			case map[string]interface{}:
				res[i] = jsonStripDefaultMapKeys(x)
			default:
				res[i] = x
			}
		}
		return res
	case map[string]interface{}:
		res := jsonStripDefaultMapKeys(val)
		if len(res) == 0 {
			return nil
		}
		return res
	default:
		if isZeroVal(val) {
			return nil
		}
	}
	return val
}

func isZeroVal(x interface{}) bool {
	return x == nil || reflect.DeepEqual(x, reflect.Zero(reflect.TypeOf(x)).Interface())
}

func jsonStripDefaultMapKeys(val map[string]interface{}) map[string]interface{} {
	res := make(map[string]interface{})
	for k, v := range val {
		v2 := jsonStripDefaults(v)
		if v2 != nil {
			res[k] = v2
		}
	}
	return res
}
