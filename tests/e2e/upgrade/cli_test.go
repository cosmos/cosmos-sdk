//go:build e2e
// +build e2e

package upgrade

import (
	"testing"

	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"cosmossdk.io/simapp"
	"github.com/cosmos/cosmos-sdk/testutil/network"
)

func TestIntegrationTestSuite(t *testing.T) {
	cfg := network.DefaultConfig(simapp.NewTestNetworkFixture)
	cfg.NumValidators = 1

	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	app.UpgradeKeeper.SetVersionSetter(app.BaseApp)
	app.UpgradeKeeper.SetModuleVersionMap(ctx, app.ModuleManager.GetVersionMap())

	suite.Run(t, NewIntegrationTestSuite(cfg, app.UpgradeKeeper, ctx))
}
