//go:build e2e
// +build e2e

package feegrant

import (
	"testing"

	"cosmossdk.io/simapp"
	"github.com/cosmos/cosmos-sdk/testutil/network"

	"github.com/stretchr/testify/suite"
)

func TestE2ETestSuite(t *testing.T) {
	cfg := network.DefaultConfig(simapp.NewTestNetworkFixture)
	cfg.NumValidators = 3
	suite.Run(t, NewE2ETestSuite(cfg))
}
