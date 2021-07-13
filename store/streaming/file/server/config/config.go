package config

import (
	"bytes"
	"text/template"

	"github.com/cosmos/cosmos-sdk/store/streaming/file"
	tmos "github.com/tendermint/tendermint/libs/os"
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

// StateFileServerConfig contains configuration parameters for the state file server
type StateFileServerConfig struct {
	GRPCAddress    string `mapstructure:"grpc-address"`
	GRPCWebEnabled bool   `mapstructure:"grpc-web-enabled"`
	GRPCWebAddress string `mapstructure:"grpc-web-address"`
	ChainID        string `mapstructure:"chain-id"`
	ReadDir        string `mapstructure:"read-dir"`
	FilePrefix     string `mapstructure:"file-prefix"`
	RemoveAfter    bool   `mapstructure:"remove-after"` // true: once data has been streamed forward it will be removed from the filesystem
	LogFile        string `mapstructure:"log-file"`
	LogLevel       string `mapstructure:"log-level"`
}

// DefaultStateFileServerConfig returns the reference to ClientConfig with default values.
func DefaultStateFileServerConfig() *StateFileServerConfig {
	return &StateFileServerConfig{
		GRPCAddress:    DefaultGRPCAddress,
		GRPCWebEnabled: true,
		GRPCWebAddress: DefaultGRPCWebAddress,
		ChainID:        "",
		ReadDir:        DefaultReadDir,
		FilePrefix:     "",
		RemoveAfter:    true,
		LogFile:        "",
		LogLevel:       "info",
	}
}

var configTemplate *template.Template

func init() {
	var err error

	tmpl := template.New("fileServerConfigFileTemplate")

	if configTemplate, err = tmpl.Parse(defaultFileServerConfigTemplate); err != nil {
		panic(err)
	}
}

// WriteConfigFile renders config using the template and writes it to
// configFilePath.
func WriteConfigFile(configFilePath string, config *StateFileServerConfig) {
	var buffer bytes.Buffer

	if err := configTemplate.Execute(&buffer, config); err != nil {
		panic(err)
	}

	tmos.MustWriteFile(configFilePath, buffer.Bytes(), 0644)
}

const defaultFileServerConfigTemplate = `# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

###############################################################################
###                           File Server Configuration                     ###
###############################################################################

[file-server]

# Address defines the gRPC server address to bind to.
grpc-address = "{{ .GRPCAddress }}"

# GRPCWebEnable defines if the gRPC-web should be enabled.
grpc-web-enable = {{ .GRPCWebEnabled }}

# Address defines the gRPC-web server address to bind to.
grpc-web-address = "{{ .GRPCWebAddress }}"

# ChainID defines the ChainID for the data we are serving
chain-id = "{{ .ChainID }}"

# ReadDir defines the directory to read the data we are serving
read-dir = "{{ .ReadDir }}"

# FilePrefix defines an (optional) prefix that has been prepended to the files from which we read data to serve
file-prefix = "{{ .ReadDir }}"

# RemoveAfter defines whether or not to remove files after reading and serving their data
remove-after = "{{ .RemoveAfter }}"

# LogFile defines the path to the file to write log messages to
log-file = "{{ .LogFile }}"

# LogLevel defines the log level- which kind of logs to write to output
log-level = "{{ .LogLevel }}"
`
