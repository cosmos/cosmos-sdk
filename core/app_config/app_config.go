package app_config

import (
	"context"
	"fmt"
	"reflect"

	"github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/core/abci"

	"github.com/cosmos/cosmos-sdk/core/container"

	"github.com/gogo/protobuf/proto"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/core/module"
	"github.com/cosmos/cosmos-sdk/core/module/app"
)

func Compose(config AppConfig, moduleRegistry *module.Registry) (abci.Handler, error) {
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cont := container.NewContainer()
	modSet := &moduleSet{
		container: cont,
		modMap:    map[string]app.Handler{},
		configMap: map[string]*ModuleConfig{},
	}

	for _, mod := range config.Modules {
		err := modSet.addModule(interfaceRegistry, moduleRegistry, mod)
		if err != nil {
			return nil, err
		}
	}

	err := modSet.addModule(interfaceRegistry, moduleRegistry, &ModuleConfig{
		Module: config.Abci.TxModule,
		Name:   "tx",
	})
	if err != nil {
		return nil, err
	}

	err = modSet.initialize()
	if err != nil {
		return nil, err
	}

	//appModules := make(map[string]app.Module)
	//moduleSet.Each(func(name string, handler module.ModuleHandler) {
	//	// TODO
	//})

	//bapp := &baseapp.baseApp{}
	//
	//for _, m := range config.Abci.BeginBlock {
	//	bapp.beginBlockers = append(bapp.beginBlockers, appModules[m].(app.BeginBlocker))
	//}
	//
	//for _, m := range config.Abci.EndBlock {
	//	bapp.endBlockers = append(bapp.endBlockers, appModules[m].(app.EndBlocker))
	//}
	//
	//return bapp, nil

	panic("TODO")
}

type moduleSet struct {
	container *container.Container
	modMap    map[string]app.Handler
	configMap map[string]*ModuleConfig
}

func (ms *moduleSet) Info(ctx context.Context, req types.RequestInfo) types.ResponseInfo {
	panic("implement me")
}

func (ms *moduleSet) SetOption(ctx context.Context, req types.RequestSetOption) types.ResponseSetOption {
	panic("implement me")
}

func (ms *moduleSet) Query(ctx context.Context, req types.RequestQuery) types.ResponseQuery {
	panic("implement me")
}

func (ms *moduleSet) CheckTx(ctx context.Context, req types.RequestCheckTx) types.ResponseCheckTx {
	panic("implement me")
}

func (ms *moduleSet) InitChain(ctx context.Context, req types.RequestInitChain) types.ResponseInitChain {
	panic("implement me")
}

func (ms *moduleSet) BeginBlock(ctx context.Context, req types.RequestBeginBlock) types.ResponseBeginBlock {
	panic("implement me")
}

func (ms *moduleSet) DeliverTx(ctx context.Context, req types.RequestDeliverTx) types.ResponseDeliverTx {
	panic("implement me")
}

func (ms *moduleSet) EndBlock(ctx context.Context, req types.RequestEndBlock) types.ResponseEndBlock {
	panic("implement me")
}

func (ms *moduleSet) Commit(ctx context.Context) types.ResponseCommit {
	panic("implement me")
}

func (ms *moduleSet) ListSnapshots(ctx context.Context, req types.RequestListSnapshots) types.ResponseListSnapshots {
	panic("implement me")
}

func (ms *moduleSet) OfferSnapshot(ctx context.Context, req types.RequestOfferSnapshot) types.ResponseOfferSnapshot {
	panic("implement me")
}

func (ms *moduleSet) LoadSnapshotChunk(ctx context.Context, req types.RequestLoadSnapshotChunk) types.ResponseLoadSnapshotChunk {
	panic("implement me")
}

func (ms *moduleSet) ApplySnapshotChunk(ctx context.Context, req types.RequestApplySnapshotChunk) types.ResponseApplySnapshotChunk {
	panic("implement me")
}

var _ abci.Handler = &moduleSet{}

func (ms *moduleSet) addModule(interfaceRegistry codectypes.InterfaceRegistry, registry *module.Registry, config *ModuleConfig) error {
	ms.configMap[config.Name] = config

	msg, err := interfaceRegistry.Resolve(config.Module.TypeUrl)
	if err != nil {
		return err
	}

	// TODO:
	//typeProvider, ok := msg.(codec.TypeProvider)
	//if !ok {
	//  typeProvider.RegisterTypes(interfaceRegistry)
	//}

	err = proto.Unmarshal(config.Module.Value, msg)
	if err != nil {
		return err
	}

	ctr := registry.Resolve(msg)
	if ctr == nil {
		return fmt.Errorf("TODO")
	}

	ctrVal := reflect.ValueOf(ctr)
	ctrTyp := ctrVal.Type()

	numIn := ctrTyp.NumIn()
	var needs []container.Input
	for i := 1; i < numIn; i++ {
		argTy := ctrTyp.In(i)
		needs = append(needs, container.Input{
			Key: container.Key{
				Type: argTy,
			},
		})
	}

	numOut := ctrTyp.NumIn()
	var provides []container.Output
	for i := 1; i < numOut; i++ {
		argTy := ctrTyp.Out(i)

		// check if is error type
		if isErrorTyp(argTy) {
			continue
		}

		provides = append(provides,
			container.Output{
				Key: container.Key{
					Type: argTy,
				},
			},
		)
	}

	return ms.container.RegisterProvider(container.Provider{
		Constructor: func(deps []reflect.Value, _ container.Scope) ([]reflect.Value, error) {
			args := []reflect.Value{reflect.ValueOf(msg)}
			args = append(args, deps...)
			res := ctrVal.Call(args)
			if len(res) < 1 {
				return nil, fmt.Errorf("expected at least one return value")
			}

			handler, ok := res[0].Interface().(app.Handler)
			if !ok {
				return nil, fmt.Errorf("expected handler got %+v", res[0])
			}

			var provides []reflect.Value
			for i := 1; i < len(res); i++ {
				if isErrorTyp(res[i].Type()) {
					continue
				}

				provides = append(provides, res[i])
			}

			ms.modMap[config.Name] = handler

			return provides, nil
		},
		Needs:    needs,
		Provides: provides,
		Scope:    container.Scope(config.Name),
	})
}

func (ms *moduleSet) initialize() error {
	err := ms.container.InitializeAll()
	if err != nil {
		return err
	}

	for name := range ms.configMap {
		if ms.modMap[name] == nil {
			return fmt.Errorf("module %s failed to initialize", name)
		}
	}

	return nil
}

func isErrorTyp(ty reflect.Type) bool {
	return ty.Implements(reflect.TypeOf((*error)(nil)).Elem())
}
