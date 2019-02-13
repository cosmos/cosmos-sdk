package config

import (
	"bytes"
	"text/template"

	"github.com/spf13/viper"
	cmn "github.com/tendermint/tendermint/libs/common"
)

const defaultConfigTemplate = `# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

##### main base config options #####

# The minimum gas prices a validator is willing to accept for processing a
# transaction. A transaction's fees must meet the minimum of any denomination
# specified in this config (e.g. 0.01photino;0.0001stake).
minimum-gas-prices = "{{ .BaseConfig.MinGasPrices }}"
`

var configTemplate *template.Template

func init() {
	var err error
	tmpl := template.New("gaiaConfigFileTemplate")
	if configTemplate, err = tmpl.Parse(defaultConfigTemplate); err != nil {
		panic(err)
	}
}

// ParseConfig retrieves the default environment configuration for Gaia.
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

	cmn.MustWriteFile(configFilePath, buffer.Bytes(), 0644)
}
