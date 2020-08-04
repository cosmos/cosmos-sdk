package codec

import (
	"encoding/json"

	"gopkg.in/yaml.v2"
)

func MarshalYAML(jsonMarshaler JSONMarshaler, toPrint interface{}) ([]byte, error) {
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
