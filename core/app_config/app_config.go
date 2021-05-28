package app_config

import (
	"github.com/tendermint/tendermint/abci/types"
)

func Compose(config AppConfig) (types.Application, error) {
	//moduleSet := module.NewModuleSet(config.Modules)

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
