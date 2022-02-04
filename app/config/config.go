package config

import (
	"sigs.k8s.io/yaml"

	configinternal "github.com/cosmos/cosmos-sdk/app/internal/config"

	"github.com/cosmos/cosmos-sdk/container"
)

func LoadJSON(bz []byte) container.Option {
	return configinternal.LoadJSON(bz)
}

func LoadYAML(bz []byte) container.Option {
	j, err := yaml.YAMLToJSON(bz)
	if err != nil {
		return container.Error(err)
	}
	return LoadJSON(j)
}
