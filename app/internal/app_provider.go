package internal

import (
	"fmt"

	codecTypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/container"

	"github.com/cosmos/cosmos-sdk/app"
	"github.com/cosmos/cosmos-sdk/app/query"
	genutilprovider "github.com/cosmos/cosmos-sdk/x/genutil/provider"
)

func AppConfigProvider(config *app.Config) container.Option {
	moduleConfigMap := map[string]*codecTypes.Any{}

	if config.Abci.TxHandler == nil {
		return container.Error(fmt.Errorf("missing tx handler"))
	}
	moduleConfigMap["tx"] = config.Abci.TxHandler

	for _, modConfig := range config.Modules {
		moduleConfigMap[modConfig.Name] = modConfig.Config
	}

	return container.Options(
		app.ProvideModules(moduleConfigMap),
		// TODO should these be here:
		container.Provide(genutilprovider.Provider),
		query.Module,
	)
}
