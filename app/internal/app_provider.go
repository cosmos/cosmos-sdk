package internal

import (
	"github.com/cosmos/cosmos-sdk/app"
	"github.com/cosmos/cosmos-sdk/container"
)

type AppProvider struct {
	*ModuleContainer
	config          *app.Config
	moduleConfigMap map[string]*app.ModuleConfig
}

const txModuleScope container.Scope = "tx"

func NewAppProvider(config *app.Config) (*AppProvider, error) {
	ctr := NewModuleContainer()
	moduleConfigMap := map[string]*app.ModuleConfig{}

	err := ctr.AddModule(txModuleScope, config.Abci.TxHandler)
	if err != nil {
		return nil, err
	}

	for _, modConfig := range config.Modules {
		err = ctr.AddModule(container.Scope(modConfig.Name), modConfig.Config)
		if err != nil {
			return nil, err
		}
		moduleConfigMap[modConfig.Name] = modConfig
	}

	return &AppProvider{
		config:          config,
		container:       ctr,
		moduleConfigMap: moduleConfigMap,
	}, nil
}
