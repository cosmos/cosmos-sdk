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
`

// InitConfigTemplate initiates config template that will be used in
// WriteConfigFile
func initConfigTemplate() (*template.Template, error) {
	tmpl := template.New("clientConfigFileTemplate")
	configTemplate, err := tmpl.Parse(defaultConfigTemplate)
	if err != nil {
		return nil, err
	}

	return configTemplate, nil
}

// writeConfigFile renders config using the template and writes it to
// configFilePath.
func writeConfigFile(cfgFile string, config *ClientConfig, configTemplate *template.Template) error {
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
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	conf := new(ClientConfig)
	if err := v.Unmarshal(conf); err != nil {
		return nil, err
	}

	return conf, nil
}
