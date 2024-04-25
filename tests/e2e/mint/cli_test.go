//go:build e2e
// +build e2e

package mint

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"cosmossdk.io/simapp"
	"cosmossdk.io/simapp/network"
)

func TestE2ETestSuite(t *testing.T) {
	cfg := network.DefaultConfig(simapp.NewTestNetworkFixture)
	cfg.NumValidators = 1
	suite.Run(t, NewE2ETestSuite(cfg))
}
