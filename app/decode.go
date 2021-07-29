package app

import (
	"github.com/gogo/protobuf/jsonpb"
	"sigs.k8s.io/yaml"
)

func ReadJSONConfig(bz []byte) (*Config, error) {
	var cfg Config
	err := jsonpb.UnmarshalString(string(bz), &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func ReadYAMLConfig(bz []byte) (*Config, error) {
	jsonBz, err := yaml.YAMLToJSON(bz)
	if err != nil {
		return nil, err
	}

	return ReadJSONConfig(jsonBz)
}
