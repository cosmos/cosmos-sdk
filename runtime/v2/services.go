package runtime

import (
	"fmt"

	appmodulev2 "cosmossdk.io/core/appmodule/v2"
	"cosmossdk.io/core/transaction"

	gogoproto "github.com/cosmos/gogoproto/proto"
	"google.golang.org/protobuf/reflect/protoregistry"
)

func registerServices[T transaction.Tx](s appmodulev2.AppModule, app *App[T], registry *protoregistry.Files) error {
	c := &configurator{
		grpcQueryDecoders: map[string]func() gogoproto.Message{},
		stfQueryRouter:    app.queryRouterBuilder,
		stfMsgRouter:      app.msgRouterBuilder,
		registry:          registry,
		err:               nil,
	}

	if services, ok := s.(hasServicesV1); ok {
		if err := services.RegisterServices(c); err != nil {
			return fmt.Errorf("unable to register services: %w", err)
		}
	} else {
		// If module not implement RegisterServices, register msg & query handler.
		if module, ok := s.(appmodulev2.HasMsgHandlers); ok {
			module.RegisterMsgHandlers(app.msgRouterBuilder)
		}

		if module, ok := s.(appmodulev2.HasQueryHandlers); ok {
			module.RegisterQueryHandlers(app.queryRouterBuilder)
			// TODO: query regist by RegisterQueryHandlers not in grpcQueryDecoders
			if module, ok := s.(interface {
				GetQueryDecoders() map[string]func() gogoproto.Message
			}); ok {
				decoderMap := module.GetQueryDecoders()
				for path, decoder := range decoderMap {
					app.GRPCMethodsToMessageMap[path] = decoder
				}
			}
		}
	}

	if c.err != nil {
		app.logger.Warn("error registering services", "error", c.err)
	}

	// merge maps
	for path, decoder := range c.grpcQueryDecoders {
		app.GRPCMethodsToMessageMap[path] = decoder
	}

	return nil
}
