package codec

import (
	"encoding/json"

	"github.com/gogo/protobuf/proto"
	"gopkg.in/yaml.v2"
)

// MarshalYAML marshals toPrint using jsonMarshaler to leverage specialized MarshalJSON methods
// (usually related to serialize data with protobuf or amin depending on a configuration).
// This involves additional roundtrip through JSON.
func MarshalYAML(jsonMarshaler JSONMarshaler, toPrint proto.Message) ([]byte, error) {
	// We are OK with the performance hit of the additional JSON roundtip. MarshalYAML is not
	// used in any critical parts of the system.
	bz, err := jsonMarshaler.MarshalJSON(toPrint)
	if err != nil {
		return nil, err
	}

	// generate YAML by decoding JSON and re-encoding to YAML
	var j interface{}
	err = json.Unmarshal(bz, &j)
	if err != nil {
		return nil, err
	}

	return yaml.Marshal(j)
}
