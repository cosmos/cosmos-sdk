package config

import (
	"bytes"
	"text/template"
	"os"
	"path"
	

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

var configTemplate *template.Template

func InitConfigTemplate() {
	var err error

	tmpl := template.New("clientConfigFileTemplate")

	if configTemplate, err = tmpl.Parse(defaultConfigTemplate); err != nil {
		panic(err)
	}

}

// ParseConfig retrieves the default environment configuration for the
// application.
func ParseConfig() *ClientConfig {
	conf := DefaultClientConfig()
	 _ = viper.Unmarshal(conf)

	return conf
}

// WriteConfigFile renders config using the template and writes it to
// configFilePath.
func WriteConfigFile(configFilePath string, config *ClientConfig) {
	var buffer bytes.Buffer

	if err := configTemplate.Execute(&buffer, config); err != nil {
		panic(err)
	}

	tmos.MustWriteFile(configFilePath, buffer.Bytes(), 0644)
}

func ensureConfFile(rootDir string) (string, error) {
	cfgPath := path.Join(rootDir, "config")
	if err := os.MkdirAll(cfgPath, os.ModePerm); err != nil { // config directory
		return "", err
	}

	return cfgPath, nil
}

func getClientConfig(cfgPath string) (*ClientConfig, error) {
	viper.AddConfigPath(cfgPath)
	viper.SetConfigName("client")
	viper.SetConfigType("toml")
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	conf := new(ClientConfig)
	if err := viper.Unmarshal(conf); err != nil {
		return nil, err
	}

	return conf,nil
}
