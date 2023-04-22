package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/spf13/viper"
)

const (
	defaultConfigTemplate = `# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

###############################################################################
###                           Client Configuration                            ###
###############################################################################

# The network chain ID
chain-id = "{{ .ChainID }}"
# The keyring's backend, where the keys are stored (os|file|kwallet|pass|test|memory)
keyring-backend = "{{ .KeyringBackend }}"
# CLI output format (text|json)
output = "{{ .Output }}"
# <host>:<port> to CometBFT RPC interface for this chain
node = "{{ .Node }}"
# Transaction broadcasting mode (sync|async)
broadcast-mode = "{{ .BroadcastMode }}"
`

	defaultHomeTemplate = `# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

###############################################################################
###                            Home Configuration                             ###
###############################################################################

# The currently configured home directory
home-dir = "{{ .HomeDir }}"`
)

// writeConfigToFile parses defaultConfigTemplate, renders config using the template and writes it to
// configFilePath.
func writeConfigToFile(configFilePath string, config *ClientConfig) error {
	var buffer bytes.Buffer

	tmpl := template.New("clientConfigFileTemplate")
	configTemplate, err := tmpl.Parse(defaultConfigTemplate)
	if err != nil {
		return err
	}

	// Check if config file exists
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		fmt.Println("creating new node configuration in", filepath.Dir(filepath.Dir(configFilePath)))
	}
	if err := configTemplate.Execute(&buffer, config); err != nil {
		return err
	}

	return os.WriteFile(configFilePath, buffer.Bytes(), 0o600)
}

// HomeDirConfig is a struct that contains the home directory and defines the mapping
// for the TOML writer.
type HomeDirConfig struct {
	HomeDir string `mapstructure:"home-dir" json:"home-dir"`
}

// WriteHomeDirToFile parses homeDirTemplate, renders the template with the given new home directory
// and writes it to the given file.
func WriteHomeDirToFile(filePath, homeDir string) error {
	var buffer bytes.Buffer

	tmpl := template.New("homeDirFileTemplate")
	homeTemplate, err := tmpl.Parse(defaultHomeTemplate)
	if err != nil {
		return err
	}

	homeDirConfig := HomeDirConfig{HomeDir: homeDir}
	if err := homeTemplate.Execute(&buffer, homeDirConfig); err != nil {
		return err
	}
	return os.WriteFile(filePath, buffer.Bytes(), 0o600)
}

// ReadHomeDir tries to return the currently stored home directory from the
// given config directory.
func ReadHomeDir(configPath string, v *viper.Viper) (string, error) {
	homeDirConfig := HomeDirConfig{}
	v.AddConfigPath(configPath)
	v.SetConfigName("home")
	v.SetConfigType("toml")

	homeFilePath := filepath.Join(configPath, "home.toml")
	if _, err := os.Stat(homeFilePath); os.IsNotExist(err) {
		folder := filepath.Dir(configPath)
		if err := WriteHomeDirToFile(homeFilePath, folder); err != nil {
			return "", err
		}
	}

	if err := v.ReadInConfig(); err != nil {
		return "", err
	}
	if err := v.Unmarshal(&homeDirConfig); err != nil {
		return "", err
	}
	return homeDirConfig.HomeDir, nil
}

// getClientConfig reads values from client.toml file and unmarshalls them into ClientConfig
func getClientConfig(configPath string, v *viper.Viper) (*ClientConfig, error) {
	v.AddConfigPath(configPath)
	v.SetConfigName("client")
	v.SetConfigType("toml")

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	conf := DefaultConfig()
	if err := v.Unmarshal(conf); err != nil {
		return nil, err
	}

	return conf, nil
}
