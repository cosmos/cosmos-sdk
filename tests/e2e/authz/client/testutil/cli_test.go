//go:build e2e
// +build e2e

package testutil

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"cosmossdk.io/simapp"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	clienttestutil "github.com/cosmos/cosmos-sdk/x/authz/client/testutil"
)

func TestIntegrationTestSuite(t *testing.T) {
	cfg := network.DefaultConfig(simapp.NewTestNetworkFixture)
	cfg.NumValidators = 1
	suite.Run(t, clienttestutil.NewIntegrationTestSuite(cfg))
}
