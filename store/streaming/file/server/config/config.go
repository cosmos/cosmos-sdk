package config

import (
	"github.com/cosmos/cosmos-sdk/store/streaming/file"
)

// Default constants
const (
	// DefaultReadDir defines the default directory to read the streamed files from
	DefaultReadDir        = file.DefaultWriteDir

	// DefaultGRPCAddress defines the default address to bind the gRPC server to.
	DefaultGRPCAddress = "0.0.0.0:9092"

	// DefaultGRPCWebAddress defines the default address to bind the gRPC-web server to.
	DefaultGRPCWebAddress = "0.0.0.0:9093"
)

type StateServerConfig struct {
	GRPCAddress string
	GRPCWebAddress string
}

// DefaultStateServerConfig returns the reference to ClientConfig with default values.
func DefaultStateServerConfig() *StateServerConfig {
	return &StateServerConfig{DefaultGRPCAddress, DefaultGRPCWebAddress}
}

type StateServerBackendConfig struct {
	ChainID string
	ReadDir string
	FilePrefix string
	Persist bool   // false: once data has been streamed forward it will be removed from the filesystem
}

// DefaultStateServerBackendConfig returns the reference to ClientConfig with default values.
func DefaultStateServerBackendConfig() *StateServerBackendConfig {
	return &StateServerBackendConfig{"", DefaultReadDir, "", true}
}
