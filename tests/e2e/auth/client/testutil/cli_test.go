//go:build e2e
// +build e2e

package testutil

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/testutil/network"
	testutil2 "github.com/cosmos/cosmos-sdk/x/auth/client/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth/testutil"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestIntegrationTestSuite(t *testing.T) {
	cfg, err := network.DefaultConfigWithAppConfig(testutil.AppConfig)
	require.NoError(t, err)
	cfg.NumValidators = 2
	suite.Run(t, testutil2.NewIntegrationTestSuite(cfg))
}
