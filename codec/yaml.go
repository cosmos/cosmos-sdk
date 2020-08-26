package codec

import (
	"encoding/json"

	"gopkg.in/yaml.v2"
)

// MarshalYAML marshals the provided bytes content by decoding the bytes to
// JSON, and then encoding YAML.
func MarshalYAML(bz []byte) ([]byte, error) {
	// generate YAML by decoding and re-encoding JSON as YAML
	var j interface{}

	err := json.Unmarshal(bz, &j)
	if err != nil {
		return nil, err
	}

	bz, err = yaml.Marshal(j)
	if err != nil {
		return nil, err
	}

	return bz, nil
}
