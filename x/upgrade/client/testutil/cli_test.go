//go:build norace
// +build norace

package testutil

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/depinject"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/cosmos/cosmos-sdk/x/upgrade/testutil"
)

func TestIntegrationTestSuite(t *testing.T) {
	appConfig := depinject.Configs(testutil.AppConfig, depinject.Supply(simtestutil.EmptyAppOptions{}))

	cfg, err := network.DefaultConfigWithAppConfig(appConfig)
	require.NoError(t, err)
	cfg.NumValidators = 1
	suite.Run(t, NewIntegrationTestSuite(cfg))
}
