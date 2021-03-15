package config

import (
	"bytes"
	"os"
	"text/template"

	"github.com/spf13/viper"
	tmos "github.com/tendermint/tendermint/libs/os"
)

const defaultConfigTemplate = `# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

###############################################################################
###                           Client Configuration                            ###
###############################################################################


chain-id = "{{ .ChainID }}"
keyring-backend = "{{ .KeyringBackend }}"
output = "{{ .Output }}"
node = "{{ .Node }}"
broadcast-mode = "{{ .BroadcastMode }}"
trace = "{{ .Trace }}"
`

// InitConfigTemplate initiates config template that will be used in
// WriteConfigFile
func InitConfigTemplate() (*template.Template, error) {
	tmpl := template.New("clientConfigFileTemplate")
	configTemplate, err := tmpl.Parse(defaultConfigTemplate)
	if err != nil {
		return nil, err
	}

	return configTemplate, nil
}

// ParseConfig retrieves the default environment configuration for the
// application.
func ParseConfig(v *viper.Viper) (*ClientConfig, error) {
	conf := DefaultClientConfig()
	if err := v.Unmarshal(conf); err != nil {
		return nil, err
	}

	return conf, nil
}

// WriteConfigFile renders config using the template and writes it to
// configFilePath.
func WriteConfigFile(cfgFile string, config *ClientConfig, configTemplate *template.Template) error {
	var buffer bytes.Buffer

	if err := configTemplate.Execute(&buffer, config); err != nil {
		return err
	}

	tmos.MustWriteFile(cfgFile, buffer.Bytes(), 0644)
	return nil

}

func ensureConfigPath(configPath string) error {
	if err := os.MkdirAll(configPath, os.ModePerm); err != nil {
		return err
	}

	return nil
}

func getClientConfig(configPath string, v *viper.Viper) (*ClientConfig, error) {

	v.AddConfigPath(configPath)
	v.SetConfigName("client")
	v.SetConfigType("toml")

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	conf := DefaultClientConfig()
	if err := v.Unmarshal(conf); err != nil {
		return nil, err
	}

	return conf, nil
}
