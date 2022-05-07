package testutil

import (
	_ "embed"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	dbm "github.com/tendermint/tm-db"

	"cosmossdk.io/core/appconfig"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/baseapp/runtime"
	"github.com/cosmos/cosmos-sdk/container"
	pruningtypes "github.com/cosmos/cosmos-sdk/pruning/types"
	"github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/testutil/network"
)

//go:embed app.yaml
var appConfig []byte

func TestIntegrationTestSuite(t *testing.T) {
	err := container.RunDebug(func(creator *runtime.AppCreator) {
		cfg := network.DefaultConfig()
		cfg.AppConstructor = func(val network.Validator) types.Application {
			app := creator.Create(val.Ctx.Logger, dbm.NewMemDB(),
				nil,
				baseapp.SetPruning(pruningtypes.NewPruningOptionsFromString(val.AppConfig.Pruning)),
				baseapp.SetMinGasPrices(val.AppConfig.MinGasPrices),
			)
			require.NoError(t, creator.Finish(true))
			return app
		}
		cfg.NumValidators = 1
		suite.Run(t, NewIntegrationTestSuite(cfg))
	},
		container.Debug(),
		appconfig.LoadYAML(appConfig))
	require.NoError(t, err)
}
