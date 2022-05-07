package testutil

import (
	_ "embed"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/core/appconfig"

	"github.com/cosmos/cosmos-sdk/testutil/network"
)

//go:embed app.yaml
var appConfig []byte

func TestIntegrationTestSuite(t *testing.T) {
	cfg, err := network.DefaultConfigWithAppConfig(appconfig.LoadYAML(appConfig))
	require.NoError(t, err)
	cfg.NumValidators = 1
	suite.Run(t, NewIntegrationTestSuite(cfg))
}
