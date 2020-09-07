package codec

import (
	"encoding/json"

	"github.com/gogo/protobuf/proto"
	"gopkg.in/yaml.v2"
)

// MarshalYAML marshals the provided toPrint content with the provided JSON marshaler
// by encoding JSON, decoding JSON, and then encoding YAML.
func MarshalYAML(jsonMarshaler JSONMarshaler, toPrint proto.Message) ([]byte, error) {
	bz, err := jsonMarshaler.MarshalJSON(toPrint)
	if err != nil {
		return nil, err
	}

	// generate YAML by decoding and re-encoding JSON as YAML
	var j interface{}

	err = json.Unmarshal(bz, &j)
	if err != nil {
		return nil, err
	}

	bz, err = yaml.Marshal(j)
	if err != nil {
		return nil, err
	}

	return bz, nil
}
