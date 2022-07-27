//go:build norace
// +build norace

package testutil

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/testutil/network"
	clienttestutil "github.com/cosmos/cosmos-sdk/x/slashing/client/testutil"
	"github.com/cosmos/cosmos-sdk/x/slashing/testutil"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestIntegrationTestSuite(t *testing.T) {
	cfg, err := network.DefaultConfigWithAppConfig(testutil.AppConfig)
	require.NoError(t, err)
	cfg.NumValidators = 1
	suite.Run(t, clienttestutil.NewIntegrationTestSuite(cfg))
}
