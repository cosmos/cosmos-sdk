package codec

import (
	"bytes"
	"encoding/json"

	"github.com/tendermint/go-amino"
)

type Codec interface {
	MarshalJSON(interface{}) ([]byte, error)
	MarshalJSONIndent(interface{}, string, string) ([]byte, error)
	UnmarshalJSON([]byte, interface{}) error

	MarshalBinary(interface{}) ([]byte, error)
	UnmarshalBinary([]byte, interface{}) error
	MustMarshalBinary(interface{}) []byte
	MustUnmarshalBinary([]byte, interface{})
	MarshalBinaryBare(interface{}) ([]byte, error)
	UnmarshalBinaryBare([]byte, interface{}) error

	Seal() Codec

	RegisterInterface(interface{}, *amino.InterfaceOptions)
	RegisterConcrete(interface{}, string, *amino.ConcreteOptions)
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
