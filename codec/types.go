package codec

import (
	"bytes"
	"encoding/json"
)

type Codec interface {
	MarshalJSON(interface{}) ([]byte, error)
	UnmarshalJSON([]byte, interface{}) error

	MarshalBinary(interface{}) ([]byte, error)
	UnmarshalBinary([]byte, interface{}) error
	MustMarshalBinary(interface{}) []byte
	MustUnmarshalBinary([]byte, interface{})
}

// attempt to make some pretty json
func MarshalJSONIndent(cdc Codec, obj interface{}) ([]byte, error) {
	bz, err := cdc.MarshalJSON(obj)
	if err != nil {
		return nil, err
	}

	var out bytes.Buffer
	err = json.Indent(&out, bz, "", "  ")
	if err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
