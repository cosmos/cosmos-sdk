package config

import (
	"bytes"
	"os"
	"path"
	"text/template"

	"github.com/spf13/viper"
	tmos "github.com/tendermint/tendermint/libs/os"
)

const defaultConfigTemplate = `# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

###############################################################################
###                           Client Configuration                            ###
###############################################################################


chain-id = "{{ .ClientConfig.ChainID }}"
keyring-backend = "{{ .ClientConfig.KeyringBackend }}"
output = "{{ .ClientConfig.Output }}"
node = "{{ .ClientConfig.Node }}"
broadcast-mode = "{{ .ClientConfig.BroadcastMode }}"
trace = "{{ .ClientConfig.Trace }}"
`

// InitConfigTemplate initiates config template that will be used in
// WriteConfigFile
func InitConfigTemplate() *template.Template {
	tmpl := template.New("clientConfigFileTemplate")
	configTemplate, err := tmpl.Parse(defaultConfigTemplate)
	if err != nil {
		panic(err)
	}

	return configTemplate
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
func WriteConfigFile(cfgFile string, config *ClientConfig, configTemplate *template.Template) {
	var buffer bytes.Buffer

	if err := configTemplate.Execute(&buffer, config); err != nil {
		panic(err)
	}

	tmos.MustWriteFile(cfgFile, buffer.Bytes(), 0644)
}

func ensureCfgPath(rootDir string) (string, error) {
	cfgPath := path.Join(rootDir, "config")
	if err := os.MkdirAll(cfgPath, os.ModePerm); err != nil { // config directory
		return "", err
	}

	return cfgPath, nil
}

func getClientConfig(cfgPath string, v *viper.Viper) (*ClientConfig, error) {
	v.AddConfigPath(cfgPath)
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
