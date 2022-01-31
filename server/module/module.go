package module

import (
	"context"
	"fmt"
	"reflect"

	tmlog "github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/server/internal"

	"github.com/cosmos/cosmos-sdk/container"
)

func Register(
	configSection string,
	configType interface{},
	constructor func(config interface{}) container.Option,
	_ ...Option,
) {
	if _, ok := internal.ModuleRegistry[configSection]; ok {
		panic(fmt.Errorf("config section %s already defined", configSection))
	}

	internal.ModuleRegistry[configSection] = &internal.ModuleInfo{
		ConfigSection: configSection,
		ConfigType:    reflect.TypeOf(configType),
		Constructor:   constructor,
	}
}

type Option interface{ isOption() }

type Service struct {
	Start func(context.Context) (err error)
}

var _ container.AutoGroupType = Service{}

func (s Service) IsAutoGroupType() {}

type TestConfig struct {
	Foo int    `mapstructure:"foo"`
	Bar string `mapstructure:"bar"`
}

func init() {
	Register("wam", TestConfig{}, func(config interface{}) container.Option {
		cfg := config.(TestConfig)
		return container.Provide(func(logger tmlog.Logger) Service {
			return Service{
				Start: func(ctx context.Context) error {
					logger.Info(fmt.Sprintf("Starting Foo:%d Bar:%s", cfg.Foo, cfg.Bar))
					return nil
				},
			}
		})
	})
}
