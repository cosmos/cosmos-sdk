package testutil

import (
	_ "embed"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/testutil/network"
	"github.com/cosmos/cosmos-sdk/x/nft/testutil"
)

func TestIntegrationTestSuite(t *testing.T) {
	cfg, err := network.DefaultConfigWithAppConfig(testutil.AppConfig)
	require.NoError(t, err)
	cfg.NumValidators = 1
	suite.Run(t, NewIntegrationTestSuite(cfg))
}
