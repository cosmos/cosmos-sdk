package app

import (
	"encoding/json"

	"github.com/gogo/protobuf/proto"

	"gopkg.in/yaml.v2"
)

func ReadJSONConfig(bz []byte) (*Config, error) {
	var cfg Config
	err := proto.Unmarshal(bz, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func ReadYAMLConfig(bz []byte) (*Config, error) {
	var cfg map[string]interface{}
	err := yaml.Unmarshal(bz, &cfg)
	if err != nil {
		return nil, err
	}

	jsonBz, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}

	return ReadJSONConfig(jsonBz)
}
