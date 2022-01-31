package server

import (
	"context"
	"fmt"
	"reflect"

	tmlog "github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/container"
	"github.com/cosmos/cosmos-sdk/server/internal"
	"github.com/cosmos/cosmos-sdk/server/module"
)

func startModules(serverCtx *Context) (context.CancelFunc, error) {
	ctx := context.Background()
	ctx, cancelFn := context.WithCancel(ctx)

	var opts []container.Option
	for configSection, info := range internal.ModuleRegistry {
		if serverCtx.Viper.Get(configSection) == nil {
			continue
		}

		cfg := reflect.New(info.ConfigType)
		err := serverCtx.Viper.UnmarshalKey(configSection, cfg.Interface())
		if err != nil {
			return cancelFn, err
		}

		fmt.Printf("Loaded %s config: %+v\n", configSection, cfg)

		opt := info.Constructor(cfg.Elem().Interface())
		opts = append(opts, opt)
	}

	err := container.Run(func(services []module.Service) {
		for _, service := range services {
			svc := service
			if svc.Start != nil {
				go func() {
					err := svc.Start(ctx)
					if err != nil {
						fmt.Printf("Error starting service: %s", err)
					}
				}()
			}
		}
	},
		container.Options(opts...),
		container.Provide(func() tmlog.Logger { return serverCtx.Logger }),
	)

	return cancelFn, err
}
