package cli

import "github.com/cosmos/cosmos-sdk/core/app_config"

type Config struct {
	DefaultAppConfig *app_config.AppConfig
	DefaultHomeDir   string
	DefaultEnvPrefix string
}

func Exec(config Config) {

}
