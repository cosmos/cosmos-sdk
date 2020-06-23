package config

import (
	"bytes"
	"text/template"

	"github.com/spf13/viper"
	tmos "github.com/tendermint/tendermint/libs/os"
)

const defaultConfigTemplate = `# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

###############################################################################
###                           Base Configuration                            ###
###############################################################################

# The minimum gas prices a validator is willing to accept for processing a
# transaction. A transaction's fees must meet the minimum of any denomination
# specified in this config (e.g. 0.25token1;0.0001token2).
minimum-gas-prices = "{{ .BaseConfig.MinGasPrices }}"

# default: the last 100 states are kept in addition to every 500th state; pruning at 10 block intervals
# nothing: all historic states will be saved, nothing will be deleted (i.e. archiving node)
# everything: all saved states will be deleted, storing only the current state; pruning at 10 block intervals
# custom: allow pruning options to be manually specified through 'pruning-keep-recent', 'pruning-keep-every', and 'pruning-interval'
pruning = "{{ .BaseConfig.Pruning }}"

# These are applied if and only if the pruning strategy is custom.
pruning-keep-recent = "{{ .BaseConfig.PruningKeepRecent }}"
pruning-keep-every = "{{ .BaseConfig.PruningKeepEvery }}"
pruning-interval = "{{ .BaseConfig.PruningInterval }}"

# HaltHeight contains a non-zero block height at which a node will gracefully
# halt and shutdown that can be used to assist upgrades and testing.
#
# Note: Commitment of state will be attempted on the corresponding block.
halt-height = {{ .BaseConfig.HaltHeight }}

# HaltTime contains a non-zero minimum block time (in Unix seconds) at which
# a node will gracefully halt and shutdown that can be used to assist upgrades
# and testing.
#
# Note: Commitment of state will be attempted on the corresponding block.
halt-time = {{ .BaseConfig.HaltTime }}

# InterBlockCache enables inter-block caching.
inter-block-cache = {{ .BaseConfig.InterBlockCache }}

###############################################################################
###                         Telemetry Configuration                         ###
###############################################################################

[telemetry]

# Prefixed with keys to separate services
service-name = "{{ .Telemetry.ServiceName }}"

# Enabled enables the application telemetry functionality. When enabled,
# an in-memory sink is also enabled by default. Operators may also enabled
# other sinks such as Prometheus.
enabled = {{ .Telemetry.Enabled }}

# Enable prefixing gauge values with hostname
enable-hostname = {{ .Telemetry.EnableHostname }}

# Enable adding hostname to labels
enable-hostname-label = {{ .Telemetry.EnableHostnameLabel }}

# Enable adding service to labels
enable-service-label = {{ .Telemetry.EnableServiceLabel }}

# PrometheusRetentionTime, when positive, enables a Prometheus metrics sink.
prometheus-retention-time = {{ .Telemetry.PrometheusRetentionTime }}

# GlobalLabels defines a global set of name/value label tuples applied to all
# metrics emitted using the wrapper functions defined in telemetry package.
#
# Example:
# [["chain_id", "cosmoshub-1"]]
global-labels = [{{ range $k, $v := .Telemetry.GlobalLabels }}
  ["{{index $v 0 }}", "{{ index $v 1}}"],{{ end }}
]

###############################################################################
###                           API Configuration                             ###
###############################################################################

[api]

# Enable defines if the API server should be enabled.
enable = {{ .API.Enable }}

# Swagger defines if swagger documentation should automatically be registered.
swagger = {{ .API.Swagger }}

# Address defines the API server to listen on
address = "{{ .API.Address }}"

# MaxOpenConnections defines the number of maximum open connections
max-open-connections = {{ .API.MaxOpenConnections }}

# RPCReadTimeout defines the Tendermint RPC read timeout (in seconds)
rpc-read-timeout = {{ .API.RPCReadTimeout }}

# RPCWriteTimeout defines the Tendermint RPC write timeout (in seconds)
rpc-write-timeout = {{ .API.RPCWriteTimeout }}

# RPCMaxBodyBytes defines the Tendermint maximum response body (in bytes)
rpc-max-body-bytes = {{ .API.RPCMaxBodyBytes }}

# EnableUnsafeCORS defines if CORS should be enabled (unsafe - use it at your own risk)
enabled-unsafe-cors = {{ .API.EnableUnsafeCORS }}
`

var configTemplate *template.Template

func init() {
	var err error

	tmpl := template.New("appConfigFileTemplate")

	if configTemplate, err = tmpl.Parse(defaultConfigTemplate); err != nil {
		panic(err)
	}
}

// ParseConfig retrieves the default environment configuration for the
// application.
func ParseConfig() (*Config, error) {
	conf := DefaultConfig()
	err := viper.Unmarshal(conf)

	return conf, err
}

// WriteConfigFile renders config using the template and writes it to
// configFilePath.
func WriteConfigFile(configFilePath string, config *Config) {
	var buffer bytes.Buffer

	if err := configTemplate.Execute(&buffer, config); err != nil {
		panic(err)
	}

	tmos.MustWriteFile(configFilePath, buffer.Bytes(), 0644)
}
