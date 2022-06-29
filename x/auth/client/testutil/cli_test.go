//go:build norace
// +build norace

package testutil

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/testutil/network"
	"github.com/cosmos/cosmos-sdk/x/auth/apptestutils"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestIntegrationTestSuite(t *testing.T) {
	cfg, err := network.DefaultConfigWithAppConfig(apptestutils.AppConfig)
	require.NoError(t, err)
	cfg.NumValidators = 2
	suite.Run(t, NewIntegrationTestSuite(cfg))
}
