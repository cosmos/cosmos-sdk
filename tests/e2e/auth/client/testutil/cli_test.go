//go:build e2e
// +build e2e

package testutil

import (
	"testing"

	"cosmossdk.io/simapp"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	clientestutil "github.com/cosmos/cosmos-sdk/x/auth/client/testutil"

	"github.com/stretchr/testify/suite"
)

func TestIntegrationTestSuite(t *testing.T) {
	cfg := network.DefaultConfig(simapp.NewTestNetworkFixture)
	cfg.NumValidators = 2
	suite.Run(t, clientestutil.NewIntegrationTestSuite(cfg))
}
