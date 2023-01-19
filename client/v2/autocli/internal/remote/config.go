package remote

import (
	"encoding/json"
	"os"
	"path"

	"github.com/pkg/errors"
)

type Config struct {
	Chains map[string]*ChainConfig `json:"chains"`
}

type ChainConfig struct {
	TrustedGRPCEndpoints []string `json:"trusted_grpc_endpoints"`
}

func LoadConfig(configDir string) (*Config, error) {
	configPath := path.Join(configDir, "config.json")
	bz, err := os.ReadFile(configPath)
	if err != nil {
		return nil, errors.Wrapf(err, "can't read config file: %s", configPath)
	}

	config := &Config{}
	err = json.Unmarshal(bz, config)
	if err != nil {
		return nil, errors.Wrapf(err, "can't load config file: %s", configPath)
	}
	return config, err
}
