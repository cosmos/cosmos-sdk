package config

import (
	"sigs.k8s.io/yaml"

	"github.com/cosmos/cosmos-sdk/container"
)

func LoadYAML(bz []byte) container.Option {
	j, err := yaml.YAMLToJSON(bz)
	if err != nil {
		return container.Error(err)
	}
	return LoadJSON(j)
}

func LoadJSON(bz []byte) container.Option {
	panic("TODO")
}
