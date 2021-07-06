package config

import (
	"github.com/cosmos/cosmos-sdk/store/streaming/file"
)

// Default constants
const (
	// DefaultReadDir defines the default directory to read the streamed files from
	DefaultReadDir = file.DefaultWriteDir

	// DefaultGRPCAddress defines the default address to bind the gRPC server to.
	DefaultGRPCAddress = "0.0.0.0:9092"

	// DefaultGRPCWebAddress defines the default address to bind the gRPC-web server to.
	DefaultGRPCWebAddress = "0.0.0.0:9093"
)

type StateServerConfig struct {
	GRPCAddress    string `mapstructure:"grpc-address"`
	GRPCWebAddress string `mapstructure:"grpc-web-address"`
	ChainID        string `mapstructure:"chain-id"`
	ReadDir        string `mapstructure:"read-dir"`
	FilePrefix     string `mapstructure:"file-prefix"`
	RemoveAfter    bool   `mapstructure:"remove-after"` // true: once data has been streamed forward it will be removed from the filesystem
}

// DefaultStateServerConfig returns the reference to ClientConfig with default values.
func DefaultStateServerConfig() *StateServerConfig {
	return &StateServerConfig{
		DefaultGRPCAddress,
		DefaultGRPCWebAddress,
		"",
		DefaultReadDir,
		"",
		true}
}

func WriteConfigFile(cfgFilePath string, cfg *StateServerConfig) {
	panic("implement me")
}
