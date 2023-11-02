package serverv2

import "github.com/spf13/viper"

type Config struct {
	*viper.Viper
}

func NewConfig() *Config {
	return &Config{
		Viper: viper.New(),
	}
}
