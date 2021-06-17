package internal

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/app"
	"github.com/cosmos/cosmos-sdk/app/query"
	"github.com/cosmos/cosmos-sdk/x/genutil/provider"
)

type AppProvider struct {
	*app.ModuleContainer
	config          *app.Config
	moduleConfigMap map[string]*app.ModuleConfig
}

func NewAppProvider(config *app.Config) (*AppProvider, error) {
	ctr := app.NewModuleContainer()
	moduleConfigMap := map[string]*app.ModuleConfig{}

	if config.Abci.TxHandler == nil {
		return nil, fmt.Errorf("missing tx handler")
	}
	err := ctr.AddProtoModule("tx", config.Abci.TxHandler)
	if err != nil {
		return nil, err
	}

	for _, modConfig := range config.Modules {
		err = ctr.AddProtoModule(modConfig.Name, modConfig.Config)
		if err != nil {
			return nil, err
		}
		moduleConfigMap[modConfig.Name] = modConfig
	}

	err = ctr.Provide(provider.Provider)
	if err != nil {
		return nil, err
	}

	err = ctr.Provide(query.Provider)
	if err != nil {
		return nil, err
	}

	return &AppProvider{
		ModuleContainer: ctr,
		config:          config,
		moduleConfigMap: moduleConfigMap,
	}, nil
}
