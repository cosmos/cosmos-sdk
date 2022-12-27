package viperutils

import (
	"bytes"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"path"
	"strings"
	"text/template"
)

type ConfigFileConfig struct {
	Path          string
	Template      string
	DefaultValues any
}

func InitiateViper(v *viper.Viper, cmd *cobra.Command, environmentPrefix string, configFiles ...ConfigFileConfig) error {
	if err := bindAllFlags(v, cmd); err != nil {
		return err
	}

	bindEnvironment(v, environmentPrefix)

	for _, config := range configFiles {
		if err := readConfigFile(v, config); err != nil {
			return err
		}
	}

	return nil
}

func bindAllFlags(v *viper.Viper, cmd *cobra.Command) error {
	if err := v.BindPFlags(cmd.Flags()); err != nil {
		return err
	}
	if err := v.BindPFlags(cmd.PersistentFlags()); err != nil {
		return err
	}

	return nil
}

func bindEnvironment(v *viper.Viper, environmentPrefix string) {
	v.SetEnvPrefix(environmentPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.AutomaticEnv()
}

func readConfigFile(v *viper.Viper, config ConfigFileConfig) error {
	// if the config file does not exist we create it and write the default values into it.
	if _, err := os.Stat(config.Path); os.IsNotExist(err) {
		dir := path.Dir(config.Path)
		if err := ensureConfigPath(dir); err != nil {
			return fmt.Errorf("couldn't make client config: %v", err)
		}
		if err := WriteConfigToFile(config.Path, config.Template, config.DefaultValues); err != nil {
			return fmt.Errorf("could not write client config to the file: %v", err)
		}
	}

	v.SetConfigFile(config.Path)
	return v.MergeInConfig()
}

// WriteConfigToFile parses tmpl, renders config using the configTemplate and writes it to
// configFilePath.
func WriteConfigToFile(configFilePath string, configTemplate string, config any) error {
	var buffer bytes.Buffer

	tmpl := template.New(configFilePath)
	tmpl, err := tmpl.Parse(configTemplate)
	if err != nil {
		return err
	}

	if err := tmpl.Execute(&buffer, config); err != nil {
		return err
	}

	return os.WriteFile(configFilePath, buffer.Bytes(), 0o600)
}

// ensureConfigPath creates a directory configPath if it does not exist
func ensureConfigPath(configPath string) error {
	return os.MkdirAll(configPath, os.ModePerm)
}

func GetConfig[T any](v *viper.Viper) (T, error) {
	var conf T
	if err := v.Unmarshal(&conf); err != nil {
		return conf, fmt.Errorf("couldn't get config: %v", err)
	}

	return conf, nil
}
