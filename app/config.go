package app

import (
	"encoding/json"
	"reflect"
	"strings"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/baseapp"

	"github.com/gogo/protobuf/proto"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"

	"github.com/cosmos/cosmos-sdk/container"
	"github.com/cosmos/cosmos-sdk/types/module"
)

func ProvideAppConfig(config *Config) container.Option {
	return container.Options(
		container.OnePerScopeTypes(
			reflect.TypeOf((*module.AppModule)(nil)).Elem(),
		),
		composeModules(config),
		container.Provide(func(modules map[string]module.AppModule, jsonCodec codec.JSONCodec) func(*baseapp.BaseApp) {
			return func(app *baseapp.BaseApp) {
				app.SetInitChainer(func(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
					var genesisState map[string]json.RawMessage
					if err := json.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
						panic(err)
					}

					var validatorUpdates []abci.ValidatorUpdate
					for _, moduleName := range config.Abci.InitGenesis {
						if genesisState[moduleName] == nil {
							continue
						}

						moduleValUpdates := modules[moduleName].InitGenesis(ctx, jsonCodec, genesisState[moduleName])

						// use these validator updates if provided, the module manager assumes
						// only one module will update the validator set
						if len(moduleValUpdates) > 0 {
							if len(validatorUpdates) > 0 {
								panic("validator InitGenesis updates already set by a previous module")
							}
							validatorUpdates = moduleValUpdates
						}
					}

					return abci.ResponseInitChain{
						Validators: validatorUpdates,
					}
				})

				app.SetBeginBlocker(func(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
					ctx = ctx.WithEventManager(sdk.NewEventManager())

					for _, moduleName := range config.Abci.BeginBlock {
						modules[moduleName].BeginBlock(ctx, req)
					}

					return abci.ResponseBeginBlock{
						Events: ctx.EventManager().ABCIEvents(),
					}
				})

				app.SetEndBlocker(func(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
					ctx = ctx.WithEventManager(sdk.NewEventManager())
					var validatorUpdates []abci.ValidatorUpdate

					for _, moduleName := range config.Abci.EndBlock {
						moduleValUpdates := modules[moduleName].EndBlock(ctx, req)

						// use these validator updates if provided, the module manager assumes
						// only one module will update the validator set
						if len(moduleValUpdates) > 0 {
							if len(validatorUpdates) > 0 {
								panic("validator EndBlock updates already set by a previous module")
							}

							validatorUpdates = moduleValUpdates
						}
					}

					return abci.ResponseEndBlock{
						ValidatorUpdates: validatorUpdates,
						Events:           ctx.EventManager().ABCIEvents(),
					}
				})
			}
		}),
	)
}

func ProvideAppConfigJSON(json []byte) container.Option {
	cfg, err := ReadJSONConfig(json)
	if err != nil {
		return container.Error(err)
	}
	return ProvideAppConfig(cfg)
}

func ProvideAppConfigYAML(yaml []byte) container.Option {
	cfg, err := ReadYAMLConfig(yaml)
	if err != nil {
		return container.Error(err)
	}
	return ProvideAppConfig(cfg)
}

func composeModules(config *Config) container.Option {
	var opts []container.Option
	for _, mod := range config.Modules {
		opts = append(opts, addProtoModule(mod.Name, mod.Config))
	}
	return container.Options(opts...)
}

func addProtoModule(name string, config *codectypes.Any) container.Option {
	// unpack Any
	msgTyp := proto.MessageType(config.TypeUrl)
	mod := reflect.New(msgTyp.Elem()).Interface().(proto.Message)
	if err := proto.Unmarshal(config.Value, mod); err != nil {
		return container.Error(err)
	}

	return addModule(name, mod)
}

func addModule(name string, mod interface{}) container.Option {
	var opts []container.Option

	if typeProvider, ok := mod.(Module); ok {
		opts = append(opts, container.Provide(func() func(registry codectypes.TypeRegistry) {
			return func(registry codectypes.TypeRegistry) {
				typeProvider.RegisterTypes(registry)
			}
		}))
	}

	// register DI Provide* methods
	modTy := reflect.TypeOf(mod)
	numMethods := modTy.NumMethod()
	for i := 0; i < numMethods; i++ {
		method := modTy.Method(i)
		if strings.HasPrefix(method.Name, "Provide") || strings.HasPrefix(method.Name, "provide") {
			ctrInfo, err := container.ExtractConstructorInfo(method.Func.Interface())
			if err != nil {
				return container.Error(err)
			}

			ctrInfo.In = ctrInfo.In[1:len(ctrInfo.In)]
			ctrInfo.Fn = func(values []reflect.Value) ([]reflect.Value, error) {
				return ctrInfo.Fn(append([]reflect.Value{reflect.ValueOf(mod)}, values...))
			}

			opts = append(opts, container.ProvideWithScope(name, ctrInfo))
		}
	}

	return container.Options(opts...)
}
