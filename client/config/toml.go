package config

import (
	"bytes"
	"os"
	"text/template"

	"github.com/spf13/viper"
	tmos "github.com/tendermint/tendermint/libs/os"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
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

// initConfigTemplate initiates config template that will be used in
// writeConfigFile
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
func writeConfigFile(configFilePath string, config *ClientConfig, configTemplate *template.Template) error {
	var buffer bytes.Buffer

	if err := configTemplate.Execute(&buffer, config); err != nil {
		return err
	}

	tmos.MustWriteFile(configFilePath, buffer.Bytes(), 0644)
	return nil

}

// ensureConfigPath creates a directory configPath if it does not exist
func ensureConfigPath(configPath string) error {
	if err := os.MkdirAll(configPath, os.ModePerm); err != nil {
		return err
	}

	return nil
}

// newClientFromNodeFlag sets up Client implementation that communicates with a Tendermint node over
// JSON RPC and WebSockets
func newClientFromNodeFlag(nodeURI string) (*rpchttp.HTTP, error) {
	return rpchttp.New(nodeURI, "/websocket")
}

// getClientConfig reads values from client.toml file and unmarshalls them into ClientConfig
func getClientConfig(configPath string, v *viper.Viper) (*ClientConfig, error) {

	v.AddConfigPath(configPath)
	v.SetConfigName("client")
	v.SetConfigType("toml")

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	conf := new(ClientConfig)
	if err := v.Unmarshal(conf); err != nil {
		return nil, err
	}

	return conf, nil
}
