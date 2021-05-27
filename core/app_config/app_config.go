package app_config

import (
	"github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/core/module"
	"github.com/cosmos/cosmos-sdk/core/module/app"
)

func Compose(config AppConfig) (types.Application, error) {
	moduleSet := module.NewModuleSet(config.Modules)

	appModules := make(map[string]app.Module)
	moduleSet.Each(func(name string, handler module.ModuleHandler) {
		// TODO
	})

	bapp := &baseApp{}

	var beginBlockers []app.BeginBlocker
	for _, m := range config.Abci.BeginBlock {
		mod, ok := appModules[m]
		if !ok {
			panic("TODO")
		}

		beginBlocker, ok := mod.(app.BeginBlocker)
		if !ok {
			panic("TODO")
		}

		beginBlockers = append(beginBlockers, beginBlocker)
	}

	bapp.beginBlockers = beginBlockers

	return bapp, nil
}
