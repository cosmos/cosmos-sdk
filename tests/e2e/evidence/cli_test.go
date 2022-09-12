//go:build e2e
// +build e2e

package evidence

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"cosmossdk.io/simapp"
	"github.com/cosmos/cosmos-sdk/testutil/network"
)

func TestIntegrationTestSuite(t *testing.T) {
	cfg := network.DefaultConfig(simapp.NewTestNetworkFixture)
	cfg.NumValidators = 1
	suite.Run(t, NewIntegrationTestSuite(cfg))
}
