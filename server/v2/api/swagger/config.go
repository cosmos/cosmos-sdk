package swagger

import (
    "cosmossdk.io/core/server"
)

const ServerName = "swagger"

type Config struct {
    Enable  bool   `toml:"enable" mapstructure:"enable"`
    Address string `toml:"address" mapstructure:"address"`
    Path    string `toml:"path" mapstructure:"path"`
}

func DefaultConfig() *Config {
    return &Config{
        Enable:  true,
        Address: "localhost:8080",
        Path:    "/swagger/",
    }
}

type CfgOption func(*Config)
