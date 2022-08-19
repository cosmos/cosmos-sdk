//go:build e2e
// +build e2e

package testutil

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/testutil/network"

	"github.com/stretchr/testify/suite"
)

func TestEndToEndTestSuite(t *testing.T) {
	cfg := network.DefaultConfig(simapp.NewTestNetworkFixture)
	cfg.NumValidators = 1
	suite.Run(t, NewEndToEndTestSuite(cfg))
}
