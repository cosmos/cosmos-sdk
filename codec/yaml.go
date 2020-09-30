package codec

import (
	"encoding/json"

	"github.com/gogo/protobuf/proto"
	"gopkg.in/yaml.v2"
)

// MarshalYAML marshals the provided toPrint content with the provided JSON marshaler
// by encoding JSON, decoding JSON, and then encoding YAML.
func MarshalYAML(jsonMarshaler JSONMarshaler, toPrint proto.Message) ([]byte, error) {
	// only the JSONMarshaler has full context as to how the JSON
	// mashalling should look (which may be different for amino & proto codecs)
	// so we need to use it to marshal toPrint first
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

	return yaml.Marshal(j)
}
